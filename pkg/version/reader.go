package version

import "context"

type Reader interface {
	GetLastVersion(context.Context) (*Spec, error)
	ComputeNextVersion(context.Context, *Spec) (*Spec, error)
}

type MockReader struct{}

func (mr *MockReader) GetLastVersion(_ context.Context) (*Spec, error) {
	return &Spec{
		String:  "v1.0.0",
		Release: "0",
	}, nil
}

func (mr *MockReader) ComputeNextVersion(_ context.Context, _ *Spec) (*Spec, error) {
	return &Spec{
		String:  "v1.0.1",
		Release: "0",
	}, nil
}
