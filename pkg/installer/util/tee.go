package util

import (
	"bytes"
	"io"
	"strings"
)

// Tee mimics the unix `tee` command by piping its
// input through to the upstream writer and also
// capturing it in a buffer.
type Tee struct {
	buffer   bytes.Buffer
	upstream io.WriteCloser
}

func (t *Tee) Write(p []byte) (int, error) {
	t.buffer.Write(p)

	return t.upstream.Write(p)
}

func (t *Tee) String() string {
	return strings.TrimSpace(t.buffer.String())
}

func (t *Tee) Close() error {
	return t.upstream.Close()
}
