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
	ContractID string // NRP: references a Contract for soft constraint evaluation
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
	Day               int    // roster day (0 = no day constraint)
	ShiftType         string // NRP: shift type identifier (e.g. "early", "late", "night")
	Mandatory         bool   // NRP: true if this demand must be covered (hard constraint)
	DemandGroup       string // NRP: groups items for coverage evaluation (e.g. "day1-early-general")
}

// TravelEntry represents the travel time between two locations.
type TravelEntry struct {
	From    string
	To      string
	Minutes int
}

// --- NRP / INRC-II Context Types ---

// Contract defines nurse employment constraints for soft constraint evaluation.
type Contract struct {
	ID                       string
	MinAssignments           int
	MaxAssignments           int
	MinConsecutiveWorkingDays int
	MaxConsecutiveWorkingDays int
	MinConsecutiveDaysOff    int
	MaxConsecutiveDaysOff    int
	MaxWorkingWeekends       int
	CompleteWeekend          bool
}

// ShiftTypeInfo defines a shift type with its consecutive assignment limits.
type ShiftTypeInfo struct {
	ID                       string
	StartMinute              int
	EndMinute                int
	MinConsecutiveAssignments int
	MaxConsecutiveAssignments int
}

// ForbiddenSuccession defines an illegal shift type transition.
// A resource must not work SuccessorShift on the day following PrecedingShift.
type ForbiddenSuccession struct {
	PrecedingShift string
	SuccessorShift string
}

// Request represents a nurse preference for or against a specific assignment.
type Request struct {
	ResourceID string
	Day        int
	ShiftType  string // empty means day-level request
	Type       string // "shiftOn", "shiftOff", "dayOn", "dayOff"
	Weight     int    // penalty weight if violated
}

// CoverageRequirement defines staffing needs for a day/shift/skill combination.
type CoverageRequirement struct {
	Day       int
	ShiftType string
	Skill     string
	Minimum   int
	Optimal   int
}
