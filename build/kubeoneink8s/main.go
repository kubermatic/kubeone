package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"encoding/base64"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/kubermatic/kubeone/pkg/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}

	kubeconfigs, err := filepath.Glob("*-kubeconfig")
	if err != nil {
		panic(err) // Only error is invalid pattern
	}

	if len(kubeconfigs) != 1 {
		panic(fmt.Errorf("expected 1 kubeconfig, found %v (Glob *-kubeconfig)", len(kubeconfigs)))
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	secrets := dynamicClient.Resource(schema.GroupVersionResource{Version: "v1", Resource: "secrets"})
	f, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err)
	}
	namespaceName := strings.TrimSpace(string(f))

	d, e := ioutil.ReadFile(kubeconfigs[0])
	if e != nil {
		panic(e)
	}
	data := map[string]string{"kubeconfig": base64.StdEncoding.EncodeToString(d)}

	existing, err := secrets.Namespace(namespaceName).Get("kubeconfig", metav1.GetOptions{})
	switch {
	case errors.IsNotFound(err):
		secret := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{"name": "kubeconfig"},
				"data":     data,
			},
		}
		if _, err = secrets.Namespace(namespaceName).Create(secret, metav1.CreateOptions{}); err != nil {
			panic(err)
		}
	case err == nil:
		existing.Object["data"] = data
	default:
		panic(err)
	}
}
