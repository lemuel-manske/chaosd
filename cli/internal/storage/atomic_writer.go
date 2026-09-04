package storage

import (
	"io/fs"
	"os"
	"path/filepath"
)

// AtomicWriter defines an interface for writing data to a file atomically, thus,
// ensuring that the file is either fully written or not written at all, preventing partial writes.
type AtomicWriter interface {
	Write(filePath string, data []byte, perm fs.FileMode) error
}

type AtomicFileWriter struct{}

func NewAtomicFileWriter() AtomicWriter {
	return &AtomicFileWriter{}
}

func (w *AtomicFileWriter) Write(filePath string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(filePath)

	tmp, err := os.CreateTemp(dir, ".chaosd-*")
	if err != nil {
		return err
	}

	tmpName := tmp.Name()

	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	if err := tmp.Chmod(perm); err != nil {
		return err
	}

	if _, err := tmp.Write(data); err != nil {
		return err
	}

	if err := tmp.Sync(); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, filePath)
}
