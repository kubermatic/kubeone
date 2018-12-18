package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
)

type tarGzip struct {
	file *os.File
	gz   *gzip.Writer
	arch *tar.Writer
}

const (
	fileMode = 06000
)

// NewTarGzip returns a new tar.gz archive.
// TODO(GvW): Why not passing io.Writer here, so we could much more persistent backends
func NewTarGzip(filename string) (Archive, error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileMode)
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

// Add a file to the archive
func (tgz tarGzip) Add(file string, content []byte) error {
	if tgz.arch == nil {
		return errors.New("archive has already been closed")
	}

	hdr := &tar.Header{
		Name: file,
		Mode: fileMode,
		Size: int64(len(content)),
	}

	if err := tgz.arch.WriteHeader(hdr); err != nil {
		return fmt.Errorf("failed to write tar file header: %v", err)
	}

	if _, err := tgz.arch.Write(content); err != nil {
		return fmt.Errorf("failed to write tar file data: %v", err)
	}

	return nil
}

// Close the archive writer
func (tgz tarGzip) Close() {
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
