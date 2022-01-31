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

package sshiofs

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"k8c.io/kubeone/pkg/ssh"
)

func New(conn ssh.Connection) MkdirFS {
	return &sshfs{conn: conn}
}

type sshfs struct {
	conn ssh.Connection
}

func (sfs *sshfs) Open(name string) (fs.File, error) {
	var hadSlashPrefix bool
	if strings.HasPrefix(name, "/") {
		name = strings.TrimPrefix(name, "/")
		hadSlashPrefix = true
	}
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}
	if hadSlashPrefix {
		name = "/" + name
	}

	return &sshfile{
		conn: sfs.conn,
		name: name,
	}, nil
}

func (sfs *sshfs) Glob(pattern string) ([]string, error) {
	const cmdTpl = `sudo bash -c 'list=(%s); echo ${list[@]}'`

	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, pattern)
	)

	_, err := sfs.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return nil, fmt.Errorf("glob failed: %w, %s %s", err, stdout.String(), stderr.String())
	}

	return strings.Split(stdout.String(), " "), nil
}

func (sfs *sshfs) MkdirAll(path string, perm fs.FileMode) error {
	const cmdTpl = `sudo mkdir --mode=%o --parents %q`

	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, perm, path)
	)

	_, err := sfs.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return &fs.PathError{
			Op:   "mkdir",
			Path: path,
			Err:  fmt.Errorf("%w %s %s", err, stdout.String(), stderr.String()),
		}
	}

	return nil
}

func (sfs *sshfs) ReadFile(name string) ([]byte, error) {
	var buf bytes.Buffer
	_, err := sfs.conn.POpen(fmt.Sprintf("sudo cat %q", name), nil, &buf, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// TODO: implement
// func (sfs *sshfs) ReadDir(name string) ([]fs.DirEntry, error) {
// 	return nil, nil
// }

func newSSHFileInfo(name string, conn ssh.Connection) (fs.FileInfo, error) {
	const cmdTpl = "sudo stat --printf='%%s %%f %%Y' %q"

	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, name)
	)

	exitCode, err := conn.POpen(cmd, nil, &stdout, &stderr)
	if exitCode != 0 || err != nil {
		return nil, &fs.PathError{
			Op:   "stat",
			Path: name,
			Err:  fmt.Errorf("%s %s %w", stderr.String(), err.Error(), fs.ErrNotExist),
		}
	}

	var (
		size    int64
		mode    fs.FileMode
		modTime int64
	)

	fia := strings.Split(stdout.String(), " ")
	if len(fia) != 3 {
		return nil, fs.ErrInvalid
	}

	if _, err = fmt.Sscanf(fia[0], "%d", &size); err != nil {
		return nil, err
	}

	if _, err = fmt.Sscanf(fia[1], "%x", &mode); err != nil {
		return nil, err
	}

	if _, err = fmt.Sscanf(fia[2], "%d", &modTime); err != nil {
		return nil, err
	}

	return &fileInfo{
		name: name,
		size: size,
		mode: mode,
		time: time.Unix(modTime, 0),
	}, nil
}

type fileInfo struct {
	name string
	size int64
	mode fs.FileMode
	time time.Time
}

func (fi *fileInfo) Size() int64 { return fi.size }

func (fi *fileInfo) Mode() fs.FileMode { return fi.mode }

func (fi *fileInfo) ModTime() time.Time { return fi.time }

func (fi *fileInfo) Name() string { return fi.name }

func (fi *fileInfo) IsDir() bool { return fi.mode.IsDir() }

func (*fileInfo) Sys() interface{} { return nil }
