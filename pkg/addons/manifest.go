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
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8c.io/kubeone/pkg/certificate/cabundle"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

func getManifests(s *state.State, templateData TemplateData) error {
	addonsPath := s.Cluster.Addons.Path
	if !filepath.IsAbs(addonsPath) && s.ManifestFilePath != "" {
		manifestAbsPath, err := filepath.Abs(filepath.Dir(s.ManifestFilePath))
		if err != nil {
			return errors.Wrap(err, "unable to get absolute path to the cluster manifest")
		}
		addonsPath = filepath.Join(manifestAbsPath, addonsPath)
	}

	overwriteRegistry := ""
	if s.Cluster.RegistryConfiguration != nil && s.Cluster.RegistryConfiguration.OverwriteRegistry != "" {
		overwriteRegistry = s.Cluster.RegistryConfiguration.OverwriteRegistry
	}

	manifests, err := loadAddonsManifests(addonsPath, s.Logger, s.Verbose, templateData, overwriteRegistry)
	if err != nil {
		return err
	}

	rawManifests, err := ensureAddonsLabelsOnResources(manifests)
	if err != nil {
		return err
	}

	combinedManifests := combineManifests(rawManifests)
	s.Configuration.AddFile("addons/addons.yaml", combinedManifests.String())

	return nil
}

// loadAddonsManifests loads all YAML files from a given directory and runs the templating logic
func loadAddonsManifests(addonsPath string, logger logrus.FieldLogger, verbose bool, templateData TemplateData, overwriteRegistry string) ([]runtime.RawExtension, error) {
	manifests := []runtime.RawExtension{}

	files, err := ioutil.ReadDir(addonsPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read the addons directory %s", addonsPath)
	}

	for _, file := range files {
		filePath := filepath.Join(addonsPath, file.Name())
		if file.IsDir() {
			logger.Infof("Found directory '%s' in the addons path. Ignoring.\n", file.Name())
			continue
		}
		ext := strings.ToLower(filepath.Ext(filePath))
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
			logger.Infof("Parsing addons manifest '%s'\n", file.Name())
		}

		manifestBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load addon %s", file.Name())
		}

		tpl, err := template.New("addons-base").Funcs(txtFuncMap(overwriteRegistry)).Parse(string(manifestBytes))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to template addons manifest %s", file.Name())
		}
		buf := bytes.NewBuffer([]byte{})
		if err := tpl.Execute(buf, templateData); err != nil {
			return nil, errors.Wrapf(err, "failed to template addons manifest %s", file.Name())
		}

		trim := strings.TrimSpace(buf.String())
		if len(trim) == 0 {
			logger.Infof("Addons manifest '%s' is empty after parsing. Skipping.\n", file.Name())
		}

		reader := kyaml.NewYAMLReader(bufio.NewReader(buf))
		for {
			b, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, errors.Wrapf(err, "failed reading from YAML reader for manifest %s", file.Name())
			}
			b = bytes.TrimSpace(b)
			if len(b) == 0 {
				continue
			}
			decoder := kyaml.NewYAMLToJSONDecoder(bytes.NewBuffer(b))
			raw := runtime.RawExtension{}
			if err := decoder.Decode(&raw); err != nil {
				return nil, errors.Wrapf(err, "failed to decode manifest %s", file.Name())
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

// ensureAddonsLabelsOnResources applies the addons label on all resources in the manifest
func ensureAddonsLabelsOnResources(manifests []runtime.RawExtension) ([]*bytes.Buffer, error) {
	var rawManifests []*bytes.Buffer

	for _, m := range manifests {
		parsedUnstructuredObj := &metav1unstructured.Unstructured{}
		if _, _, err := metav1unstructured.UnstructuredJSONScheme.Decode(m.Raw, nil, parsedUnstructuredObj); err != nil {
			return nil, errors.Wrapf(err, "failed to parse unstructured fields")
		}

		existingLabels := parsedUnstructuredObj.GetLabels()
		if existingLabels == nil {
			existingLabels = map[string]string{}
		}
		existingLabels[addonLabel] = ""
		parsedUnstructuredObj.SetLabels(existingLabels)

		jsonBuffer := &bytes.Buffer{}
		if err := metav1unstructured.UnstructuredJSONScheme.Encode(parsedUnstructuredObj, jsonBuffer); err != nil {
			return nil, errors.Wrap(err, "encoding json failed")
		}

		// Must be encoded back to YAML, otherwise kubectl fails to apply because it tries to parse the whole
		// thing as json
		yamlBytes, err := yaml.JSONToYAML(jsonBuffer.Bytes())
		if err != nil {
			return nil, err
		}

		rawManifests = append(rawManifests, bytes.NewBuffer(yamlBytes))
	}

	return rawManifests, nil
}

// combineManifests combines all manifest into a single one.
// This is needed so we can properly utilize kubectl apply --prune
func combineManifests(manifests []*bytes.Buffer) *bytes.Buffer {
	parts := make([]string, len(manifests))
	for i, m := range manifests {
		s := m.String()
		s = strings.TrimSuffix(s, "\n")
		s = strings.TrimSpace(s)
		parts[i] = s
	}

	return bytes.NewBufferString(strings.Join(parts, "\n---\n") + "\n")
}

func txtFuncMap(overwriteRegistry string) template.FuncMap {
	funcs := sprig.TxtFuncMap()

	funcs["Registry"] = func(registry string) string {
		if overwriteRegistry != "" {
			return overwriteRegistry
		}
		return registry
	}

	funcs["caBundleEnvVar"] = func() (string, error) {
		buf, err := yaml.Marshal([]corev1.EnvVar{cabundle.EnvVar()})
		return string(buf), err
	}

	funcs["caBundleVolume"] = func() (string, error) {
		buf, err := yaml.Marshal([]corev1.Volume{cabundle.Volume()})
		return string(buf), err
	}

	funcs["caBundleVolumeMount"] = func() (string, error) {
		buf, err := yaml.Marshal([]corev1.VolumeMount{cabundle.VolumeMount()})
		return string(buf), err
	}

	return funcs
}
