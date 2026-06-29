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
	ShiftStart int // minutes from midnight
	ShiftEnd   int // minutes from midnight
	Location   string
}

// WorkItemInput provides work item optimisation input to the optimiser.
//
// This is not a domain object. It is structured input that the application
// layer prepares by interpreting business knowledge from work item details.
type WorkItemInput struct {
	WorkItemID        string
	Priority          int
	RequiredSkill     string
	Duration          int // minutes required to complete this work item
	EarliestStart     int // minutes from midnight (0 = no constraint)
	LatestFinish      int // minutes from midnight (0 = no constraint)
	Location          string
	PreferredResource string // soft constraint: preferred resource ID
}

// TravelEntry represents the travel time between two locations.
type TravelEntry struct {
	From    string
	To      string
	Minutes int
}
