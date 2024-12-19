// SPDX-FileCopyrightText: Copyright 2024 U Servers Comunicaciones, S.C.
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"strings"
)

type Manifest struct {
	Component  `yaml:",inline"`
	URL        string
	Version    string
	Release    string
	Components []*Component
}

type Component struct {
	Name        string
	License     string
	Summary     string
	Description string
	NoDeps      bool
	Requires    []string
	Files       []*File
}

func (c *Component) RequiresString() string {
	return strings.Join(c.Requires, ", ")
}

type File struct {
	Source      string
	Destination string
	Mode        string
	UID         string
	GID         string
}

// DeepCopy returns a pointer to a copy of the manifest
func (m *Manifest) DeepCopy() *Manifest {
	c := m.Component.DeepCopy()
	m2 := &Manifest{
		Component:  *c,
		URL:        m.URL,
		Version:    m.Version,
		Release:    m.Release,
		Components: []*Component{},
	}

	for _, c := range m.Components {
		m2.Components = append(m2.Components, c.DeepCopy())
	}

	return m2
}

func (c *Component) DeepCopy() *Component {
	c2 := Component{
		Name:        c.Name,
		License:     c.License,
		Summary:     c.Summary,
		Description: c.Description,
		Requires:    c.Requires,
		Files:       []*File{},
	}

	for _, f := range c.Files {
		c2.Files = append(c2.Files, f.DeepCopy())
	}
	return &c2
}

func (f *File) DeepCopy() *File {
	return &File{
		Source:      f.Source,
		Destination: f.Destination,
		Mode:        f.Mode,
		UID:         f.UID,
		GID:         f.GID,
	}
}

// EnsureDefaults makes sure that UIDs, GIDs and Modes are populated
func (m *Manifest) EnsureDefaults() {
	m.Component.EnsureDefaults()
	for i := range m.Components {
		m.Components[i].EnsureDefaults()
	}
}

func (c *Component) EnsureDefaults() {
	for i := range c.Files {
		if c.Files[i].Mode == "" {
			c.Files[i].Mode = "-"
		}

		if c.Files[i].UID == "" {
			c.Files[i].UID = "-"
		}

		if c.Files[i].GID == "" {
			c.Files[i].GID = "-"
		}
	}
}
