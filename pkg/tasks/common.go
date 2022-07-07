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
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/yaml"
)

const (
	kubeadmEnvFlagsFile   = "/var/lib/kubelet/kubeadm-flags.env"
	kubeletKubeadmArgsEnv = "KUBELET_KUBEADM_ARGS"
	kubeletConfigFile     = "/var/lib/kubelet/config.yaml"
)

func updateRemoteFile(s *state.State, filePath string, modifier func(content []byte) ([]byte, error)) error {
	sshfs := s.Runner.NewFS()
	f, err := sshfs.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return fail.Runtime(err, "reading %q file", filePath)
	}

	buf, err = modifier(buf)
	if err != nil {
		return fail.Runtime(err, "updating remote file %q", filePath)
	}

	fw, ok := f.(executor.ExtendedFile)
	if !ok {
		return fail.RuntimeError{
			Op:  "checking if file satisfy sshiofs.ExtendedFile interface",
			Err: errors.New("file is not writable"),
		}
	}

	if err = fw.Truncate(0); err != nil {
		return err
	}

	if _, err = fw.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if _, err = io.Copy(fw, bytes.NewBuffer(buf)); err != nil {
		return fail.Runtime(err, "copying data into %q", filePath)
	}

	return nil
}

func unmarshalKubeletFlags(buf []byte) (map[string]string, error) {
	// throw away KUBELET_KUBEADM_ARGS=
	s1 := strings.SplitN(string(buf), "=", 2)
	if len(s1) != 2 {
		return nil, fail.RuntimeError{
			Op:  "splitting kubelet kubeadm args",
			Err: errors.New("can't parse: wrong split length"),
		}
	}

	envValue := strings.Trim(s1[1], "\n")
	envValue = strings.Trim(envValue, `"`)
	flagsvalues := strings.Split(envValue, " ")
	kubeletflagsMap := map[string]string{}

	for _, flg := range flagsvalues {
		fl := strings.SplitN(flg, "=", 2)
		if len(fl) != 2 {
			return nil, fail.RuntimeError{
				Op:  "splitting kubelet args",
				Err: errors.New("wrong split length"),
			}
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

	return []byte(fmt.Sprintf(`%s="%s"`, kubeletKubeadmArgsEnv, strings.Join(kvpairs, " ")))
}

func unmarshalKubeletConfig(configBytes []byte) (*kubeletconfigv1beta1.KubeletConfiguration, error) {
	var config kubeletconfigv1beta1.KubeletConfiguration
	err := yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, fail.Runtime(err, "unmarshalling %T", config)
	}

	return &config, nil
}

func marshalKubeletConfig(config *kubeletconfigv1beta1.KubeletConfiguration) ([]byte, error) {
	encodedCfg, err := yaml.Marshal(config)
	if err != nil {
		return nil, fail.Runtime(err, "marshalling %T", config)
	}

	return encodedCfg, nil
}
