package inrc2

import (
	"encoding/json"
	"math/rand"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// NRPProblem implements optimisation.Problem for Nurse Rostering (INRC-II).
//
// Each worker owns its own NRPProblem instance. The struct contains all
// pre-computed data needed for fast move generation and evaluation.
// No shared mutable state — workers are fully independent.
type NRPProblem struct {
	sc             Scenario
	wd             WeekData
	hist           History
	nurseSkills    []map[string]bool
	forbidden      map[string]bool
	histLastShift  []string
	ws             *ScoringWorkspace
	scoringMode    string
	numNurses      int
}

// NRPProblemConfig holds the parameters needed to construct an NRPProblem.
type NRPProblemConfig struct {
	Scenario    Scenario
	WeekData    WeekData
	History     History
	ScoringMode string // "official-penalty" or "soft-violation-count"
}

// NewNRPProblem creates a new NRP problem instance for a single worker.
// Each worker should call this to get its own independent instance.
func NewNRPProblem(cfg NRPProblemConfig) *NRPProblem {
	sc := cfg.Scenario
	wd := cfg.WeekData
	hist := cfg.History

	// Pre-compute nurse skills.
	nurseSkills := make([]map[string]bool, len(sc.Nurses))
	for i, n := range sc.Nurses {
		skills := make(map[string]bool, len(n.Skills))
		for _, s := range n.Skills {
			skills[s] = true
		}
		nurseSkills[i] = skills
	}

	// Pre-compute forbidden succession set.
	forbidden := buildForbiddenSet2(sc)

	// Pre-compute history last shift for day-0 succession validation.
	histLastShift := make([]string, len(sc.Nurses))
	for i, n := range sc.Nurses {
		for _, nh := range hist.NurseHistory {
			if nh.Nurse == n.ID {
				if nh.NumberOfConsecutiveDaysOff == 0 && nh.LastAssignedShiftType != "None" && nh.LastAssignedShiftType != "" {
					histLastShift[i] = nh.LastAssignedShiftType
				}
				break
			}
		}
	}

	scoringMode := cfg.ScoringMode
	if scoringMode == "" {
		scoringMode = "official-penalty"
	}

	return &NRPProblem{
		sc:            sc,
		wd:            wd,
		hist:          hist,
		nurseSkills:   nurseSkills,
		forbidden:     forbidden,
		histLastShift: histLastShift,
		ws:            NewScoringWorkspace(sc, wd, hist),
		scoringMode:   scoringMode,
		numNurses:     len(sc.Nurses),
	}
}

// --- optimisation.Problem implementation ---

// nrpMove stores the data needed to undo a nurse swap.
type nrpMove struct {
	nurseA int
	nurseB int
	day    int
	oldA   ShiftAssignment
	oldB   ShiftAssignment
}

// CreateInitialSolution builds a hard-feasible roster.
func (p *NRPProblem) CreateInitialSolution() (optimisation.Solution, error) {
	roster, err := BuildFeasibleRoster(p.sc, p.wd, p.hist)
	if err != nil {
		return nil, err
	}
	return roster, nil
}

// CloneSolution creates a deep copy of the roster.
func (p *NRPProblem) CloneSolution(s optimisation.Solution) optimisation.Solution {
	return s.(*Roster).Clone()
}

// Evaluate returns the soft penalty of the roster.
func (p *NRPProblem) Evaluate(s optimisation.Solution) int {
	return scorePenaltyWithMode(s.(*Roster), p.ws, p.scoringMode)
}

// TryMove generates a random nurse/day swap, validates hard constraints,
// and applies it if valid. Zero allocations in the hot path.
func (p *NRPProblem) TryMove(s optimisation.Solution, rng *rand.Rand) optimisation.MoveResult {
	roster := s.(*Roster)

	day := rng.Intn(7)
	nurseA := rng.Intn(p.numNurses)
	nurseB := rng.Intn(p.numNurses)
	if nurseA == nurseB {
		nurseB = (nurseA + 1) % p.numNurses
	}

	// Save old assignments for undo.
	oldA := roster.Get(nurseA, day)
	oldB := roster.Get(nurseB, day)

	// Try swap — validates hard constraints and applies if valid.
	rejectReason := swapNurses(roster, nurseA, nurseB, day, p.sc, p.nurseSkills, p.forbidden, p.histLastShift)
	if rejectReason >= 0 {
		return optimisation.MoveResult{Valid: false}
	}

	return optimisation.MoveResult{
		Valid: true,
		Move: nrpMove{
			nurseA: nurseA,
			nurseB: nurseB,
			day:    day,
			oldA:   oldA,
			oldB:   oldB,
		},
	}
}

// UndoMove reverts a previously applied swap.
func (p *NRPProblem) UndoMove(s optimisation.Solution, m optimisation.Move) {
	roster := s.(*Roster)
	mv := m.(nrpMove)
	roster.Set(mv.nurseA, mv.day, mv.oldA)
	roster.Set(mv.nurseB, mv.day, mv.oldB)
}

// SolutionFingerprint returns a 12-char MD5 hash of the roster assignments.
func (p *NRPProblem) SolutionFingerprint(s optimisation.Solution) string {
	roster := s.(*Roster)
	return RosterFingerprint(roster)
}

// SerializeSolution converts the roster to the dashboard roster.json format.
func (p *NRPProblem) SerializeSolution(s optimisation.Solution) ([]byte, error) {
	roster := s.(*Roster)
	sol := RosterToSolution(roster, p.sc, p.hist.Week)
	entries := solutionToRosterEntries(sol, p.sc)
	return json.Marshal(entries)
}

// --- Helper: convert Solution to dashboard RosterEntry format ---

type dashboardRosterEntry struct {
	Week        int      `json:"week"`
	Nurse       string   `json:"nurse"`
	Day         string   `json:"day"`
	DayIndex    int      `json:"dayIndex"`
	ShiftType   string   `json:"shiftType"`
	Skill       string   `json:"skill"`
	Contract    string   `json:"contract"`
	NurseSkills []string `json:"nurseSkills"`
}

func solutionToRosterEntries(sol Solution, sc Scenario) []dashboardRosterEntry {
	nurseMap := make(map[string]Nurse)
	for _, n := range sc.Nurses {
		nurseMap[n.ID] = n
	}

	var entries []dashboardRosterEntry
	for _, a := range sol.Assignments {
		nurse := nurseMap[a.Nurse]
		entries = append(entries, dashboardRosterEntry{
			Week:        sol.Week + 1,
			Nurse:       a.Nurse,
			Day:         a.Day,
			DayIndex:    DayIndex(a.Day),
			ShiftType:   a.ShiftType,
			Skill:       a.Skill,
			Contract:    nurse.Contract,
			NurseSkills: nurse.Skills,
		})
	}
	return entries
}

// --- Accessors for existing code that needs domain-specific data ---

// Scenario returns the INRC-II scenario (for beam search orchestration).
func (p *NRPProblem) Scenario() Scenario { return p.sc }

// WeekData returns the week data (for beam search orchestration).
func (p *NRPProblem) WeekData() WeekData { return p.wd }

// History returns the nurse history (for beam search orchestration).
func (p *NRPProblem) History() History { return p.hist }

// ScoringWorkspace returns the pre-computed scoring workspace.
func (p *NRPProblem) ScoringWorkspace() *ScoringWorkspace { return p.ws }

// NurseSkills returns the pre-computed nurse skills lookup.
func (p *NRPProblem) NurseSkills() []map[string]bool { return p.nurseSkills }

// Forbidden returns the pre-computed forbidden succession set.
func (p *NRPProblem) Forbidden() map[string]bool { return p.forbidden }

// HistLastShift returns the pre-computed history last shift array.
func (p *NRPProblem) HistLastShift() []string { return p.histLastShift }

// --- Compile-time interface check ---

var _ optimisation.Problem = (*NRPProblem)(nil)
