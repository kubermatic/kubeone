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
	"bytes"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"text/template"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"
)

var testManifests = []string{
	`kind: ConfigMap
apiVersion: v1
metadata:
  name: test1
  namespace: kube-system
  labels:
    app: test
data:
  foo: bar
`,

	`kind: ConfigMap
apiVersion: v1
metadata:
  name: test2
  namespace: kube-system
  labels:
    app: test
data:
  foo: bar`,

	`kind: ConfigMap
apiVersion: v1
metadata:
  name: test3
  namespace: kube-system
  labels:
    app: test
data:
  foo: bar
`}

const (
	testManifest1WithoutLabel = `apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  labels:
    app: test
    cluster: {{ .Config.Name }}
  name: test1
  namespace: kube-system
`

	testManifest1WithEmptyLabel = `apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  labels:
    app: test
    cluster: kubeone-test
    kubeone.io/addon: ""
  name: test1
  namespace: kube-system
`

	testManifest1WithNamedLabel = `apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  labels:
    app: test
    cluster: kubeone-test
    kubeone.io/addon: test-addon
  name: test1
  namespace: kube-system
`
)

const (
	testManifest1WithImage = `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: test
    cluster: kubeone-test
    kubeone.io/addon: ""
  name: test1
  namespace: kube-system
spec:
  containers:
  - image: {{ Registry "k8s.gcr.io" }}/kube-apiserver:v1.19.3
`

	testManifest1WithImageParsed = `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: test
    cluster: kubeone-test
    kubeone.io/addon: ""
  name: test1
  namespace: kube-system
spec:
  containers:
  - image: k8s.gcr.io/kube-apiserver:v1.19.3
`

	testManifest1WithCustomImageParsed = `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: test
    cluster: kubeone-test
    kubeone.io/addon: ""
  name: test1
  namespace: kube-system
spec:
  containers:
  - image: 127.0.0.1:5000/kube-apiserver:v1.19.3
`
)

var (
	// testManifest1 & testManifest3 have a linebreak at the end, testManifest2 not
	combinedTestManifest = fmt.Sprintf("%s---\n%s\n---\n%s", testManifests[0], testManifests[1], testManifests[2])
)

func TestEnsureAddonsLabelsOnResources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		addonName        string
		addonManifest    string
		expectedManifest string
	}{
		{
			name:             "addon with no name (root directory addons)",
			addonName:        "",
			addonManifest:    testManifest1WithoutLabel,
			expectedManifest: testManifest1WithEmptyLabel,
		},
		{
			name:             "addon with name",
			addonName:        "test-addon",
			addonManifest:    testManifest1WithoutLabel,
			expectedManifest: testManifest1WithNamedLabel,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			addonsDir := t.TempDir()

			if writeErr := os.WriteFile(path.Join(addonsDir, "testManifest.yaml"), []byte(tc.addonManifest), 0600); writeErr != nil {
				t.Fatalf("unable to create temporary addon manifest: %v", writeErr)
			}

			td := templateData{
				Config: &kubeoneapi.KubeOneCluster{
					Name: "kubeone-test",
				},
			}

			applier := &applier{
				TemplateData: td,
				LocalFS:      os.DirFS(addonsDir),
			}

			manifests, err := applier.loadAddonsManifests(applier.LocalFS, ".", nil, nil, false, "")
			if err != nil {
				t.Fatalf("unable to load manifests: %v", err)
			}
			if len(manifests) != 1 {
				t.Fatalf("expected to load 1 manifest, got %d", len(manifests))
			}

			b, err := ensureAddonsLabelsOnResources(manifests, tc.addonName)
			if err != nil {
				t.Fatalf("unable to ensure labels: %v", err)
			}
			manifest := b[0].String()

			if manifest != tc.expectedManifest {
				t.Fatalf("invalid manifest returned. expected \n%s, got \n%s", tc.expectedManifest, manifest)
			}
		})
	}
}

func TestCombineManifests(t *testing.T) {
	var manifests []*bytes.Buffer
	for _, m := range testManifests {
		manifests = append(manifests, bytes.NewBufferString(m))
	}

	manifest := combineManifests(manifests)

	if manifest.String() != combinedTestManifest {
		t.Fatalf("invalid combined manifest returned. expected \n%s, got \n%s", combinedTestManifest, manifest.String())
	}
}

func TestImageRegistryParsing(t *testing.T) {
	testCases := []struct {
		name                  string
		registryConfiguration *kubeoneapi.RegistryConfiguration
		inputManifest         string
		expectedManifest      string
	}{
		{
			name:                  "default registry configuration",
			registryConfiguration: nil,
			inputManifest:         testManifest1WithImage,
			expectedManifest:      testManifest1WithImageParsed,
		},
		{
			name: "custom registry",
			registryConfiguration: &kubeoneapi.RegistryConfiguration{
				OverwriteRegistry: "127.0.0.1:5000",
			},
			inputManifest:    testManifest1WithImage,
			expectedManifest: testManifest1WithCustomImageParsed,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			addonsDir := t.TempDir()

			if writeErr := os.WriteFile(path.Join(addonsDir, "testManifest.yaml"), []byte(tc.inputManifest), 0600); writeErr != nil {
				t.Fatalf("unable to create temporary addon manifest: %v", writeErr)
			}

			td := templateData{
				Config: &kubeoneapi.KubeOneCluster{
					Name:                  "kubeone-test",
					RegistryConfiguration: tc.registryConfiguration,
				},
			}

			overwriteRegistry := ""
			if tc.registryConfiguration != nil && tc.registryConfiguration.OverwriteRegistry != "" {
				overwriteRegistry = tc.registryConfiguration.OverwriteRegistry
			}

			applier := &applier{
				TemplateData: td,
				LocalFS:      os.DirFS(addonsDir),
			}

			manifests, err := applier.loadAddonsManifests(applier.LocalFS, ".", nil, nil, false, overwriteRegistry)
			if err != nil {
				t.Fatalf("unable to load manifests: %v", err)
			}
			if len(manifests) != 1 {
				t.Fatalf("expected to load 1 manifest, got %d", len(manifests))
			}

			b, err := ensureAddonsLabelsOnResources(manifests, "")
			if err != nil {
				t.Fatalf("unable to ensure labels: %v", err)
			}
			manifest := b[0].String()

			if manifest != tc.expectedManifest {
				t.Fatalf("invalid manifest returned. expected \n%s, got \n%s", tc.expectedManifest, manifest)
			}
		})
	}
}

func TestCABundleFuncs(t *testing.T) {
	tests := []string{
		"caBundleEnvVar",
		"caBundleVolume",
		"caBundleVolumeMount",
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt, func(t *testing.T) {
			tpl, err := template.New("addons-base").Funcs(txtFuncMap("")).Parse(fmt.Sprintf(`{{ %s }}`, tt))

			if err != nil {
				t.Errorf("failed to parse template: %v", err)
				t.FailNow()
			}

			var out strings.Builder
			if err := tpl.Execute(&out, nil); err != nil {
				t.Errorf("failed to parse template: %v", err)
			}
			t.Logf("\n%s", out.String())
		})
	}
}

func mustMarshal(obj runtime.Object) []byte {
	buf, err := yaml.Marshal(obj)
	if err != nil {
		panic(err)
	}

	buf, err = yaml.YAMLToJSON(buf)
	if err != nil {
		panic(err)
	}

	return buf
}

func Test_addSecretCSIVolume(t *testing.T) {
	testPodSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "container1",
			},
		},
	}

	testPodTemplateSpec := corev1.PodTemplateSpec{
		Spec: testPodSpec,
	}

	secretProviderClassName := "something"

	testdataPodSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name: "container1",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "secrets-store",
						MountPath: "/mnt/secrets-store",
						ReadOnly:  true,
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "secrets-store",
				VolumeSource: corev1.VolumeSource{
					CSI: &corev1.CSIVolumeSource{
						Driver:   "secrets-store.csi.k8s.io",
						ReadOnly: pointer.Bool(true),
						VolumeAttributes: map[string]string{
							"secretProviderClass": secretProviderClassName,
						},
					},
				},
			},
		},
	}

	testdata := map[string]runtime.Object{
		"deployment": &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				APIVersion: appsv1.SchemeGroupVersion.String(),
				Kind:       "Deployment",
			},
			Spec: appsv1.DeploymentSpec{
				Template: testPodTemplateSpec,
			},
		},
		"statefulset": &appsv1.StatefulSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: appsv1.SchemeGroupVersion.String(),
				Kind:       "StatefulSet",
			},
			Spec: appsv1.StatefulSetSpec{
				Template: testPodTemplateSpec,
			},
		},
		"daemonset": &appsv1.DaemonSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: appsv1.SchemeGroupVersion.String(),
				Kind:       "DaemonSet",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: testPodTemplateSpec,
			},
		},
		"replicaset": &appsv1.ReplicaSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: appsv1.SchemeGroupVersion.String(),
				Kind:       "ReplicaSet",
			},
			Spec: appsv1.ReplicaSetSpec{
				Template: testPodTemplateSpec,
			},
		},
		"pod": &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "Pod",
			},
			Spec: testPodSpec,
		},
		"job": &batchv1.Job{
			TypeMeta: metav1.TypeMeta{
				APIVersion: batchv1.SchemeGroupVersion.String(),
				Kind:       "Job",
			},
			Spec: batchv1.JobSpec{
				Template: testPodTemplateSpec,
			},
		},
		"cronjob": &batchv1.CronJob{
			TypeMeta: metav1.TypeMeta{
				APIVersion: batchv1.SchemeGroupVersion.String(),
				Kind:       "CronJob",
			},
			Spec: batchv1.CronJobSpec{
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: testPodTemplateSpec,
					},
				},
			},
		},
	}

	type args struct {
		docs                    []runtime.RawExtension
		secretProviderClassName string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "deployment",
			args: args{
				docs: []runtime.RawExtension{
					{
						Raw: mustMarshal(testdata["deployment"]),
					},
				},
				secretProviderClassName: secretProviderClassName,
			},
		},
		{
			name: "statefulset",
			args: args{
				docs: []runtime.RawExtension{
					{
						Raw: mustMarshal(testdata["statefulset"]),
					},
				},
				secretProviderClassName: secretProviderClassName,
			},
		},
		{
			name: "daemonset",
			args: args{
				docs: []runtime.RawExtension{
					{
						Raw: mustMarshal(testdata["daemonset"]),
					},
				},
				secretProviderClassName: secretProviderClassName,
			},
		},
		{
			name: "replicaset",
			args: args{
				docs: []runtime.RawExtension{
					{
						Raw: mustMarshal(testdata["replicaset"]),
					},
				},
				secretProviderClassName: secretProviderClassName,
			},
		},
		{
			name: "pod",
			args: args{
				docs: []runtime.RawExtension{
					{
						Raw: mustMarshal(testdata["pod"]),
					},
				},
				secretProviderClassName: secretProviderClassName,
			},
		},
		{
			name: "job",
			args: args{
				docs: []runtime.RawExtension{
					{
						Raw: mustMarshal(testdata["job"]),
					},
				},
				secretProviderClassName: secretProviderClassName,
			},
		},
		{
			name: "cronjob",
			args: args{
				docs: []runtime.RawExtension{
					{
						Raw: mustMarshal(testdata["cronjob"]),
					},
				},
				secretProviderClassName: secretProviderClassName,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := addSecretCSIVolume(tt.args.docs, tt.args.secretProviderClassName)
			if (err != nil) != tt.wantErr {
				t.Errorf("addSecretCSIVolume() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			switch testdata[tt.name].(type) {
			case *appsv1.Deployment:
				var kbject appsv1.Deployment
				_ = yaml.Unmarshal(tt.args.docs[0].Raw, &kbject)
				if !reflect.DeepEqual(kbject.Spec.Template.Spec, testdataPodSpec) {
					t.Errorf("missing volume/volumeMount")
				}
			case *appsv1.StatefulSet:
				var kbject appsv1.StatefulSet
				_ = yaml.Unmarshal(tt.args.docs[0].Raw, &kbject)
				if !reflect.DeepEqual(kbject.Spec.Template.Spec, testdataPodSpec) {
					t.Errorf("missing volume/volumeMount")
				}
			case *appsv1.DaemonSet:
				var kbject appsv1.DaemonSet
				_ = yaml.Unmarshal(tt.args.docs[0].Raw, &kbject)
				if !reflect.DeepEqual(kbject.Spec.Template.Spec, testdataPodSpec) {
					t.Errorf("missing volume/volumeMount")
				}
			case *appsv1.ReplicaSet:
				var kbject appsv1.ReplicaSet
				_ = yaml.Unmarshal(tt.args.docs[0].Raw, &kbject)
				if !reflect.DeepEqual(kbject.Spec.Template.Spec, testdataPodSpec) {
					t.Errorf("missing volume/volumeMount")
				}
			case *corev1.Pod:
				var kbject corev1.Pod
				_ = yaml.Unmarshal(tt.args.docs[0].Raw, &kbject)
				if !reflect.DeepEqual(kbject.Spec, testdataPodSpec) {
					t.Errorf("missing volume/volumeMount")
				}
			case *batchv1.Job:
				var kbject batchv1.Job
				_ = yaml.Unmarshal(tt.args.docs[0].Raw, &kbject)
				if !reflect.DeepEqual(kbject.Spec.Template.Spec, testdataPodSpec) {
					t.Errorf("missing volume/volumeMount")
				}
			case *batchv1.CronJob:
				var kbject batchv1.CronJob
				_ = yaml.Unmarshal(tt.args.docs[0].Raw, &kbject)
				if !reflect.DeepEqual(kbject.Spec.JobTemplate.Spec.Template.Spec, testdataPodSpec) {
					t.Errorf("missing volume/volumeMount")
				}
			}
		})
	}
}
