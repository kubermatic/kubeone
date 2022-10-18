/*
Copyright 2022 The KubeOne Authors.

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

package examples

import (
	"embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed terraform/*
var terraformFS embed.FS

func CopyTo(dst, src string) error {
	return fs.WalkDir(terraformFS, src, func(fspath string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dstPartPath := strings.Split(fspath, src)[1]
		if dstPartPath == "" {
			// skip it
			return nil
		}

		dstFullPath := filepath.Join(dst, dstPartPath)
		if de.IsDir() {
			return os.MkdirAll(dstFullPath, 0755)
		}

		fh, err := os.Create(dstFullPath)
		if err != nil {
			return err
		}
		defer fh.Close()

		sourceFileName := filepath.Join(src, de.Name())
		srcFh, err := terraformFS.Open(sourceFileName)
		if err != nil {
			return err
		}

		_, err = io.Copy(fh, srcFh)

		return err
	})
}
