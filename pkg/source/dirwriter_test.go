// SPDX-FileCopyrightText: Copyright 2024 U Servers Comunicaciones, S.C.
// SPDX-License-Identifier: Apache-2.0

package source

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/liamg/memoryfs"
	"github.com/stretchr/testify/require"
	"github.com/uservers/baggr/pkg/spec"
)

type failReader struct{}

func (*failReader) Read(b []byte) (int, error) {
	return 0, errors.New("ERROR")
}

func TestDWCopyPaths(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	bfs := memoryfs.New()
	require.NoError(t, bfs.MkdirAll("/dir1", os.FileMode(0o755)))
	require.NoError(t, bfs.MkdirAll("/dir2/sub", os.FileMode(0o755)))
	require.NoError(t, bfs.WriteFile("/test.txt", []byte("Hola"), os.FileMode(0o644)))
	require.NoError(t, bfs.WriteFile("/dir1/test.txt", []byte("Hola"), os.FileMode(0o644)))
	require.NoError(t, bfs.WriteFile("/dir2/test.txt", []byte("Hola"), os.FileMode(0o644)))
	require.NoError(t, bfs.WriteFile("/dir2/sub/test.txt", []byte("Hola"), os.FileMode(0o644)))

	fsr := NewFilesystemReader(bfs)

	for _, tc := range []struct {
		name      string
		specFiles []*spec.File
		expected  []string
		mustErr   bool
	}{
		{"file", []*spec.File{{Source: "/test.txt"}}, []string{"/test.txt"}, false},
		{"dir", []*spec.File{{Source: "/dir1"}}, []string{"/dir1/test.txt"}, false},
		{"dir-and-file", []*spec.File{{Source: "/test.txt"}, {Source: "/dir1"}}, []string{"/dir1/test.txt", "/test.txt"}, false},
		{"not-found", []*spec.File{{Source: "/404.txt"}}, []string{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dirPath := t.TempDir()
			dw := NewDirWriter(dirPath)
			err := dw.CopyPaths(ctx, fsr, tc.specFiles)
			if tc.mustErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			resFiles := []string{}
			require.NoError(t, filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				resFiles = append(resFiles, strings.TrimPrefix(path, dirPath))
				return nil
			}))

			slices.Sort(resFiles)
			require.Equal(t, tc.expected, resFiles)
		})
	}
}

func TestDWCopyDirectory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	bfs := memoryfs.New()
	require.NoError(t, bfs.MkdirAll("/dir1", os.FileMode(0o755)))
	require.NoError(t, bfs.MkdirAll("/dir2/sub", os.FileMode(0o755)))
	require.NoError(t, bfs.WriteFile("/dir1/test.txt", []byte("Hola"), os.FileMode(0o644)))
	require.NoError(t, bfs.WriteFile("/dir2/test.txt", []byte("Hola"), os.FileMode(0o644)))
	require.NoError(t, bfs.WriteFile("/dir2/sub/test.txt", []byte("Hola"), os.FileMode(0o644)))

	fsr := NewFilesystemReader(bfs)

	for _, tc := range []struct {
		name     string
		specFile *spec.File
		expected []string
		mustErr  bool
	}{
		{"normal", &spec.File{Source: "/dir1"}, []string{"/dir1/test.txt"}, false},
		{"subdir", &spec.File{Source: "/dir2"}, []string{"/dir2/sub/test.txt", "/dir2/test.txt"}, false},
		{"nonexistent", &spec.File{Source: "/dir3"}, []string{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dirPath := t.TempDir()
			dw := NewDirWriter(dirPath)
			err := dw.CopyDirectory(ctx, fsr, tc.specFile)
			if tc.mustErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			resFiles := []string{}
			require.NoError(t, filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				resFiles = append(resFiles, strings.TrimPrefix(path, dirPath))
				return nil
			}))

			slices.Sort(resFiles)

			require.Equal(t, tc.expected, resFiles)
		})
	}
}

func TestDWCopyFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fileData := []byte("Hola y adios")

	for _, tc := range []struct {
		name     string
		specFile *spec.File
		prepare  func() (*DirWriter, io.Reader)
		mustErr  bool
	}{
		{
			"normal",
			&spec.File{Destination: "/test.txt"},
			func() (*DirWriter, io.Reader) {
				t.Helper()
				dw := NewDirWriter(t.TempDir())
				var b bytes.Buffer
				_, err := b.Write(fileData)
				require.NoError(t, err)
				return dw, &b
			}, false,
		},
		{
			"reader-fails",
			&spec.File{Destination: "/test.txt"},
			func() (*DirWriter, io.Reader) {
				t.Helper()
				dw := NewDirWriter(t.TempDir())
				return dw, &failReader{}
			}, true,
		},
		{
			"source-not-set",
			&spec.File{Destination: "/test.txt"},
			func() (*DirWriter, io.Reader) {
				t.Helper()
				dw := NewDirWriter("")
				var b bytes.Buffer
				_, err := b.Write(fileData)
				require.NoError(t, err)
				return dw, &b
			}, true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dw, b := tc.prepare()

			err := dw.CopyFile(ctx, b, tc.specFile)
			if tc.mustErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			data, err := os.ReadFile(filepath.Join(dw.path, tc.specFile.Destination))
			require.NoError(t, err)
			require.Equal(t, data, fileData)
		})
	}
}
