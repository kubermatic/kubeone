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

package archive

import (
	"archive/tar"
	"compress/gzip"
	"os"

	"github.com/pkg/errors"
)

type tarGzip struct {
	file *os.File
	gz   *gzip.Writer
	arch *tar.Writer
}

// NewTarGzip returns a new tar.gz archive.
func NewTarGzip(filename string) (Archive, error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	gz := gzip.NewWriter(f)
	arch := tar.NewWriter(gz)

	return &tarGzip{
		file: f,
		gz:   gz,
		arch: arch,
	}, nil
}

func (tgz tarGzip) Add(file string, content string) error {
	if tgz.arch == nil {
		return errors.New("archive has already been closed")
	}

	hdr := &tar.Header{
		Name: file,
		Mode: 0600,
		Size: int64(len(content)),
	}

	if err := tgz.arch.WriteHeader(hdr); err != nil {
		return errors.Wrap(err, "failed to write tar file header")
	}

	if _, err := tgz.arch.Write([]byte(content)); err != nil {
		return errors.Wrap(err, "failed to write tar file data")
	}

	return nil
}

func (tgz *tarGzip) Close() {
	if tgz.arch != nil {
		tgz.arch.Close()
		tgz.arch = nil
	}

	if tgz.gz != nil {
		tgz.gz.Close()
		tgz.gz = nil
	}

	if tgz.file != nil {
		tgz.file.Close()
		tgz.file = nil
	}
}
