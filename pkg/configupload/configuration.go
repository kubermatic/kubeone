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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/kubermatic/kubeone/pkg/archive"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// Configuration holds a map of generated files
type Configuration struct {
	files map[string]string
}

// NewConfiguration constructor
func NewConfiguration() *Configuration {
	return &Configuration{
		files: make(map[string]string),
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
			return errors.Wrap(err, "unable to get absolute path to the cluster manifest")
		}
		filePath = filepath.Join(manifestAbsPath, filePath)
	}

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "unable to open given file")
	}

	c.AddFile(filename, string(b))
	return nil
}

// UploadTo directory all the files
func (c *Configuration) UploadTo(conn ssh.Connection, directory string) error {
	for filename, content := range c.files {
		target := filepath.Join(directory, filename)

		// ensure the base dir exists
		dir := filepath.Dir(target)
		_, _, _, err := conn.Exec(fmt.Sprintf(`mkdir -p -- "%s"`, dir))
		if err != nil {
			return errors.Wrapf(err, "failed to create ./%s directory", dir)
		}

		w, err := conn.File(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
		if err != nil {
			return errors.Wrapf(err, "failed to open remote file for write: %s", filename)
		}
		defer w.Close()

		_, err = io.Copy(w, strings.NewReader(content))
		if err != nil {
			return errors.Wrapf(err, "failed to write remote file %s", filename)
		}

		if wchmod, ok := w.(interface{ Chmod(os.FileMode) error }); ok {
			if err := wchmod.Chmod(0644); err != nil {
				return errors.Wrapf(err, "failed to chmod %s", filename)
			}
		}
	}

	return nil
}

// Download a files matching `source` pattern
func (c *Configuration) Download(conn ssh.Connection, source string, prefix string) error {
	// list files
	stdout, stderr, _, err := conn.Exec(fmt.Sprintf(`cd -- "%s" && find * -type f`, source))
	if err != nil {
		return errors.Wrapf(err, "%s", stderr)
	}

	filenames := strings.Split(stdout, "\n")
	for _, filename := range filenames {
		fullsource := source + "/" + filename

		localfile := filename
		if len(prefix) > 0 {
			localfile = prefix + "/" + localfile
		}

		var buf bytes.Buffer
		r, err := conn.File(fullsource, os.O_RDONLY)
		if err != nil {
			return errors.Wrapf(err, "failed to open remote file for read: %s", fullsource)
		}

		_, err = io.Copy(&buf, r)
		if err != nil {
			return errors.Wrapf(err, "failed to read remote file: %s", fullsource)
		}

		c.files[localfile] = buf.String()
	}

	return nil
}

// Debug list filenames and their size to the standard output
func (c *Configuration) Debug() {
	for filename, content := range c.files {
		fmt.Printf("%s: %d bytes\n", filename, len(content))
	}
}

// Backup dumps the files into a .tar.gz archive.
func (c *Configuration) Backup(target string) error {
	archive, err := archive.NewTarGzip(target)
	if err != nil {
		return errors.Wrap(err, "failed to open archive")
	}
	defer archive.Close()

	for filename, content := range c.files {
		err = archive.Add(filename, content)
		if err != nil {
			return errors.Wrapf(err, "failed to add %s to archive", filename)
		}
	}

	return nil
}

// Get returns contents of the generated file by filename
func (c *Configuration) Get(filename string) (string, error) {
	content, ok := c.files[filename]
	if !ok {
		return "", errors.Errorf("could not find file %s", filename)
	}

	return content, nil
}
