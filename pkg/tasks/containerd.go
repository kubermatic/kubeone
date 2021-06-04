/*
Copyright 2021 The KubeOne Authors.

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

package tasks

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/scripts"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/ssh/sshiofs"
	"k8c.io/kubeone/pkg/state"

	corev1 "k8s.io/api/core/v1"
)

const (
	kubeadmCRISocket      = "kubeadm.alpha.kubernetes.io/cri-socket"
	kubeadmEnvFlagsFile   = "/var/lib/kubelet/kubeadm-flags.env"
	kubeletKubeadmArgsEnv = "KUBELET_KUBEADM_ARGS"
)

var (
	containerdKubeletFlags = map[string]string{
		"--container-runtime":          "remote",
		"--container-runtime-endpoint": "unix:///run/containerd/containerd.sock",
	}
)

func validateContainerdInConfig(s *state.State) error {
	if s.Cluster.ContainerRuntime.Containerd == nil {
		return errors.New("containerd must be enabled in config")
	}

	return nil
}

func patchCRISocketAnnotation(s *state.State) error {
	var nodes corev1.NodeList

	if err := s.DynamicClient.List(s.Context, &nodes); err != nil {
		return err
	}

	for _, node := range nodes.Items {
		if socketPath, found := node.Annotations[kubeadmCRISocket]; found {
			if socketPath != "/var/run/dockershim.sock" {
				continue
			}

			if node.Annotations == nil {
				node.Annotations = map[string]string{}
			}
			node.Annotations[kubeadmCRISocket] = "unix:///run/containerd/containerd.sock"

			if err := s.DynamicClient.Update(s.Context, &node); err != nil {
				return err
			}
		}
	}

	return nil
}

func migrateToContainerd(s *state.State) error {
	return s.RunTaskOnAllNodes(migrateToContainerdTask, state.RunSequentially)
}

func migrateToContainerdTask(s *state.State, node *kubeone.HostConfig, conn ssh.Connection) error {
	s.Logger.Info("Migrating container runtime to containerd")

	sshfs := s.Runner.NewFS()
	f, err := sshfs.Open(kubeadmEnvFlagsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	kubeletFlags, err := unmarshalKubeletFlags(buf)
	if err != nil {
		return err
	}

	for k, v := range containerdKubeletFlags {
		kubeletFlags[k] = v
	}

	buf = marshalKubeletFlags(kubeletFlags)
	fw, ok := f.(sshiofs.ExtendedFile)
	if !ok {
		return errors.New("file is not writable")
	}

	err = fw.Truncate(0)
	if err != nil {
		return err
	}

	_, err = fw.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = io.Copy(fw, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	migrateScript, err := scripts.MigrateToContainerd(s.Cluster.RegistryConfiguration.InsecureRegistryAddress())
	if err != nil {
		return err
	}

	_, _, err = s.Runner.RunRaw(migrateScript)
	if err != nil {
		return err
	}

	// TODO(kron4eg): replace with better waiting polling
	sleepTime := 30 * time.Second
	s.Logger.Infof("Waiting %s to ensure main control plane components are up...", sleepTime)
	time.Sleep(sleepTime)

	return nil
}

func unmarshalKubeletFlags(buf []byte) (map[string]string, error) {
	// throw away KUBELET_KUBEADM_ARGS=
	s1 := strings.SplitN(string(buf), "=", 2)
	if len(s1) != 2 {
		return nil, errors.New("can't parse: wrong split lenght")
	}

	envValue := strings.Trim(s1[1], `"`)
	flagsvalues := strings.Split(envValue, " ")
	kubeletflagsMap := map[string]string{}

	for _, flg := range flagsvalues {
		fl := strings.Split(flg, "=")
		if len(fl) != 2 {
			return nil, errors.New("wrong split lenght")
		}
		kubeletflagsMap[fl[0]] = fl[1]
	}

	return kubeletflagsMap, nil
}

func marshalKubeletFlags(kubeletflags map[string]string) []byte {
	kvpairs := []string{}
	for k, v := range kubeletflags {
		kvpairs = append(kvpairs, fmt.Sprintf("%s=%s", k, v))
	}

	sort.Strings(kvpairs)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, `%s="`, kubeletKubeadmArgsEnv)

	for i, val := range kvpairs {
		format := "%s "
		if i == len(kvpairs)-1 {
			format = "%s"
		}
		fmt.Fprintf(&buf, format, val)
	}

	buf.WriteString(`"`)

	return buf.Bytes()
}
