// SPDX-FileCopyrightText: Copyright 2024 U Servers Comunicaciones, S.C.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/uservers/baggr/internal/cmd"
)

func main() {
	root := cmd.New()

	if err := root.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
