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
	"io"
	"io/fs"
)

// ExtendedFile extends fs.File bringing it closer in abilities to the os.File.
type ExtendedFile interface {
	fs.File
	io.Writer
	io.Seeker

	// Chmod changes the mode of the file to mode. If there is an error, it will be of type *fs.PathError.
	Chmod(mode fs.FileMode) error

	// Chown changes the numeric uid and gid of the named file. If there is an error, it will be of type *fs.PathError.
	Chown(uid, gid int) error

	// Truncate changes the size of the file. It does not change the I/O offset. If there is an error, it will be of
	// type *fs.PathError.
	Truncate(size int64) error
}

// MkdirFS is the interface implemented by a file system that provides mkdir capabilities.
type MkdirFS interface {
	fs.FS

	// MkdirAll creates a directory named path, along with any necessary parents, and returns nil, or else returns an
	// error. The permission bits perm (before umask) are used for last directory that MkdirAll creates. If path is
	// already a directory, MkdirAll does nothing and returns nil.
	MkdirAll(path string, perm fs.FileMode) error
}
