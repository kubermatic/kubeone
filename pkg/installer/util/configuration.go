package util

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kubermatic/kubeone/pkg/archive"
	"github.com/kubermatic/kubeone/pkg/ssh"
)

type Configuration struct {
	files map[string]string
}

func NewConfiguration() *Configuration {
	return &Configuration{
		files: make(map[string]string),
	}
}

func (c *Configuration) AddFile(filename string, content string) {
	c.files[filename] = strings.TrimSpace(content) + "\n"
}

func (c *Configuration) UploadTo(conn ssh.Connection, directory string) error {
	for filename, content := range c.files {
		size := int64(len(content))
		target := filepath.Join(directory, filename)

		// ensure the base dir exists
		dir := filepath.Dir(target)
		_, _, _, err := conn.Exec(fmt.Sprintf(`mkdir -p -- "%s"`, dir))
		if err != nil {
			return fmt.Errorf("failed to create ./%s directory: %v", dir, err)
		}

		err = conn.Upload(strings.NewReader(content), size, 0644, target)
		if err != nil {
			return fmt.Errorf("failed to upload file %s: %v", filename, err)
		}
	}

	return nil
}

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

		c.files[localfile] = buf.String()
	}

	return nil
}

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
