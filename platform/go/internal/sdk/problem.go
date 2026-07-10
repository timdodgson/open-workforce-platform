package sdk

import "github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"

// Problem is the stable search engine contract for BYOD domains.
type Problem = searchdef.Problem

// DatasetLoader loads a Problem from an instance file path.
type DatasetLoader = searchdef.DatasetLoader

// InstanceMeta carries display metadata after loading an instance.
type InstanceMeta struct {
	Name   string
	Fields map[string]string
}

// ProblemLoader loads a searchdef.Problem and instance metadata from a file path.
type ProblemLoader func(instancePath string) (searchdef.Problem, InstanceMeta, error)

// ProblemDefaults holds CLI defaults for a registered problem domain.
type ProblemDefaults struct {
	Mode        string
	Iterations  int
	Temperature float64
	Seed        int64
}

// ProblemDescriptor registers a BYOD domain with the platform SDK.
type ProblemDescriptor struct {
	Name     string
	Command  string
	Usage    string
	Load     ProblemLoader
	Defaults ProblemDefaults
}
