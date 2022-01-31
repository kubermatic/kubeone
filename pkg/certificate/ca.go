/*
Copyright 2019 The KubeOne Authors.

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

package certificate

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/ssh"
	"k8c.io/kubeone/pkg/ssh/sshiofs"
	"k8c.io/kubeone/pkg/state"
)

const (
	KubernetesCACertPath = "/etc/kubernetes/pki/ca.crt"
	KubernetesCAKeyPath  = "/etc/kubernetes/pki/ca.key"
)

func kubernetesPKIFiles() []string {
	return []string{
		KubernetesCACertPath,
		KubernetesCAKeyPath,
		"/etc/kubernetes/pki/sa.key",
		"/etc/kubernetes/pki/sa.pub",
		"/etc/kubernetes/pki/front-proxy-ca.crt",
		"/etc/kubernetes/pki/front-proxy-ca.key",
		"/etc/kubernetes/pki/etcd/ca.crt",
		"/etc/kubernetes/pki/etcd/ca.key",
	}
}

func DownloadKubePKI(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
	sshfs := s.Runner.NewFS()

	for _, fname := range kubernetesPKIFiles() {
		buf, err := fs.ReadFile(sshfs, fname)
		if err != nil {
			return err
		}
		s.Configuration.KubernetesPKI[fname] = buf
	}

	if s.BackupFile != "" {
		s.Logger.Infoln("Creating local backup...")

		err := s.Configuration.Backup(s.BackupFile)
		if err != nil {
			// do not stop in case of failed backups, the user can
			// always create the backup themselves if needed
			s.Logger.Warnf("Failed to create backup: %v", err)
		}
	}

	return nil
}

func UploadKubePKI(s *state.State, _ *kubeoneapi.HostConfig, conn ssh.Connection) error {
	sshfs := s.Runner.NewFS()

	for _, fname := range kubernetesPKIFiles() {
		buf, found := s.Configuration.KubernetesPKI[fname]
		if !found {
			return fmt.Errorf("file %q found found in PKI", fname)
		}

		if err := sshfs.MkdirAll(path.Dir(fname), 0700); err != nil {
			return err
		}

		f, err := sshfs.Open(fname)
		if err != nil {
			return err
		}
		defer f.Close()
		fw, _ := f.(sshiofs.ExtendedFile)

		if err = fw.Truncate(0); err != nil {
			return err
		}

		if err = fw.Chmod(0600); err != nil {
			return err
		}

		if _, err = io.Copy(fw, bytes.NewBuffer(buf)); err != nil {
			return err
		}
	}

	return nil
}
