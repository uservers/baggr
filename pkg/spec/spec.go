// SPDX-FileCopyrightText: Copyright 2024 U Servers Comunicaciones, S.C.
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// NewManifestFromFile parses a file and returns a manifest struct
func NewManifestFromFile(path string) (*Manifest, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	manifest := &Manifest{}
	if err := yaml.Unmarshal(f, manifest); err != nil {
		return nil, err
	}
	logrus.Infof("parsed manifest from %s", path)
	return manifest, nil
}
