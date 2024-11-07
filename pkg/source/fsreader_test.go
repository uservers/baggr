// SPDX-FileCopyrightText: Copyright 2024 U Servers Comunicaciones, S.C.
// SPDX-License-Identifier: Apache-2.0

package source

import (
	"context"
	"errors"
	"io"
	"os"
	"slices"
	"testing"

	"github.com/liamg/memoryfs"
	"github.com/stretchr/testify/require"
	"github.com/uservers/baggr/pkg/spec"
)

func TestListDirFiles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	bfs := memoryfs.New()
	require.NoError(t, bfs.WriteFile("/test.txt", []byte("Hola"), os.FileMode(0o644)))
	require.NoError(t, bfs.MkdirAll("/test/dir", os.FileMode(0o755)))
	require.NoError(t, bfs.WriteFile("/test/dir/test2.txt", []byte("Hola"), os.FileMode(0o644)))

	fsr := NewFilesystemReader(bfs)

	for _, tc := range []struct {
		name     string
		path     string
		mustErr  bool
		fileList []string // Miust be sorted
	}{
		{"root", "/", false, []string{"/test.txt", "/test/dir/test2.txt"}},
		{"subdir", "test", false, []string{"test/dir/test2.txt"}},
		{"non-exitent", "not-found", true, []string{}},
		{"path-is-file", "/test.txt", false, []string{"/test.txt"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := fsr.ListDirFiles(ctx, tc.path)
			if tc.mustErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			resList := []string{}
			for _, e := range res {
				resList = append(resList, e.Source)
			}
			slices.Sort(resList)
			require.Equal(t, tc.fileList, resList)
		})
	}
}

func TestFSROpenPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	bfs := memoryfs.New()
	require.NoError(t, bfs.WriteFile("/test.txt", []byte("Hola"), os.FileMode(0o644)))
	require.NoError(t, bfs.MkdirAll("/test/dir", os.FileMode(0o755)))

	fsr := NewFilesystemReader(bfs)

	for _, tc := range []struct {
		name      string
		specFile  *spec.File
		mustBeDir bool
		mustErr   bool
	}{
		{"normal", &spec.File{Source: "/test.txt"}, false, false},
		{"dir", &spec.File{Source: "/test/dir"}, true, false},
		{"non-existent", &spec.File{Source: "/not-found.txt"}, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rdr, err := fsr.OpenPath(ctx, tc.specFile)
			if tc.mustBeDir {
				require.Error(t, err)
				require.True(t, errors.Is(err, ErrIsDir))
				return
			}

			if tc.mustErr {
				require.Error(t, err)
				return
			}

			require.NotNil(t, rdr)
			require.NoError(t, rdr.(io.Closer).Close())
		})
	}

}
