// Package sdk registers BYOD problem domains for the Open Workforce Platform.
package sdk

import "github.com/timdodgson/open-workforce-platform/owp-sdk/searchdef"

// Problem is the stable search engine contract for BYOD domains.
type Problem = searchdef.Problem

// DatasetLoader loads a Problem from an instance file path.
type DatasetLoader = searchdef.DatasetLoader

// InstanceMeta carries display metadata after loading an instance.
type InstanceMeta struct {
	Name         string
	InstancePath string
	Fields       map[string]string
	Data         any // optional domain-specific payload for finalize hooks
}

// ProblemLoader loads a Problem and instance metadata from a file path.
type ProblemLoader func(instancePath string) (Problem, InstanceMeta, error)

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
	Command  string // deprecated legacy CLI name (optional)
	Usage    string
	Load     ProblemLoader
	Defaults ProblemDefaults

	// SolveUI configures generic `owp solve` display when no platform-specific hooks are registered.
	Title          string
	PolicyDomain   string // defaults to Name
	ObjectiveLabel string // defaults to "Objective"
}
