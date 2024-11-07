// SPDX-FileCopyrightText: Copyright 2024 U Servers Comunicaciones, S.C.
// SPDX-License-Identifier: Apache-2.0

package source

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/uservers/baggr/pkg/spec"
)

// DirWriter implements a writer that writes to a directory in the filesystem
type DirWriter struct {
	path string
}

func NewDirWriter(path string) *DirWriter {
	return &DirWriter{
		path: path,
	}
}

func (dw *DirWriter) Path() string {
	return dw.path
}

// CopyPaths copies all the file paths received to the DirWriter path
func (dw *DirWriter) CopyPaths(ctx context.Context, r Reader, files []*spec.File) error {
	if dw.path == "" {
		return fmt.Errorf("unable to copy file, no path defined")
	}
	for _, specFile := range files {
		f, openErr := r.OpenPath(ctx, specFile)
		var err error
		switch {
		case openErr != nil && errors.Is(openErr, ErrIsDir):
			err = dw.CopyDirectory(ctx, r, specFile)
		case openErr != nil && !errors.Is(openErr, ErrIsDir):
			return fmt.Errorf("opening path %q: %w", specFile.Source, err)
		default:
			err = dw.CopyFile(ctx, f, specFile)
		}

		if err != nil {
			return fmt.Errorf("attempting to open path %q: %s", specFile.Source, err)
		}
	}
	return nil
}

// CopyDirectory copies a directory recursively
func (dw DirWriter) CopyDirectory(ctx context.Context, r Reader, specFile *spec.File) error {
	if dw.path == "" {
		return fmt.Errorf("unable to copy file, no path defined")
	}

	fileList, err := r.ListDirFiles(ctx, specFile.Source)
	if err != nil {
		return fmt.Errorf("listing directory files")
	}

	for _, f := range fileList {
		// replace paths in the destination
		if specFile.Destination != "" {
			f.Destination = strings.ReplaceAll(f.Source, specFile.Source, specFile.Destination)
		}
		fr, err := r.OpenPath(ctx, f)
		if err != nil {
			return fmt.Errorf("opening path from directory: %w", err)
		}
		if err := dw.CopyFile(ctx, fr, f); err != nil {
			return fmt.Errorf("copying file from directory: %w", err)
		}
	}

	return nil
}

// CopyFile copies the data stream we got from the reader to a file in the
// package filesystem
func (dw DirWriter) CopyFile(ctx context.Context, r io.Reader, specFile *spec.File) error {
	if dw.path == "" {
		return fmt.Errorf("unable to copy file, no path defined")
	}
	destPath := specFile.Destination
	if destPath == "" {
		destPath = specFile.Source
	}

	destPath = filepath.Join(dw.path, path.Clean(destPath))

	// Create all the directories
	dir := path.Dir(destPath)
	if !strings.HasPrefix(dir, dw.path) {
		return fmt.Errorf("access violation")
	}

	if err := os.MkdirAll(dir, os.FileMode(0o755)); err != nil {
		return fmt.Errorf("creating directory in package filesystem: %w", err)
	}

	// Create the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating file in package filesystem")
	}

	// Copy the reader stream
	_, err = io.Copy(destFile, r)
	if err != nil {
		return fmt.Errorf("copying data stream: %w", err)
	}

	// Close the destination file
	if _, ok := r.(io.Closer); ok {
		r.(io.Closer).Close()
	}

	return nil
}
