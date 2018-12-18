package util

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kubermatic/kubeone/pkg/archive"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

// Configuration holds a map of generated files
type Configuration struct {
	files map[string][]byte
}

// NewConfiguration constructor
func NewConfiguration() *Configuration {
	return &Configuration{
		files: make(map[string][]byte),
	}
}

// AddFile save file contents for future references
func (c *Configuration) AddFile(filename string, content []byte) {
	c.files[filename] = content
}

// UploadTo directory all the files
func (c *Configuration) UploadTo(conn ssh.Connection, directory string) error {
	const (
		fileMode = 0644
	)
	for filename, content := range c.files {
		size := int64(len(content))
		target := filepath.Join(directory, filename)

		// ensure the base dir exists
		dir := filepath.Dir(target)
		_, _, _, err := conn.Exec(fmt.Sprintf(`mkdir -p -- "%s"`, dir))
		if err != nil {
			return fmt.Errorf("failed to create ./%s directory: %v", dir, err)
		}

		err = conn.Upload(bytes.NewReader(content), size, fileMode, target)
		if err != nil {
			return fmt.Errorf("failed to upload file %s: %v", filename, err)
		}
	}

	return nil
}

// Download a files matching `source` pattern
func (c *Configuration) Download(conn ssh.Connection, source string, prefix string) error {
	// list files
	stdout, stderr, _, err := conn.Exec(fmt.Sprintf(`cd -- "%s" && find * -type f`, source))
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr)
	}

	filenames := strings.Split(stdout, "\n")
	for _, filename := range filenames {
		fullsource := source + "/" + filename

		localfile := filename
		if len(prefix) > 0 {
			localfile = prefix + "/" + localfile
		}

		var buf bytes.Buffer
		err := conn.Download(fullsource, &buf)
		if err != nil {
			return err
		}

		c.files[localfile] = buf.Bytes()
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
		return fmt.Errorf("failed to open archive: %v", err)
	}
	defer archive.Close()

	for filename, content := range c.files {
		err = archive.Add(filename, content)
		if err != nil {
			return fmt.Errorf("failed to add %s to archive: %v", filename, err)
		}
	}

	return nil
}

// Get returns contents of the generated file by filename
func (c *Configuration) Get(filename string) ([]byte, error) {
	content, ok := c.files[filename]
	if !ok {
		return []byte{}, fmt.Errorf("could not find file %s", filename)
	}

	return content, nil
}
