package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// OptimisationContext represents the complete optimisation problem.
//
// It bundles all inputs required by algorithms into a single stable contract.
// This allows future inputs (travel matrices, time windows, objectives) to be
// added without changing algorithm signatures.
type OptimisationContext struct {
	items     []workitem.WorkItem
	resources []ResourceInput
	workItems []WorkItemInput
}

// NewContext creates an OptimisationContext from the provided inputs.
func NewContext(items []workitem.WorkItem, resources []ResourceInput, workItems []WorkItemInput) OptimisationContext {
	itemsCopy := make([]workitem.WorkItem, len(items))
	copy(itemsCopy, items)

	resourcesCopy := make([]ResourceInput, len(resources))
	copy(resourcesCopy, resources)

	workItemsCopy := make([]WorkItemInput, len(workItems))
	copy(workItemsCopy, workItems)

	return OptimisationContext{
		items:     itemsCopy,
		resources: resourcesCopy,
		workItems: workItemsCopy,
	}
}

// Items returns the work item domain objects.
func (c OptimisationContext) Items() []workitem.WorkItem {
	cp := make([]workitem.WorkItem, len(c.items))
	copy(cp, c.items)
	return cp
}

// Resources returns the resource inputs.
func (c OptimisationContext) Resources() []ResourceInput {
	cp := make([]ResourceInput, len(c.resources))
	copy(cp, c.resources)
	return cp
}

// WorkItems returns the work item inputs.
func (c OptimisationContext) WorkItems() []WorkItemInput {
	cp := make([]WorkItemInput, len(c.workItems))
	copy(cp, c.workItems)
	return cp
}
