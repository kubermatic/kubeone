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

package configupload

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/archive"
	"k8c.io/kubeone/pkg/executor"
	"k8c.io/kubeone/pkg/executor/executorfs"
	"k8c.io/kubeone/pkg/fail"
)

// Configuration holds a map of generated files
type Configuration struct {
	files         map[string]string
	KubernetesPKI map[string][]byte
}

// NewConfiguration constructor
func NewConfiguration() *Configuration {
	return &Configuration{
		files:         make(map[string]string),
		KubernetesPKI: make(map[string][]byte),
	}
}

// AddFile save file contents for future references
func (c *Configuration) AddFile(filename string, content string) {
	c.files[filename] = strings.TrimSpace(content) + "\n"
}

// AddFilePath saves file contents from a file on filesystem for future references
func (c *Configuration) AddFilePath(filename, filePath, manifestFilePath string) error {
	// Normalize the file path. In the case when the relative path is provided,
	// the path is relative to the KubeOne configuration file.
	if !filepath.IsAbs(filePath) && manifestFilePath != "" {
		manifestAbsPath, err := filepath.Abs(filepath.Dir(manifestFilePath))
		if err != nil {
			return fail.Runtime(err, "getting absolut path to the manifest file")
		}
		filePath = filepath.Join(manifestAbsPath, filePath)
	}

	b, err := os.ReadFile(filePath)
	if err != nil {
		return fail.Runtime(err, "reading file")
	}

	c.AddFile(filename, string(b))

	return nil
}

// UploadTo directory all the files
func (c *Configuration) UploadTo(conn executor.Interface, directory string) error {
	virtfs := executorfs.New(conn)

	for filename, content := range c.files {
		target := filepath.Join(directory, filename)

		// ensure the base dir exists
		dir := filepath.Dir(target)
		if err := virtfs.MkdirAll(dir, 0700); err != nil {
			return err
		}

		f, err := virtfs.Open(target)
		if err != nil {
			return err
		}
		defer f.Close()

		file, _ := f.(executor.ExtendedFile)
		if err = file.Truncate(0); err != nil {
			return err
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}

		_, err = io.Copy(file, strings.NewReader(content))
		if err != nil {
			return fail.Runtime(err, "copying to %s file", target)
		}

		if err := file.Chmod(0600); err != nil {
			return err
		}
	}

	return nil
}

// Backup dumps the files into a .tar.gz archive.
func (c *Configuration) Backup(target string) error {
	archive, err := archive.NewTarGzip(target)
	if err != nil {
		return err
	}
	defer archive.Close()

	for filename, content := range c.files {
		err = archive.Add(filename, content)
		if err != nil {
			return err
		}
	}

	for filename, content := range c.KubernetesPKI {
		err = archive.Add(strings.TrimPrefix(filename, "/"), string(content))
		if err != nil {
			return err
		}
	}

	return nil
}

// Get returns contents of the generated file by filename
func (c *Configuration) Get(filename string) (string, error) {
	content, ok := c.files[filename]
	if !ok {
		return "", fail.RuntimeError{
			Op: "getting file from internal configuration store",
			Err: &os.PathError{
				Op:   "read",
				Path: filename,
				Err:  errors.New("no such file or directory"),
			},
		}
	}

	return content, nil
}
