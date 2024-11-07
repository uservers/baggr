package rpm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uservers/baggr/pkg/build"
	"github.com/uservers/baggr/pkg/source"
	"github.com/uservers/baggr/pkg/spec"
	"github.com/uservers/baggr/pkg/version"
)

func TestBuildRpmSpec(t *testing.T) {
	type specCheck struct {
		version string
	}
	di := defaultImplementation{}
	// Build a context to pass a version
	ctx := context.WithValue(context.Background(), build.ContextKey{}, build.Context{
		Version: &version.Spec{
			String:  "v1.0.0",
			Release: "1",
		},
	})

	// Source for the source writer
	swTemp := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(swTemp, "subdir"), os.FileMode(0o755)))
	require.NoError(t, os.WriteFile(filepath.Join(swTemp, "subdir", "test.txt"), []byte("hey"), os.FileMode(0o644)))

	require.NoError(t, os.Mkdir(filepath.Join(swTemp, "docs"), os.FileMode(0o755)))
	require.NoError(t, os.WriteFile(filepath.Join(swTemp, "docs", "index.html"), []byte("hey"), os.FileMode(0o644)))

	man := &spec.Manifest{
		Component: spec.Component{
			Name:        "test",
			License:     "Apache-2.0",
			Summary:     "Test project",
			Description: "Empty project to test",
			Requires: []string{
				"comp1", "comp2",
			},
			Files: []*spec.File{
				{
					Source:      "%DIR%",
					Destination: "/testdir",
				},
				{
					Source:      "/subdir",
					Destination: "/subdir",
				},
				{
					Source:      "/testfile.txt",
					Destination: "/testfile.txt",
				},
			},
		},
		Components: []*spec.Component{
			{
				Name:        "docs",
				License:     "Apache-2.0",
				Summary:     "Documentos del deste",
				Description: "Documentos del programa este para que leas",
				Requires:    []string{"man"},
				Files: []*spec.File{
					{
						Source:      "docs",
						Destination: "docs",
					},
				},
			},
		},
	}

	path, err := di.BuildRpmSpec(ctx, &build.Options{}, source.NewDirWriter(swTemp), man)
	require.NoError(t, err)
	t.Logf("Spec file: %s", path)
	require.FileExists(t, path)
}

func TestFindFiles(t *testing.T) {
	logData, err := os.ReadFile("testdata/sample.log")
	require.NoError(t, err)
	files := findFiles(string(logData))
	require.Equal(t, []string{"/home/urbano/rpmbuild/RPMS/noarch/test-docs-v1.0.0-1.noarch.rpm", "/home/urbano/rpmbuild/RPMS/noarch/test-v1.0.0-1.noarch.rpm"}, files)
}
