// Package rpm is a baggr implementation that builds RPM packages
package rpm

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/uservers/baggr/pkg/build"
	"github.com/uservers/baggr/pkg/source"
	"github.com/uservers/baggr/pkg/spec"
)

const (
	DIR = "%DIR%"
)

//go:embed template/spec.tmpl
var Template string

func New() *Worker {
	return &Worker{
		implementation: &defaultImplementation{},
	}
}

type Worker struct {
	implementation Implementation
}

// BuildPackages takes a manifest and build the rpms defined in it
func (w *Worker) BuildPackages(ctx context.Context, manifest *spec.Manifest, opts *build.Options) (build.Result, error) {
	var results build.Result
	// Create a temp directory to use as the build root
	tmp, err := os.MkdirTemp("", "baggr-rpmbuildroot-*")
	if err != nil {
		return results, fmt.Errorf("creating temporary BUILD_ROOT: %w", err)
	}
	sourceWriter := source.NewDirWriter(tmp)

	if err := w.implementation.CopySourceFiles(ctx, opts, sourceWriter, manifest); err != nil {
		return results, fmt.Errorf("copying package files: %w", err)
	}

	rpmSpecPath, err := w.implementation.BuildRpmSpec(ctx, opts, sourceWriter, manifest)
	if err != nil {
		return results, fmt.Errorf("building RPM spec: %w", err)
	}

	results, err = w.implementation.BuildRpms(ctx, opts, rpmSpecPath, sourceWriter)
	if err != nil {
		return results, fmt.Errorf("packaging RPMs: %w", err)
	}

	return results, nil
}
