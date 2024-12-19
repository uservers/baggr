// Package source
package source

import (
	"context"
	"errors"
	"io"

	"github.com/uservers/baggr/pkg/spec"
)

var ErrIsDir = errors.New("path is a directory")

// Reader is an interface that defines the method to read the
// files which will be packaged in the new package.
type Reader interface {
	OpenPath(context.Context, *spec.File) (io.Reader, error)
	ListDirFiles(context.Context, string) ([]*spec.File, error)
}

// Writer abstracts an object that takes files from the reader and writes
// them to the destination
type Writer interface {
	CopyPaths(context.Context, Reader, []*spec.File) error
	Path() string
}
