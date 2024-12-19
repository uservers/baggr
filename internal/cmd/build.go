// SPDX-FileCopyrightText: Copyright 2024 U Servers Comunicaciones, S.C.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uservers/baggr/pkg/build"
	"github.com/uservers/baggr/pkg/builder"
)

func addBuild(parentCmd *cobra.Command) {
	opts := build.Default

	buildCmd := &cobra.Command{
		Short:             fmt.Sprintf("%s build: build OS packages", appname),
		Long:              fmt.Sprintf(`%s build: build OS pacakges from a manifest`, appname),
		Use:               "build",
		SilenceUsage:      false,
		SilenceErrors:     false,
		PersistentPreRunE: initLogging,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if opts.ManifestPath != "" {
					return errors.New("cannot define -m and pass a manifest")
				}
				opts.ManifestPath = args[0]
			}
			if err := opts.Validate(); err != nil {
				return fmt.Errorf("validating options: %w", err)
			}

			// Args are already validated
			cmd.SilenceErrors = true

			// Run the build
			return builder.New().Build(context.Background(), &opts)
		},
	}
	buildCmd.PersistentFlags().StringVarP(
		&opts.ManifestPath, "manifest", "m", "", "path to the package manifest",
	)
	buildCmd.PersistentFlags().StringVarP(
		&opts.Version.String, "version", "v", "", "version to set in the package",
	)
	buildCmd.PersistentFlags().StringVarP(
		&opts.Version.Release, "release", "r", "0", "release to set in the package",
	)
	parentCmd.AddCommand(buildCmd)
}
