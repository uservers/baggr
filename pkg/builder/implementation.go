// SPDX-FileCopyrightText: Copyright 2024 U Servers Comunicaciones, S.C.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"fmt"
	"os"

	"github.com/uservers/baggr/pkg/build"
	"github.com/uservers/baggr/pkg/spec"
)

type EngineImplementation interface {
	ParseManifest(context.Context, string) (*spec.Manifest, error)
	CheckSourceFiles(context.Context, *spec.Manifest) error
	EnsureVersion(context.Context, *build.Options) error
}

type defaultEngineImplementation struct{}

// EnsureVersion dynamically computes the next version and ships it in the build
// context if it is not specifically set on the options
func (di *defaultEngineImplementation) EnsureVersion(ctx context.Context, opts *build.Options) error {
	if opts.Version != nil && opts.Version.String != "" {
		return nil
	}

	if opts.VersionReader == nil {
		return fmt.Errorf("unable to compute next version, no version reader set")
	}

	buildContext := ctx.Value(build.ContextKey{})
	if buildContext == nil {
		return fmt.Errorf("unable to read build context")
	}

	lastVer, err := opts.VersionReader.GetLastVersion(ctx)
	if err != nil {
		return fmt.Errorf("reading last project version")
	}
	ver, err := opts.VersionReader.ComputeNextVersion(ctx, lastVer)
	if err != nil {
		return fmt.Errorf("computing next version: %w", err)
	}

	buildContext.(*build.Context).Version = ver
	return nil
}

// ParseManifest parses the yaml file and returns a new manifest object
func (di *defaultEngineImplementation) ParseManifest(_ context.Context, path string) (*spec.Manifest, error) {
	manifest, err := spec.NewManifestFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("parsing manifest file: %w", err)
	}
	return manifest, nil
}

// CheckSourceFiles
func (di *defaultEngineImplementation) CheckSourceFiles(ctx context.Context, manifest *spec.Manifest) error {
	// TODO: Extract cwd from context
	// TODO use source reader
	notFound := []string{}
	for _, file := range manifest.Files {
		if _, err := os.Stat(file.Source); err != nil {
			if err != os.ErrNotExist {
				return fmt.Errorf("error checking for file: %w", err)
			} else {
				notFound = append(notFound, file.Source)
			}
		}
	}

	// TODO(puerco): Components

	if len(notFound) > 0 {
		return fmt.Errorf("source files not found: %v", notFound)
	}
	return nil
}
