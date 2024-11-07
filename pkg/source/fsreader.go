package source

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/uservers/baggr/pkg/spec"
)

// FilesystemReader is a source reader that takes an io.FS filesystem as the
// source of data.
type FilesystemReader struct {
	FS fs.StatFS
}

// NewFilesystemReader creates a new filesystem reader
func NewFilesystemReader(myfs fs.StatFS) *FilesystemReader {
	return &FilesystemReader{
		FS: myfs,
	}
}

// ListDirFiles returns a list of al the files in a directory
func (fsr *FilesystemReader) ListDirFiles(ctx context.Context, path string) ([]*spec.File, error) {
	if fsr.FS == nil {
		return nil, fmt.Errorf("reader filesystem not set")
	}
	res := []*spec.File{}
	if err := fs.WalkDir(fsr.FS, path, func(subPath string, f fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		res = append(res, &spec.File{
			Source: filepath.Join(subPath),
			// Destination: "",
		})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking filesystem %w", err)
	}
	return res, nil
}

// OpenPath opens a path from the underlying FS. If fails it will return an error
// if the path cannot be openened because it is a directory it will return ErrIsDir
func (fsr *FilesystemReader) OpenPath(ctx context.Context, specFile *spec.File) (io.Reader, error) {
	if fsr.FS == nil {
		return nil, fmt.Errorf("reader filesystem not set")
	}
	info, err := fsr.FS.Stat(specFile.Source)
	if err != nil {
		return nil, fmt.Errorf("opening path from spec: %w", err)
	}

	// If it is a dir, just say so
	if info.IsDir() {
		return nil, ErrIsDir
	}

	f, err := fsr.FS.Open(specFile.Source)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	return f, nil

}
