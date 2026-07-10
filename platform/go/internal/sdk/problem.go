package sdk

import owpsdk "github.com/timdodgson/open-workforce-platform/owp-sdk/sdk"

type (
	Problem           = owpsdk.Problem
	DatasetLoader     = owpsdk.DatasetLoader
	InstanceMeta      = owpsdk.InstanceMeta
	ProblemLoader     = owpsdk.ProblemLoader
	ProblemDefaults   = owpsdk.ProblemDefaults
	ProblemDescriptor = owpsdk.ProblemDescriptor
)

var (
	RegisterProblem = owpsdk.RegisterProblem
	GetProblem      = owpsdk.GetProblem
	Problems        = owpsdk.Problems
)
