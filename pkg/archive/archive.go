package archive

// Archive represents a file-backed collection of files.
type Archive interface {
	Add(file string, content string) error
	Close()
}
