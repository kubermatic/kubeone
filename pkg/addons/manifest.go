/*
Copyright 2020 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package addons

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/certificate/cabundle"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
	"k8c.io/kubeone/pkg/templates/resources"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	kyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

const (
	ParamsEnvPrefix = "env:"
)

func (a *applier) getManifestsFromDirectory(st *state.State, fsys fs.FS, addonName string) ([]*metav1unstructured.Unstructured, error) {
	var (
		addonParams       map[string]string
		disableTemplating = false
	)
	if st.Cluster.Addons.Enabled() {
		for _, addon := range st.Cluster.Addons.DeclaredAddonsOnly() {
			if addon.Name == addonName {
				addonParams = addon.Params
				disableTemplating = addon.DisableTemplating

				break
			}
		}
	}

	if addonName == resources.AddonUnattendedUpgrades {
		var err error

		addonParams, err = withFlatcarNodesPresentParam(st, addonParams)
		if err != nil {
			return nil, err
		}
	}

	manifests, err := a.loadAddonsManifests(fsys, addonName, addonParams, st.Logger, st.Verbose, st.Cluster, disableTemplating)
	if err != nil {
		return nil, err
	}

	addonsToMutate := sets.NewString(
		append(
			resources.CloudAddons(),
			resources.AddonBackupsRestic,
			resources.AddonOperatingSystemManager,
			resources.AddonMachineController,
		)...,
	)

	if !disableTemplating && addonsToMutate.Has(addonName) {
		if st.Cluster.CertificateAuthority.Bundle != "" {
			if err = mutatePodTemplateSpec(manifests, func(podTpl *corev1.PodTemplateSpec) {
				cabundle.Inject(st.Cluster.CertificateAuthority.Bundle, podTpl)
			}); err != nil {
				return nil, err
			}
		}

		if st.Cluster.CloudProvider.SecretProviderClassName != "" {
			if err = mutatePodTemplateSpec(manifests, func(podSpec *corev1.PodTemplateSpec) {
				volume := corev1.Volume{
					Name: "secrets-store",
					VolumeSource: corev1.VolumeSource{
						CSI: &corev1.CSIVolumeSource{
							Driver:   "secrets-store.csi.k8s.io",
							ReadOnly: new(true),
							VolumeAttributes: map[string]string{
								"secretProviderClass": st.Cluster.CloudProvider.SecretProviderClassName,
							},
						},
					},
				}

				volumeMount := corev1.VolumeMount{
					Name:      "secrets-store",
					MountPath: "/mnt/secrets-store",
					ReadOnly:  true,
				}

				podSpec.Spec.Volumes = append(podSpec.Spec.Volumes, volume)
				for i := range podSpec.Spec.Containers {
					podSpec.Spec.Containers[i].VolumeMounts = append(podSpec.Spec.Containers[i].VolumeMounts, volumeMount)
				}
			}); err != nil {
				return nil, err
			}
		}
	}

	return ensureAddonsLabelsOnResources(manifests, addonName, disableTemplating)
}

// loadAddonsManifests loads all YAML files from a given directory and runs the templating logic
func (a *applier) loadAddonsManifests(
	fsys fs.FS,
	addonName string,
	addonParams map[string]string,
	logger logrus.FieldLogger,
	verbose bool,
	k1cluster *kubeoneapi.KubeOneCluster,
	disableTemplating bool,
) ([]runtime.RawExtension, error) {
	var manifests []runtime.RawExtension

	files, err := fs.ReadDir(fsys, filepath.Join(".", addonName))
	if err != nil {
		return nil, fail.Runtime(err, "reading addons directory")
	}

	for _, file := range files {
		filePath := filepath.Join(addonName, file.Name())
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		// Only YAML, YML and JSON manifests are supported
		switch ext {
		case ".yaml", ".yml", ".json":
		default:
			if verbose {
				logger.Infof("Skipping file %q because it's not .yaml/.yml/.json file\n", file.Name())
			}

			continue
		}

		if verbose {
			logger.Debugf("Parsing addons manifest '%s'\n", file.Name())
		}

		manifestBytes, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			return nil, fail.Runtime(err, "reading addon")
		}

		// We need to escape occurrences of Go text template in the manifests.
		if disableTemplating {
			res := strings.ReplaceAll(string(manifestBytes), "{{", "^{{")
			res = strings.ReplaceAll(res, "}}", "}}^")
			manifestBytes = []byte(res)
		}
		manifest := bytes.NewBuffer(manifestBytes)

		if !disableTemplating {
			overwriteRegistry := k1cluster.RegistryConfiguration.ImageRegistry("")

			tpl, err := template.New("addons-base").
				Funcs(txtFuncMap(overwriteRegistry)).
				Funcs(template.FuncMap{
					"CABundle": func() string {
						return k1cluster.CertificateAuthority.Bundle
					},
				}).
				Parse(string(manifestBytes))
			if err != nil {
				return nil, fail.Runtime(err, "parsing addons manifest template %q", file.Name())
			}

			// Make a copy and merge Params
			tplDataParams := map[string]string{}
			maps.Copy(tplDataParams, a.TemplateData.Params)

			if credAddonParams := a.AddonParams[addonName]; len(credAddonParams) > 0 {
				for k, v := range credAddonParams {
					if _, exists := addonParams[k]; exists {
						return nil, fail.RuntimeError{
							Op: "merging addon params",
							Err: fmt.Errorf(
								"param %q for addon %q is set both in the KubeOneCluster manifest and in "+
									"the credentials file's addonParams section - remove it from one of them",
								k, addonName,
							),
						}
					}

					tplDataParams[k] = v
				}
			}

			maps.Copy(tplDataParams, addonParams)

			defaultAddonParams(k1cluster, addonName, tplDataParams)

			// Resolve environment variables in Params
			for k, v := range tplDataParams {
				if envName, ok := strings.CutPrefix(v, ParamsEnvPrefix); ok {
					if env, ok := os.LookupEnv(envName); ok {
						tplDataParams[k] = env
					} else {
						return nil, fail.RuntimeError{
							Op:  "resolving template data environment variables",
							Err: fmt.Errorf("%q not found", envName),
						}
					}
				}
			}

			tplData := a.TemplateData
			tplData.Params = tplDataParams

			manifest = bytes.NewBuffer([]byte{})
			if err := tpl.Execute(manifest, tplData); err != nil {
				return nil, fail.Runtime(err, "executing addons manifest template %q", file.Name())
			}

			if len(bytes.TrimSpace(manifest.Bytes())) == 0 {
				logger.Infof("Addons manifest %q is empty after parsing. Skipping.\n", file.Name())
			}
		}

		reader := kyaml.NewYAMLReader(bufio.NewReader(manifest))
		for {
			yamlDoc, err := reader.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, fail.Runtime(err, "reading YAML reader for manifest %q", file.Name())
			}

			yamlDoc = bytes.TrimSpace(yamlDoc)
			if len(yamlDoc) == 0 {
				continue
			}

			decoder := kyaml.NewYAMLToJSONDecoder(bytes.NewBuffer(yamlDoc))
			raw := runtime.RawExtension{}
			if err := decoder.Decode(&raw); err != nil {
				return nil, fail.Runtime(err, "unmarshalling manifest %q", file.Name())
			}

			if len(raw.Raw) == 0 {
				// This can happen if the manifest contains only comments
				continue
			}

			manifests = append(manifests, raw)
		}
	}

	return manifests, nil
}

// withFlatcarNodesPresentParam returns a copy of the addon params with the
// flatcarNodesPresent param set, unless the user already set it.
func withFlatcarNodesPresentParam(st *state.State, addonParams map[string]string) (map[string]string, error) {
	if _, ok := addonParams[paramFlatcarNodesPresent]; ok {
		return addonParams, nil
	}

	present, err := flatcarNodesPresent(st.Context, st.Cluster, st.DynamicClient)
	if err != nil {
		return nil, err
	}

	params := maps.Clone(addonParams)
	if params == nil {
		params = map[string]string{}
	}

	params[paramFlatcarNodesPresent] = strconv.FormatBool(present)

	return params, nil
}

func defaultAddonParams(k1cluster *kubeoneapi.KubeOneCluster, addonName string, tplDataParams map[string]string) {
	if addonName == resources.AddonCNICanal {
		if k1cluster.CloudProvider.Hetzner != nil {
			// Customize cni-canal addon on hetzner
			defaultIfaceParam(tplDataParams, "en|eth1")
		}
	}
}

func defaultIfaceParam(tplDataParams map[string]string, reg string) {
	_, iface := tplDataParams["IFACE"]
	_, ifaceregex := tplDataParams["IFACE_REGEX"]
	if !iface && !ifaceregex {
		tplDataParams["IFACE_REGEX"] = reg
	}
}

// ensureAddonsLabelsOnResources decodes the manifests, flattening List objects, and applies the addons label on
// all resulting objects. With templating disabled, the escaping of template delimiters done when loading the
// manifests is reverted.
func ensureAddonsLabelsOnResources(manifests []runtime.RawExtension, addonName string, disableTemplating bool) ([]*metav1unstructured.Unstructured, error) {
	var objs []*metav1unstructured.Unstructured

	for _, rawManifest := range manifests {
		parsedUnstructuredObj := &metav1unstructured.Unstructured{}
		if _, _, err := metav1unstructured.UnstructuredJSONScheme.Decode(rawManifest.Raw, nil, parsedUnstructuredObj); err != nil {
			return nil, fail.Runtime(err, "parsing unstructured fields")
		}

		items := []*metav1unstructured.Unstructured{parsedUnstructuredObj}

		if parsedUnstructuredObj.IsList() {
			list, err := parsedUnstructuredObj.ToList()
			if err != nil {
				return nil, fail.Runtime(err, "parsing list items")
			}

			items = items[:0]
			for i := range list.Items {
				items = append(items, &list.Items[i])
			}
		}

		for _, obj := range items {
			if disableTemplating {
				obj.Object, _ = unescapeTemplateDelimiters(obj.Object).(map[string]any)
			}

			existingLabels := obj.GetLabels()
			if existingLabels == nil {
				existingLabels = map[string]string{}
			}
			existingLabels[addonLabel] = addonName
			obj.SetLabels(existingLabels)

			objs = append(objs, obj)
		}
	}

	return objs, nil
}

// templateDelimitersUnescaper reverts the "^{{" and "}}^" escaping that loadAddonsManifests applies when
// templating is disabled, so YAML to JSON conversion doesn't choke on unquoted template delimiters.
var templateDelimitersUnescaper = strings.NewReplacer("^{{", "{{", "}}^", "}}")

func unescapeTemplateDelimiters(v any) any {
	switch v := v.(type) {
	case string:
		return templateDelimitersUnescaper.Replace(v)
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, val := range v {
			out[templateDelimitersUnescaper.Replace(key)] = unescapeTemplateDelimiters(val)
		}

		return out
	case []any:
		for i := range v {
			v[i] = unescapeTemplateDelimiters(v[i])
		}

		return v
	default:
		return v
	}
}

type vsphereCSIWebhookConfig struct {
	Port     string `toml:"port"`
	CertFile string `toml:"cert-file"`
	KeyFile  string `toml:"key-file"`
}

type vsphereCSIWebhookConfigWrapper struct {
	WebHookConfig vsphereCSIWebhookConfig `toml:"WebHookConfig"`
}

func txtFuncMap(overwriteRegistry string) template.FuncMap {
	funcs := sprig.TxtFuncMap()

	funcs["Registry"] = func(registry string) string {
		if overwriteRegistry != "" {
			return overwriteRegistry
		}

		return registry
	}

	funcs["required"] = requiredTemplateFunc
	funcs["caBundleEnvVar"] = caBundleEnvVarTemplateFunc
	funcs["caBundleVolume"] = caBundleVolumeTemplateFunc
	funcs["caBundleVolumeMount"] = caBundleVolumeMountTemplateFunc
	funcs["EquinixMetalSecret"] = equinixMetalSecretTemplateFunc
	funcs["vSphereCSIWebhookConfig"] = vSphereCSIWebhookConfigTemplateFunc

	return funcs
}

func requiredTemplateFunc(warn string, input any) (any, error) {
	switch val := input.(type) {
	case nil:
		return val, errors.New(warn)
	case string:
		if val == "" {
			return val, errors.New(warn)
		}
	}

	return input, nil
}

func caBundleEnvVarTemplateFunc() (string, error) {
	buf, err := yaml.Marshal([]corev1.EnvVar{cabundle.EnvVar()})

	return string(buf), err
}

func caBundleVolumeTemplateFunc() (string, error) {
	buf, err := yaml.Marshal([]corev1.Volume{cabundle.Volume()})

	return string(buf), err
}

func caBundleVolumeMountTemplateFunc() (string, error) {
	buf, err := yaml.Marshal([]corev1.VolumeMount{cabundle.VolumeMount()})

	return string(buf), err
}

func equinixMetalSecretTemplateFunc(apiKey, projectID string) (string, error) {
	equinixMetalSecret := struct {
		APIKey    string `json:"apiKey"`
		ProjectID string `json:"projectID"`
	}{
		APIKey:    apiKey,
		ProjectID: projectID,
	}

	buf, err := json.Marshal(equinixMetalSecret)

	return string(buf), err
}

func vSphereCSIWebhookConfigTemplateFunc() (string, error) {
	cfg := vsphereCSIWebhookConfigWrapper{
		WebHookConfig: vsphereCSIWebhookConfig{
			Port:     "8443",
			CertFile: "/run/secrets/tls/cert.pem",
			KeyFile:  "/run/secrets/tls/key.pem",
		},
	}

	var buf strings.Builder
	enc := toml.NewEncoder(&buf)
	enc.Indent = ""
	err := enc.Encode(cfg)

	return buf.String(), err
}

func mutatePodTemplateSpec(docs []runtime.RawExtension, mutatorFn func(podTpl *corev1.PodTemplateSpec)) error {
	for i := range docs {
		ubject := metav1unstructured.Unstructured{}
		_, _, err := metav1unstructured.UnstructuredJSONScheme.Decode(docs[i].Raw, nil, &ubject)
		if err != nil {
			return err
		}

		switch ubject.GroupVersionKind().GroupKind() {
		case appsv1.SchemeGroupVersion.WithKind("Deployment").GroupKind():
			var obj appsv1.Deployment
			err = repackObject(&obj, &docs[i], func() {
				mutatorFn(&obj.Spec.Template)
			})
		case appsv1.SchemeGroupVersion.WithKind("StatefulSet").GroupKind():
			var obj appsv1.StatefulSet
			err = repackObject(&obj, &docs[i], func() {
				mutatorFn(&obj.Spec.Template)
			})
		case appsv1.SchemeGroupVersion.WithKind("DaemonSet").GroupKind():
			var obj appsv1.DaemonSet
			err = repackObject(&obj, &docs[i], func() {
				mutatorFn(&obj.Spec.Template)
			})
		case appsv1.SchemeGroupVersion.WithKind("ReplicaSet").GroupKind():
			var obj appsv1.ReplicaSet
			err = repackObject(&obj, &docs[i], func() {
				mutatorFn(&obj.Spec.Template)
			})
		case batchv1.SchemeGroupVersion.WithKind("Job").GroupKind():
			var obj batchv1.Job
			err = repackObject(&obj, &docs[i], func() {
				mutatorFn(&obj.Spec.Template)
			})
		case batchv1.SchemeGroupVersion.WithKind("CronJob").GroupKind():
			var obj batchv1.CronJob
			err = repackObject(&obj, &docs[i], func() {
				mutatorFn(&obj.Spec.JobTemplate.Spec.Template)
			})
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func repackObject(kubeobject runtime.Object, obj *runtime.RawExtension, mutator func()) error {
	if err := yaml.Unmarshal(obj.Raw, kubeobject); err != nil {
		return err
	}

	mutator()

	buf, err := yaml.Marshal(kubeobject)
	if err != nil {
		return err
	}

	js, err := yaml.YAMLToJSON(buf)
	if err != nil {
		return err
	}

	obj.Raw = js

	return nil
}
