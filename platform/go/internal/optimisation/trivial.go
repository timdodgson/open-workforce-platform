// Package optimisation provides optimisation capabilities for the platform.
//
// The trivial optimiser is the first implementation. It accepts all work items
// without filtering, scoring or reordering. It exists to prove the vertical
// slice works end-to-end.
package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// Solve accepts work items and produces an OptimisedPlan.
//
// The trivial implementation selects all work items.
// Future implementations will consider resources, constraints and objectives.
func Solve(items []workitem.WorkItem) (plan.OptimisedPlan, error) {
	return plan.New(items)
}
