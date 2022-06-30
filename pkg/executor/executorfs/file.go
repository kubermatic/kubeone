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
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/executor"
)

var (
	// validate that sshfile still implements those non-obligatory but very useful interfaces
	_ executor.ExtendedFile = &virtfile{}
	_ io.WriteSeeker        = &virtfile{}
)

type virtfile struct {
	conn   executor.Interface
	name   string
	cursor int64
	fi     fs.FileInfo
}

func (vf *virtfile) Stat() (fs.FileInfo, error) {
	if vf.fi != nil {
		return vf.fi, nil
	}

	fi, err := newSSHFileInfo(vf.name, vf.conn)
	vf.fi = fi

	return fi, err
}

func (vf *virtfile) Truncate(size int64) error {
	const cmdTpl = `sudo truncate --size=%d %q`
	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, size, vf.name)
	)

	_, err := vf.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return vf.pathError("truncate", stdout.String(), stderr.String(), err)
	}

	return nil
}

func (vf *virtfile) Chown(uid, gid int) error {
	const cmdTpl = `sudo chown %d:%d %q`
	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, uid, gid, vf.name)
	)

	_, err := vf.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return vf.pathError("chown", stdout.String(), stderr.String(), err)
	}

	return nil
}

func (vf *virtfile) Chmod(mode os.FileMode) error {
	const cmdTpl = `sudo chmod %o %q`
	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, mode, vf.name)
	)

	_, err := vf.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return vf.pathError("chmod", stdout.String(), stderr.String(), err)
	}

	return nil
}

func (vf *virtfile) Read(p []byte) (int, error) {
	const cmdTpl = `sudo dd status=none iflag=count_bytes,skip_bytes skip=%d count=%d if=%q`
	var (
		stdout, stderr bytes.Buffer
		cmd            = fmt.Sprintf(cmdTpl, vf.cursor, len(p), vf.name)
	)

	_, err := vf.conn.POpen(cmd, nil, &stdout, &stderr)
	if err != nil {
		return 0, vf.pathError("read", stdout.String(), stderr.String(), err)
	}

	n, err := stdout.Read(p)
	vf.cursor += int64(n)

	return n, err
}

func (vf *virtfile) Write(p []byte) (int, error) {
	const cmdTpl = `sudo dd status=none oflag=seek_bytes conv=notrunc seek=%d of=%q`
	var (
		stdout, stderr strings.Builder
		cmd            = fmt.Sprintf(cmdTpl, vf.cursor, vf.name)
		stdin          = bytes.NewBuffer(p)
	)

	_, err := vf.conn.POpen(cmd, stdin, &stdout, &stderr)
	if err != nil {
		return 0, vf.pathError("write", stdout.String(), stderr.String(), err)
	}

	n, err := stdout.Write(p)
	vf.cursor += int64(n)

	return n, err
}

func (vf *virtfile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		vf.cursor = offset
	case io.SeekCurrent:
		vf.cursor += offset
	case io.SeekEnd:
		return vf.cursor, fmt.Errorf("io.SeekEnd unimplemented")
	}

	return vf.cursor, nil
}

func (vf *virtfile) Close() error {
	vf.cursor = 0
	vf.fi = nil

	return nil
}

func (vf *virtfile) pathError(op, stdout, stderr string, err error) *fs.PathError {
	return &fs.PathError{
		Path: vf.name,
		Op:   op,
		Err:  errors.Wrapf(err, "%v %v", stderr, stdout),
	}
}
