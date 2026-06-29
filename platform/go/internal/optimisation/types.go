package optimisation

// ResourceInput provides resource constraint information to the optimiser.
//
// This is not a domain object. It is structured input that the application
// layer prepares by interpreting business knowledge from resource details.
type ResourceInput struct {
	ResourceID string
	Capacity   int
	Available  bool
	Skills     []string
}

// WorkItemInput provides work item optimisation input to the optimiser.
//
// This is not a domain object. It is structured input that the application
// layer prepares by interpreting business knowledge from work item details.
type WorkItemInput struct {
	WorkItemID    string
	Priority      int
	RequiredSkill string
	Duration      int // minutes required to complete this work item
}
