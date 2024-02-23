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

package executorfs

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/fail"
)

var _ executor.MkdirFS = &virtfs{}

func New(conn executor.Interface) executor.MkdirFS {
	return &virtfs{conn: conn}
}

type virtfs struct {
	conn executor.Interface
}

func (vfs *virtfs) Open(name string) (fs.File, error) {
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

	return &virtfile{
		conn: vfs.conn,
		name: name,
	}, nil
}

func (vfs *virtfs) Glob(pattern string) ([]string, error) {
	const cmdTpl = `sudo bash -c 'list=(%s); echo ${list[@]}'`

	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, pattern)
	)

	_, err := vfs.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return nil, fail.SSH(&fs.PathError{
			Op:   "glob",
			Path: pattern,
			Err:  fmt.Errorf("glob failed: %w, %s %s", err, stdout.String(), stderr.String()),
		}, "glob")
	}

	return strings.Split(stdout.String(), " "), nil
}

func (vfs *virtfs) MkdirAll(path string, perm fs.FileMode) error {
	const cmdTpl = `sudo mkdir --mode=%o --parents %q`

	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, perm, path)
	)

	_, err := vfs.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return fail.SSH(&fs.PathError{
			Op:   "mkdir",
			Path: path,
			Err:  fmt.Errorf("%w %s %s", err, stdout.String(), stderr.String()),
		}, "mkdirall")
	}

	return nil
}

func (vfs *virtfs) ReadFile(name string) ([]byte, error) {
	var buf bytes.Buffer
	var stderr strings.Builder

	_, err := vfs.conn.POpen(fmt.Sprintf("sudo cat %q", name), nil, &buf, &stderr)
	if err != nil {
		return nil, fail.SSH(&fs.PathError{
			Op:   "read",
			Path: name,
			Err:  fmt.Errorf("%w %s", err, stderr.String()),
		}, "read file")
	}

	return buf.Bytes(), nil
}

func newSSHFileInfo(name string, conn executor.Interface) (fs.FileInfo, error) {
	const cmdTpl = "sudo stat --printf='%%s %%f %%Y' %q"

	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, name)
	)

	exitCode, err := conn.POpen(cmd, nil, &stdout, &stderr)
	if exitCode != 0 || err != nil {
		return nil, fail.SSH(&fs.PathError{
			Op:   "stat",
			Path: name,
			Err:  err,
		}, "stat file")
	}

	var (
		size    int64
		mode    fs.FileMode
		modTime int64
	)

	fia := strings.Split(stdout.String(), " ")
	if len(fia) != 3 {
		return nil, fail.Runtime(fs.ErrInvalid, "wrong number of stat output")
	}

	if _, err = fmt.Sscanf(fia[0], "%d", &size); err != nil {
		return nil, fail.Runtime(&fs.PathError{
			Err:  err,
			Path: name,
			Op:   "stat",
		}, "scanning file size")
	}

	if _, err = fmt.Sscanf(fia[1], "%x", &mode); err != nil {
		return nil, fail.Runtime(&fs.PathError{
			Err:  err,
			Path: name,
			Op:   "stat",
		}, "scanning file mode")
	}

	if _, err = fmt.Sscanf(fia[2], "%d", &modTime); err != nil {
		return nil, fail.Runtime(&fs.PathError{
			Err:  err,
			Path: name,
			Op:   "stat",
		}, "scanning file modtime")
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
