package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// OptimisationContext represents the complete optimisation problem.
//
// It bundles all inputs required by algorithms into a single stable contract.
type OptimisationContext struct {
	items               []workitem.WorkItem
	resources           []ResourceInput
	workItems           []WorkItemInput
	travelMatrix        []TravelEntry
	existingAssignments []assignment.Assignment
	weights             ObjectiveWeights
	profile             AlgorithmProfile

	// NRP / INRC-II context.
	contracts             []Contract
	shiftTypes            []ShiftTypeInfo
	forbiddenSuccessions  []ForbiddenSuccession
	requests              []Request
	coverageRequirements  []CoverageRequirement
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

// WithExistingPlan returns a copy of the context with an existing plan set.
// Search algorithms may use this as a warm start.
func (c OptimisationContext) WithExistingPlan(assignments []assignment.Assignment) OptimisationContext {
	cp := make([]assignment.Assignment, len(assignments))
	copy(cp, assignments)
	c.existingAssignments = cp
	return c
}

// WithWeights returns a copy of the context with custom objective weights.
func (c OptimisationContext) WithWeights(w ObjectiveWeights) OptimisationContext {
	c.weights = w
	return c
}

// Weights returns the objective weights. Returns defaults if none set.
func (c OptimisationContext) Weights() ObjectiveWeights {
	if c.weights == (ObjectiveWeights{}) {
		return DefaultWeights()
	}
	return c.weights
}

// WithProfile returns a copy of the context with the given algorithm profile.
func (c OptimisationContext) WithProfile(p AlgorithmProfile) OptimisationContext {
	c.profile = p
	return c
}

// Profile returns the algorithm profile. Returns default if none set.
func (c OptimisationContext) Profile() AlgorithmProfile {
	if c.profile == (AlgorithmProfile{}) {
		return DefaultProfile()
	}
	return c.profile
}

// ExistingAssignments returns the existing plan assignments if present.
// Returns nil if no existing plan was provided.
func (c OptimisationContext) ExistingAssignments() []assignment.Assignment {
	if len(c.existingAssignments) == 0 {
		return nil
	}
	cp := make([]assignment.Assignment, len(c.existingAssignments))
	copy(cp, c.existingAssignments)
	return cp
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

// --- NRP / INRC-II context builders ---

// WithContracts returns a copy of the context with contracts set.
func (c OptimisationContext) WithContracts(contracts []Contract) OptimisationContext {
	cp := make([]Contract, len(contracts))
	copy(cp, contracts)
	c.contracts = cp
	return c
}

// WithShiftTypes returns a copy of the context with shift types set.
func (c OptimisationContext) WithShiftTypes(shiftTypes []ShiftTypeInfo) OptimisationContext {
	cp := make([]ShiftTypeInfo, len(shiftTypes))
	copy(cp, shiftTypes)
	c.shiftTypes = cp
	return c
}

// WithForbiddenSuccessions returns a copy of the context with forbidden successions set.
func (c OptimisationContext) WithForbiddenSuccessions(successions []ForbiddenSuccession) OptimisationContext {
	cp := make([]ForbiddenSuccession, len(successions))
	copy(cp, successions)
	c.forbiddenSuccessions = cp
	return c
}

// WithRequests returns a copy of the context with nurse requests set.
func (c OptimisationContext) WithRequests(requests []Request) OptimisationContext {
	cp := make([]Request, len(requests))
	copy(cp, requests)
	c.requests = cp
	return c
}

// WithCoverageRequirements returns a copy of the context with coverage requirements set.
func (c OptimisationContext) WithCoverageRequirements(reqs []CoverageRequirement) OptimisationContext {
	cp := make([]CoverageRequirement, len(reqs))
	copy(cp, reqs)
	c.coverageRequirements = cp
	return c
}

// --- NRP / INRC-II context accessors ---

// Contracts returns the NRP contracts. Returns nil if none set.
func (c OptimisationContext) Contracts() []Contract {
	if len(c.contracts) == 0 {
		return nil
	}
	cp := make([]Contract, len(c.contracts))
	copy(cp, c.contracts)
	return cp
}

// ShiftTypes returns the NRP shift type definitions. Returns nil if none set.
func (c OptimisationContext) ShiftTypes() []ShiftTypeInfo {
	if len(c.shiftTypes) == 0 {
		return nil
	}
	cp := make([]ShiftTypeInfo, len(c.shiftTypes))
	copy(cp, c.shiftTypes)
	return cp
}

// ForbiddenSuccessions returns the illegal shift successions. Returns nil if none set.
func (c OptimisationContext) ForbiddenSuccessions() []ForbiddenSuccession {
	if len(c.forbiddenSuccessions) == 0 {
		return nil
	}
	cp := make([]ForbiddenSuccession, len(c.forbiddenSuccessions))
	copy(cp, c.forbiddenSuccessions)
	return cp
}

// Requests returns the nurse requests. Returns nil if none set.
func (c OptimisationContext) Requests() []Request {
	if len(c.requests) == 0 {
		return nil
	}
	cp := make([]Request, len(c.requests))
	copy(cp, c.requests)
	return cp
}

// CoverageRequirements returns the coverage requirements. Returns nil if none set.
func (c OptimisationContext) CoverageRequirements() []CoverageRequirement {
	if len(c.coverageRequirements) == 0 {
		return nil
	}
	cp := make([]CoverageRequirement, len(c.coverageRequirements))
	copy(cp, c.coverageRequirements)
	return cp
}
