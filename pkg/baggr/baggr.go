package baggr

import (
	"context"

	"github.com/uservers/baggr/pkg/build"
	"github.com/uservers/baggr/pkg/rpm"
	"github.com/uservers/baggr/pkg/spec"
)

var WorkerTypes = map[spec.PackageType]Worker{
	spec.PackageTypeRPM: rpm.New(),
}

type Worker interface {
	BuildPackages(context.Context, *spec.Manifest, *build.Options) (build.Result, error)
}
