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
	"io"
	"io/fs"
	"os"
	"strings"

	"k8c.io/kubeone/pkg/ssh"
)

var (
	// validate that sshfile still implements those non-obligatory but very useful interfaces
	_ ExtendedFile   = &sshfile{}
	_ io.WriteSeeker = &sshfile{}
)

type sshfile struct {
	conn   ssh.Connection
	name   string
	cursor int64
	fi     fs.FileInfo
}

func (sf *sshfile) Stat() (fs.FileInfo, error) {
	if sf.fi != nil {
		return sf.fi, nil
	}

	fi, err := newSSHFileInfo(sf.name, sf.conn)
	sf.fi = fi
	return fi, err
}

func (sf *sshfile) Truncate(size int64) error {
	const cmdTpl = `sudo truncate --size=%d %q`
	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, size, sf.name)
	)

	_, err := sf.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return sf.pathError("truncate", stdout.String(), stderr.String(), err)
	}

	return nil
}

func (sf *sshfile) Chown(uid, gid int) error {
	const cmdTpl = `sudo chown %d:%d %q`
	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, uid, gid, sf.name)
	)

	_, err := sf.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return sf.pathError("chown", stdout.String(), stderr.String(), err)
	}

	return nil
}

func (sf *sshfile) Chmod(mode os.FileMode) error {
	const cmdTpl = `sudo chmod %o %q`
	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, mode, sf.name)
	)

	_, err := sf.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return sf.pathError("chmod", stdout.String(), stderr.String(), err)
	}

	return nil
}

func (sf *sshfile) Read(p []byte) (int, error) {
	const cmdTpl = `sudo dd status=none iflag=count_bytes,skip_bytes skip=%d count=%d if=%q`
	var (
		stdout, stderr bytes.Buffer
		cmd            = fmt.Sprintf(cmdTpl, sf.cursor, len(p), sf.name)
	)

	_, err := sf.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return 0, sf.pathError("read", stdout.String(), stderr.String(), err)
	}

	n, err := stdout.Read(p)
	sf.cursor += int64(n)
	return n, err
}

func (sf *sshfile) Write(p []byte) (int, error) {
	const cmdTpl = `sudo dd status=none oflag=seek_bytes conv=notrunc seek=%d of=%q`
	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, sf.cursor, sf.name)
		stdin          = bytes.NewBuffer(p)
	)

	_, err := sf.conn.POpen(cmd, stdin, &stdout, &stderr)
	if err != nil {
		return 0, sf.pathError("write", stdout.String(), stderr.String(), err)
	}

	n, err := stdout.Write(p)
	sf.cursor += int64(n)
	return n, err
}

func (sf *sshfile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		sf.cursor = offset
	case io.SeekCurrent:
		sf.cursor += offset
	case io.SeekEnd:
		return sf.cursor, fmt.Errorf("io.SeekEnd unimplemented")
	}

	return sf.cursor, nil
}

func (sf *sshfile) Close() error {
	sf.cursor = 0
	sf.fi = nil
	return nil
}

func (sf *sshfile) pathError(op, stdout, stderr string, err error) *fs.PathError {
	return &fs.PathError{
		Path: sf.name,
		Op:   op,
		Err:  fmt.Errorf("%w: %v %v", err, stderr, stdout),
	}
}
