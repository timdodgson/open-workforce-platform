package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// OptimisationContext represents the complete optimisation problem.
//
// It bundles all inputs required by algorithms into a single stable contract.
type OptimisationContext struct {
	items        []workitem.WorkItem
	resources    []ResourceInput
	workItems    []WorkItemInput
	travelMatrix []TravelEntry
}

// NewContext creates an OptimisationContext from the provided inputs.
func NewContext(items []workitem.WorkItem, resources []ResourceInput, workItems []WorkItemInput) OptimisationContext {
	return NewContextWithTravel(items, resources, workItems, nil)
}

// NewContextWithTravel creates an OptimisationContext including a travel matrix.
func NewContextWithTravel(items []workitem.WorkItem, resources []ResourceInput, workItems []WorkItemInput, travel []TravelEntry) OptimisationContext {
	itemsCopy := make([]workitem.WorkItem, len(items))
	copy(itemsCopy, items)

	resourcesCopy := make([]ResourceInput, len(resources))
	copy(resourcesCopy, resources)

	workItemsCopy := make([]WorkItemInput, len(workItems))
	copy(workItemsCopy, workItems)

	travelCopy := make([]TravelEntry, len(travel))
	copy(travelCopy, travel)

	return OptimisationContext{
		items:        itemsCopy,
		resources:    resourcesCopy,
		workItems:    workItemsCopy,
		travelMatrix: travelCopy,
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

// TravelMatrix returns the travel time entries.
func (c OptimisationContext) TravelMatrix() []TravelEntry {
	cp := make([]TravelEntry, len(c.travelMatrix))
	copy(cp, c.travelMatrix)
	return cp
}
