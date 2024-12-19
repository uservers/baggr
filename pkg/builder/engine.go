// SPDX-FileCopyrightText: Copyright 2024 U Servers Comunicaciones, S.C.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/uservers/baggr/pkg/baggr"
	"github.com/uservers/baggr/pkg/build"
	"github.com/uservers/baggr/pkg/source"
	"github.com/uservers/baggr/pkg/spec"
)

func New() *Engine {
	return &Engine{
		Options:        build.Default,
		implementation: &defaultEngineImplementation{},
	}
}

type Engine struct {
	Options        build.Options
	implementation EngineImplementation
}

type Result struct{}

// Build takes a manifest and builds the RPM files
func (eng *Engine) Build(ctx context.Context, opts *build.Options) error {
	// Append the build context:
	buildContext := &build.Context{}
	ctx = context.WithValue(ctx, build.ContextKey{}, buildContext)

	// Read the package manifest
	manifest, err := eng.implementation.ParseManifest(ctx, opts.ManifestPath)
	if err != nil {
		return fmt.Errorf("parsing manifest: %w", err)
	}

	logrus.Infof("Manifest: \n%+v", manifest)

	reader, err := getSourceReader(manifest)
	if err != nil {
		return fmt.Errorf("getting reader: %w", err)
	}
	opts.SourceReader = reader

	// ENsure we have a version to work with
	if err := eng.implementation.EnsureVersion(ctx, opts); err != nil {
		return fmt.Errorf("ensuring package versions: %w", err)
	}

	// Cycle all packagte types and build them
	for _, t := range opts.PackageTypes {
		worker := eng.GetPackageWorker(t)
		if worker == nil {
			return fmt.Errorf("no bagger worker defined for type %s", t)
		}
		logrus.Infof("Building %s", t)
		_, err := worker.BuildPackages(ctx, manifest, opts)
		if err != nil {
			return fmt.Errorf("building %s packages: %w", t, err)
		}
	}

	return nil
}

// getSourceReader returns a source.Reader appropriate for the files defined
// in the manifest.
func getSourceReader(manifest *spec.Manifest) (source.Reader, error) {
	for _, c := range append([]*spec.Component{&manifest.Component}, manifest.Components...) {
		for _, f := range c.Files {
			if f.Source != "%DIR%" && strings.HasPrefix(f.Source, "/") {
				return nil, errors.New("cannot build a proper fs source for now (found absolute paths)")
			}
		}
	}

	if sfs, ok := os.DirFS(".").(fs.StatFS); ok {
		return source.NewFilesystemReader(sfs), nil
	}
	return nil, fmt.Errorf("filesystem not usable, need to implement io.StatFS")
}

// Returns a package worker for the specified type
func (eng *Engine) GetPackageWorker(t spec.PackageType) baggr.Worker {
	if _, ok := baggr.WorkerTypes[t]; ok {
		return baggr.WorkerTypes[t]
	}
	return nil
}
