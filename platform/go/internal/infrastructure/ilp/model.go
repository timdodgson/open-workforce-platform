package ilp

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// ModelInfo holds metadata about the generated LP model for diagnostics.
type ModelInfo struct {
	NumVariables           int
	NumConstraints         int
	NumNurses              int
	NumDays                int
	NumShiftTypes          int
	SupportedConstraints   []string
	UnsupportedConstraints []string
}

// BuildModel generates a complete LP-format file for the INRC-II nurse rostering problem.
// Uses transition-based flow formulation with continuous penalty variables for
// min-consecutive constraints (Gemini formulation).
//
// Decision variables:
//   x_n_d_s_k  = 1 if nurse n assigned shift s skill k on day d (binary)
//   w_n_d      = 1 if nurse n works on day d (binary, linked to x)
//   st_n_d     = 1 if working streak starts at day d (binary transition)
//   en_n_d     = 1 if working streak ends at day d (binary transition)
//   sts_n_d_s  = 1 if shift-type streak starts at day d (binary transition)
//   p*         = continuous penalty variables (≥ 0)
func BuildModel(sc inrc2.Scenario, weekDataFiles []string, initialHist inrc2.History,
	weeks int, outputPath string) (ModelInfo, error) {

	if weeks > len(weekDataFiles) {
		weeks = len(weekDataFiles)
	}
	if weeks == 0 {
		return ModelInfo{}, fmt.Errorf("no weeks to solve")
	}

	numNurses := len(sc.Nurses)
	numDays := weeks * 7
	shiftTypes := make([]string, len(sc.ShiftTypes))
	for i, st := range sc.ShiftTypes {
		shiftTypes[i] = st.ID
	}
	skills := sc.Skills

	weekDatas := make([]inrc2.WeekData, weeks)
	for w := 0; w < weeks; w++ {
		wd, err := inrc2.LoadWeekData(weekDataFiles[w])
		if err != nil {
			return ModelInfo{}, fmt.Errorf("failed to load week %d data: %w", w+1, err)
		}
		weekDatas[w] = wd
	}

	nurseSkills := make([]map[string]bool, numNurses)
	for i, n := range sc.Nurses {
		sk := make(map[string]bool)
		for _, s := range n.Skills {
			sk[s] = true
		}
		nurseSkills[i] = sk
	}

	forbidden := make(map[string]bool)
	for _, fs := range sc.ForbiddenShiftTypeSuccessions {
		for _, succ := range fs.SucceedingShiftTypes {
			forbidden[fs.PrecedingShiftType+"|"+succ] = true
		}
	}

	contractMap := make(map[string]inrc2.Contract)
	for _, c := range sc.Contracts {
		contractMap[c.ID] = c
	}
	nurseContracts := make([]inrc2.Contract, numNurses)
	for i, n := range sc.Nurses {
		nurseContracts[i] = contractMap[n.Contract]
	}

	nurseHistMap := make(map[string]inrc2.NurseHistory)
	for _, nh := range initialHist.NurseHistory {
		nurseHistMap[nh.Nurse] = nh
	}
	nurseHists := make([]inrc2.NurseHistory, numNurses)
	for i, n := range sc.Nurses {
		nurseHists[i] = nurseHistMap[n.ID]
	}

	shiftMinConsec := make(map[string]int)
	shiftMaxConsec := make(map[string]int)
	for _, st := range sc.ShiftTypes {
		shiftMinConsec[st.ID] = st.MinimumNumberOfConsecutiveAssignments
		shiftMaxConsec[st.ID] = st.MaximumNumberOfConsecutiveAssignments
	}

	isFinalWeek := (initialHist.Week + weeks) >= sc.NumberOfWeeks

	var b strings.Builder
	var objTerms []string
	conIdx := 0
	numVars := 0

	type slackInfo struct{ name string; ub int }
	var slackVars []slackInfo
	var binaryVars []string
	var contVars []string // continuous penalty variables

	// --- Objective: S1 optimal coverage ---
	for w := 0; w < weeks; w++ {
		wd := weekDatas[w]
		for _, req := range wd.Requirements {
			for d := 0; d < 7; d++ {
				dayReq := req.RequirementForDay(d)
				if dayReq.Optimal > dayReq.Minimum {
					globalDay := w*7 + d
					sv := fmt.Sprintf("s1_%d_%s_%s", globalDay, req.ShiftType, req.Skill)
					objTerms = append(objTerms, fmt.Sprintf("30 %s", sv))
					slackVars = append(slackVars, slackInfo{sv, dayReq.Optimal - dayReq.Minimum})
				}
			}
		}
	}

	// --- Objective: S2 max consecutive working days (30 per unit) ---
	for ni := 0; ni < numNurses; ni++ {
		c := nurseContracts[ni]
		if c.MaximumNumberOfConsecutiveWorkingDays <= 0 {
			continue
		}
		maxW := c.MaximumNumberOfConsecutiveWorkingDays
		for start := 0; start+maxW+1 <= numDays; start++ {
			sv := fmt.Sprintf("s2mx_%d_%d", ni, start)
			objTerms = append(objTerms, fmt.Sprintf("30 %s", sv))
			contVars = append(contVars, sv)
		}
	}

	// --- Objective: S2 min consecutive working days (30 per unit) ---
	for ni := 0; ni < numNurses; ni++ {
		c := nurseContracts[ni]
		if c.MinimumNumberOfConsecutiveWorkingDays <= 1 {
			continue
		}
		for d := 0; d+c.MinimumNumberOfConsecutiveWorkingDays-1 < numDays; d++ {
			sv := fmt.Sprintf("s2mn_%d_%d", ni, d)
			objTerms = append(objTerms, fmt.Sprintf("30 %s", sv))
			contVars = append(contVars, sv)
		}
	}

	// --- Objective: S3 max consecutive days off (30 per unit) ---
	for ni := 0; ni < numNurses; ni++ {
		c := nurseContracts[ni]
		if c.MaximumNumberOfConsecutiveDaysOff <= 0 {
			continue
		}
		maxO := c.MaximumNumberOfConsecutiveDaysOff
		for start := 0; start+maxO+1 <= numDays; start++ {
			sv := fmt.Sprintf("s3mx_%d_%d", ni, start)
			objTerms = append(objTerms, fmt.Sprintf("30 %s", sv))
			contVars = append(contVars, sv)
		}
	}

	// --- Objective: S3 min consecutive days off (30 per unit) ---
	for ni := 0; ni < numNurses; ni++ {
		c := nurseContracts[ni]
		if c.MinimumNumberOfConsecutiveDaysOff <= 1 {
			continue
		}
		for d := 0; d+c.MinimumNumberOfConsecutiveDaysOff-1 < numDays; d++ {
			sv := fmt.Sprintf("s3mn_%d_%d", ni, d)
			objTerms = append(objTerms, fmt.Sprintf("30 %s", sv))
			contVars = append(contVars, sv)
		}
	}

	// --- Objective: S4 max consecutive shift type (15 per unit) ---
	for ni := 0; ni < numNurses; ni++ {
		for _, s := range shiftTypes {
			maxC := shiftMaxConsec[s]
			if maxC <= 0 { continue }
			for start := 0; start+maxC+1 <= numDays; start++ {
				sv := fmt.Sprintf("s4mx_%d_%d_%s", ni, start, s)
				objTerms = append(objTerms, fmt.Sprintf("15 %s", sv))
				contVars = append(contVars, sv)
			}
		}
	}

	// --- Objective: S4 min consecutive shift type (15 per unit) ---
	for ni := 0; ni < numNurses; ni++ {
		for _, s := range shiftTypes {
			minC := shiftMinConsec[s]
			if minC <= 1 { continue }
			for d := 0; d+minC-1 < numDays; d++ {
				sv := fmt.Sprintf("s4mn_%d_%d_%s", ni, d, s)
				objTerms = append(objTerms, fmt.Sprintf("15 %s", sv))
				contVars = append(contVars, sv)
			}
		}
	}

	// --- Objective: S5 shift-off requests (10 per violation) ---
	for w := 0; w < weeks; w++ {
		for ri := range weekDatas[w].ShiftOffRequests {
			sv := fmt.Sprintf("s5_%d_%d", w, ri)
			objTerms = append(objTerms, fmt.Sprintf("10 %s", sv))
			binaryVars = append(binaryVars, sv)
		}
	}

	// --- Objective: S6 complete weekends (30 per violation) ---
	for ni := 0; ni < numNurses; ni++ {
		if nurseContracts[ni].CompleteWeekends != 1 { continue }
		for w := 0; w < weeks; w++ {
			sv := fmt.Sprintf("s6_%d_%d", ni, w)
			objTerms = append(objTerms, fmt.Sprintf("30 %s", sv))
			binaryVars = append(binaryVars, sv)
		}
	}

	// --- Objective: S7 total assignments (20 per unit, final week only) ---
	if isFinalWeek {
		for ni := 0; ni < numNurses; ni++ {
			c := nurseContracts[ni]
			if c.MinimumNumberOfAssignments > 0 {
				sv := fmt.Sprintf("s7u_%d", ni)
				objTerms = append(objTerms, fmt.Sprintf("20 %s", sv))
				contVars = append(contVars, sv)
			}
			if c.MaximumNumberOfAssignments > 0 {
				sv := fmt.Sprintf("s7o_%d", ni)
				objTerms = append(objTerms, fmt.Sprintf("20 %s", sv))
				contVars = append(contVars, sv)
			}
		}
	}

	// --- Objective: S8 total working weekends (30 per unit, final week only) ---
	if isFinalWeek {
		for ni := 0; ni < numNurses; ni++ {
			c := nurseContracts[ni]
			if c.MaximumNumberOfWorkingWeekends > 0 {
				sv := fmt.Sprintf("s8_%d", ni)
				objTerms = append(objTerms, fmt.Sprintf("30 %s", sv))
				contVars = append(contVars, sv)
			}
		}
	}

	// --- Write LP header ---
	b.WriteString("\\* INRC-II ILP (transition-based) *\\\n")
	b.WriteString(fmt.Sprintf("\\* %s: %d nurses, %d days, %d shifts *\\\n",
		sc.ID, numNurses, numDays, len(shiftTypes)))
	b.WriteString("\nMinimize\n obj: ")
	for i, term := range objTerms {
		if i > 0 { b.WriteString(" + ") }
		if i > 0 && i%8 == 0 { b.WriteString("\n   ") }
		b.WriteString(term)
	}
	b.WriteString("\n\nSubject To\n")

	// === LINKING: w[n][d] = sum of x[n][d][*][*] ===
	for ni := 0; ni < numNurses; ni++ {
		for d := 0; d < numDays; d++ {
			var terms []string
			for _, s := range shiftTypes {
				for _, k := range skills {
					if nurseSkills[ni][k] {
						terms = append(terms, xVar(ni, d, s, k))
					}
				}
			}
			conIdx++
			b.WriteString(fmt.Sprintf(" lw_%d_%d: %s - %s = 0\n",
				ni, d, strings.Join(terms, " + "), wVar(ni, d)))
		}
	}

	// === H1: w[n][d] <= 1 (one shift per day) ===
	for ni := 0; ni < numNurses; ni++ {
		for d := 0; d < numDays; d++ {
			conIdx++
			b.WriteString(fmt.Sprintf(" h1_%d_%d: %s <= 1\n", ni, d, wVar(ni, d)))
		}
	}

	// === H3: Minimum coverage ===
	for w := 0; w < weeks; w++ {
		wd := weekDatas[w]
		for _, req := range wd.Requirements {
			for d := 0; d < 7; d++ {
				dayReq := req.RequirementForDay(d)
				if dayReq.Minimum <= 0 { continue }
				globalDay := w*7 + d
				var terms []string
				for ni := 0; ni < numNurses; ni++ {
					if nurseSkills[ni][req.Skill] {
						terms = append(terms, xVar(ni, globalDay, req.ShiftType, req.Skill))
					}
				}
				if len(terms) > 0 {
					conIdx++
					b.WriteString(fmt.Sprintf(" h3_%d_%s_%s: %s >= %d\n",
						globalDay, req.ShiftType, req.Skill,
						strings.Join(terms, " + "), dayReq.Minimum))
				}
			}
		}
	}

	// === H4: Forbidden succession from history ===
	for ni := 0; ni < numNurses; ni++ {
		nh := nurseHists[ni]
		if nh.LastAssignedShiftType == "" || nh.LastAssignedShiftType == "None" || nh.NumberOfConsecutiveDaysOff > 0 {
			continue
		}
		for _, s2 := range shiftTypes {
			if !forbidden[nh.LastAssignedShiftType+"|"+s2] { continue }
			for _, k := range skills {
				if nurseSkills[ni][k] {
					conIdx++
					b.WriteString(fmt.Sprintf(" h4h_%d_%s_%s: %s <= 0\n",
						ni, s2, k, xVar(ni, 0, s2, k)))
				}
			}
		}
	}

	// === H4: Forbidden succession within horizon ===
	for ni := 0; ni < numNurses; ni++ {
		for d := 0; d < numDays-1; d++ {
			for _, s1 := range shiftTypes {
				for _, s2 := range shiftTypes {
					if !forbidden[s1+"|"+s2] { continue }
					var terms []string
					for _, k := range skills {
						if nurseSkills[ni][k] {
							terms = append(terms, xVar(ni, d, s1, k))
							terms = append(terms, xVar(ni, d+1, s2, k))
						}
					}
					if len(terms) > 0 {
						conIdx++
						b.WriteString(fmt.Sprintf(" h4_%d_%d_%s_%s: %s <= 1\n",
							ni, d, s1, s2, strings.Join(terms, " + ")))
					}
				}
			}
		}
	}

	// === TRANSITION DETECTION ===
	// st[n][d] = w[d] AND NOT w[d-1] (exact start detection)
	// en[n][d] = w[d] AND NOT w[d+1] (exact end detection)
	// Exact bounds: st[d] >= w[d]-w[d-1], st[d] <= w[d], st[d] <= 1-w[d-1]
	// These 3 constraints together force st[d] = max(0, w[d]-w[d-1]).
	for ni := 0; ni < numNurses; ni++ {
		for d := 0; d < numDays; d++ {
			// --- Start ---
			if d == 0 {
				wasOff := nurseHists[ni].NumberOfConsecutiveDaysOff > 0 ||
					nurseHists[ni].LastAssignedShiftType == "None" ||
					nurseHists[ni].LastAssignedShiftType == ""
				if wasOff {
					// st[0] = w[0]
					conIdx++
					b.WriteString(fmt.Sprintf(" stl_%d_%d: %s - %s >= 0\n", ni, d, stVar(ni, d), wVar(ni, d)))
					conIdx++
					b.WriteString(fmt.Sprintf(" stu_%d_%d: %s - %s <= 0\n", ni, d, stVar(ni, d), wVar(ni, d)))
				} else {
					// Was working: st[0] = 0
					conIdx++
					b.WriteString(fmt.Sprintf(" stf_%d_%d: %s = 0\n", ni, d, stVar(ni, d)))
				}
			} else {
				conIdx++
				b.WriteString(fmt.Sprintf(" stl_%d_%d: %s - %s + %s >= 0\n",
					ni, d, stVar(ni, d), wVar(ni, d), wVar(ni, d-1)))
				conIdx++
				b.WriteString(fmt.Sprintf(" stu1_%d_%d: %s - %s <= 0\n",
					ni, d, stVar(ni, d), wVar(ni, d)))
				conIdx++
				b.WriteString(fmt.Sprintf(" stu2_%d_%d: %s + %s <= 1\n",
					ni, d, stVar(ni, d), wVar(ni, d-1)))
			}

			// --- End ---
			if d == numDays-1 {
				// en[last] = w[last] (if working on last day, streak ends)
				conIdx++
				b.WriteString(fmt.Sprintf(" enl_%d_%d: %s - %s >= 0\n", ni, d, enVar(ni, d), wVar(ni, d)))
				conIdx++
				b.WriteString(fmt.Sprintf(" enu_%d_%d: %s - %s <= 0\n", ni, d, enVar(ni, d), wVar(ni, d)))
			} else {
				conIdx++
				b.WriteString(fmt.Sprintf(" enl_%d_%d: %s - %s + %s >= 0\n",
					ni, d, enVar(ni, d), wVar(ni, d), wVar(ni, d+1)))
				conIdx++
				b.WriteString(fmt.Sprintf(" enu1_%d_%d: %s - %s <= 0\n",
					ni, d, enVar(ni, d), wVar(ni, d)))
				conIdx++
				b.WriteString(fmt.Sprintf(" enu2_%d_%d: %s + %s <= 1\n",
					ni, d, enVar(ni, d), wVar(ni, d+1)))
			}
		}
	}

	// Shift-type start: lower bound only. The forward window constraint (S4 MIN)
	// structurally forces sts=1 when a genuine transition occurs, because:
	//   sts >= y[d] - y[d-1] → sts >= 1 when transition happens
	//   sts is bounded [0,1] → sts = 1
	//   Window: sum(y[d..d+P-1]) + p >= P * sts → forces p > 0 for short streaks
	// When no transition: sts >= 0, and minimizing p drives sts to 0 naturally.
	for ni := 0; ni < numNurses; ni++ {
		for d := 0; d < numDays; d++ {
			for _, s := range shiftTypes {
				var termsD, termsPrev []string
				for _, k := range skills {
					if nurseSkills[ni][k] {
						termsD = append(termsD, xVar(ni, d, s, k))
						if d > 0 {
							termsPrev = append(termsPrev, xVar(ni, d-1, s, k))
						}
					}
				}
				conIdx++
				if d == 0 {
					wasOnShift := nurseHists[ni].LastAssignedShiftType == s &&
						nurseHists[ni].NumberOfConsecutiveDaysOff == 0
					if wasOnShift {
						b.WriteString(fmt.Sprintf(" stsl_%d_%d_%s: %s >= 0\n",
							ni, d, s, stsVar(ni, d, s)))
					} else {
						b.WriteString(fmt.Sprintf(" stsl_%d_%d_%s: %s - %s >= 0\n",
							ni, d, s, stsVar(ni, d, s), strings.Join(termsD, " + ")))
					}
				} else {
					b.WriteString(fmt.Sprintf(" stsl_%d_%d_%s: %s - %s + %s >= 0\n",
						ni, d, s, stsVar(ni, d, s),
						strings.Join(termsD, " + "),
						strings.Join(termsPrev, " + ")))
				}
			}
		}
	}

	// === S2 MAX: sliding window sum(w[start..start+max]) - max <= penalty ===
	for ni := 0; ni < numNurses; ni++ {
		c := nurseContracts[ni]
		maxW := c.MaximumNumberOfConsecutiveWorkingDays
		if maxW <= 0 { continue }
		for start := 0; start+maxW+1 <= numDays; start++ {
			sv := fmt.Sprintf("s2mx_%d_%d", ni, start)
			var terms []string
			for d := start; d < start+maxW+1; d++ {
				terms = append(terms, wVar(ni, d))
			}
			conIdx++
			b.WriteString(fmt.Sprintf(" s2mxc_%d_%d: %s - %s <= %d\n",
				ni, start, strings.Join(terms, " + "), sv, maxW))
		}
	}

	// === S2 MIN: sum(w[d..d+M-1]) >= M * st[d] - penalty ===
	for ni := 0; ni < numNurses; ni++ {
		c := nurseContracts[ni]
		minW := c.MinimumNumberOfConsecutiveWorkingDays
		if minW <= 1 { continue }
		for d := 0; d+minW-1 < numDays; d++ {
			sv := fmt.Sprintf("s2mn_%d_%d", ni, d)
			var terms []string
			for j := 0; j < minW && d+j < numDays; j++ {
				terms = append(terms, wVar(ni, d+j))
			}
			conIdx++
			b.WriteString(fmt.Sprintf(" s2mnc_%d_%d: %s + %s - %d %s >= 0\n",
				ni, d, strings.Join(terms, " + "), sv, minW, stVar(ni, d)))
		}
	}
	// === S3 MIN: sum(w[d..d+N-1]) + N*en[d-1] - p <= N ===
	for ni := 0; ni < numNurses; ni++ {
		c := nurseContracts[ni]
		minO := c.MinimumNumberOfConsecutiveDaysOff
		if minO <= 1 { continue }
		for d := 1; d+minO-1 < numDays; d++ {
			sv := fmt.Sprintf("s3mn_%d_%d", ni, d)
			var terms []string
			for j := 0; j < minO; j++ {
				terms = append(terms, wVar(ni, d+j))
			}
			conIdx++
			b.WriteString(fmt.Sprintf(" s3mnc_%d_%d: %s + %d %s - %s <= %d\n",
				ni, d, strings.Join(terms, " + "), minO, enVar(ni, d-1), sv, minO))
		}
	}
	// === S4 MIN: sum(y[d..d+P-1][s]) >= P * sts[d][s] - penalty ===
	// (sts uses lower-bound only, so this underestimates S4 min penalty)
	for ni := 0; ni < numNurses; ni++ {
		for _, s := range shiftTypes {
			minC := shiftMinConsec[s]
			if minC <= 1 { continue }
			for d := 0; d+minC-1 < numDays; d++ {
				sv := fmt.Sprintf("s4mn_%d_%d_%s", ni, d, s)
				var terms []string
				for j := 0; j < minC && d+j < numDays; j++ {
					for _, k := range skills {
						if nurseSkills[ni][k] {
							terms = append(terms, xVar(ni, d+j, s, k))
						}
					}
				}
				conIdx++
				b.WriteString(fmt.Sprintf(" s4mnc_%d_%d_%s: %s + %s - %d %s >= 0\n",
					ni, d, s, strings.Join(terms, " + "), sv, minC, stsVar(ni, d, s)))
			}
		}
	}

	// === S3 MAX: sliding window (maxOff+1 - sum(w)) <= maxOff + penalty ===
	// Rewritten: sum(w[window]) + penalty >= 1
	for ni := 0; ni < numNurses; ni++ {
		c := nurseContracts[ni]
		maxO := c.MaximumNumberOfConsecutiveDaysOff
		if maxO <= 0 { continue }
		for start := 0; start+maxO+1 <= numDays; start++ {
			sv := fmt.Sprintf("s3mx_%d_%d", ni, start)
			var terms []string
			for d := start; d < start+maxO+1; d++ {
				terms = append(terms, wVar(ni, d))
			}
			conIdx++
			// If all days are off, sum=0, penalty forced to 1.
			b.WriteString(fmt.Sprintf(" s3mxc_%d_%d: %s + %s >= 1\n",
				ni, start, strings.Join(terms, " + "), sv))
		}
	}

	// (S3 MIN constraints are in the block above with S2 MIN and S4 MIN)

	// === S4 MAX: sliding window for shift type ===
	for ni := 0; ni < numNurses; ni++ {
		for _, s := range shiftTypes {
			maxC := shiftMaxConsec[s]
			if maxC <= 0 { continue }
			for start := 0; start+maxC+1 <= numDays; start++ {
				sv := fmt.Sprintf("s4mx_%d_%d_%s", ni, start, s)
				var terms []string
				for d := start; d < start+maxC+1; d++ {
					for _, k := range skills {
						if nurseSkills[ni][k] {
							terms = append(terms, xVar(ni, d, s, k))
						}
					}
				}
				conIdx++
				b.WriteString(fmt.Sprintf(" s4mxc_%d_%d_%s: %s - %s <= %d\n",
					ni, start, s, strings.Join(terms, " + "), sv, maxC))
			}
		}
	}

	// === S4 MIN: DISABLED FOR DEBUG ===

	// === S1: Optimal coverage slack ===
	for w := 0; w < weeks; w++ {
		wd := weekDatas[w]
		for _, req := range wd.Requirements {
			for d := 0; d < 7; d++ {
				dayReq := req.RequirementForDay(d)
				if dayReq.Optimal <= dayReq.Minimum { continue }
				globalDay := w*7 + d
				sv := fmt.Sprintf("s1_%d_%s_%s", globalDay, req.ShiftType, req.Skill)
				var terms []string
				for ni := 0; ni < numNurses; ni++ {
					if nurseSkills[ni][req.Skill] {
						terms = append(terms, xVar(ni, globalDay, req.ShiftType, req.Skill))
					}
				}
				if len(terms) > 0 {
					conIdx++
					b.WriteString(fmt.Sprintf(" s1c_%d_%s_%s: %s + %s >= %d\n",
						globalDay, req.ShiftType, req.Skill,
						strings.Join(terms, " + "), sv, dayReq.Optimal))
				}
			}
		}
	}

	// === S5: Shift-off requests ===
	for w := 0; w < weeks; w++ {
		wd := weekDatas[w]
		for ri, req := range wd.ShiftOffRequests {
			dayIdx := inrc2.DayIndex(req.Day)
			if dayIdx < 0 { continue }
			globalDay := w*7 + dayIdx
			sv := fmt.Sprintf("s5_%d_%d", w, ri)
			nurseIdx := -1
			for ni, n := range sc.Nurses {
				if n.ID == req.Nurse { nurseIdx = ni; break }
			}
			if nurseIdx < 0 { continue }
			if req.ShiftType == "Any" {
				conIdx++
				b.WriteString(fmt.Sprintf(" s5c_%d_%d: %s - %s <= 0\n",
					w, ri, wVar(nurseIdx, globalDay), sv))
			} else {
				for _, k := range skills {
					if nurseSkills[nurseIdx][k] {
						conIdx++
						b.WriteString(fmt.Sprintf(" s5c_%d_%d_%s: %s - %s <= 0\n",
							w, ri, k, xVar(nurseIdx, globalDay, req.ShiftType, k), sv))
					}
				}
			}
		}
	}

	// === S6: Complete weekends ===
	for ni := 0; ni < numNurses; ni++ {
		if nurseContracts[ni].CompleteWeekends != 1 { continue }
		for w := 0; w < weeks; w++ {
			sv := fmt.Sprintf("s6_%d_%d", ni, w)
			satDay := w*7 + 5
			sunDay := w*7 + 6
			conIdx++
			b.WriteString(fmt.Sprintf(" s6a_%d_%d: %s - %s - %s <= 0\n",
				ni, w, wVar(ni, satDay), wVar(ni, sunDay), sv))
			conIdx++
			b.WriteString(fmt.Sprintf(" s6b_%d_%d: %s - %s - %s <= 0\n",
				ni, w, wVar(ni, sunDay), wVar(ni, satDay), sv))
		}
	}

	// === S7: Total assignments (final week only) ===
	if isFinalWeek {
		for ni := 0; ni < numNurses; ni++ {
			c := nurseContracts[ni]
			histAssign := nurseHists[ni].NumberOfAssignments
			var allW []string
			for d := 0; d < numDays; d++ {
				allW = append(allW, wVar(ni, d))
			}
			totalExpr := strings.Join(allW, " + ")
			if c.MinimumNumberOfAssignments > 0 {
				sv := fmt.Sprintf("s7u_%d", ni)
				rhs := c.MinimumNumberOfAssignments - histAssign
				if rhs > 0 {
					conIdx++
					b.WriteString(fmt.Sprintf(" s7uc_%d: %s + %s >= %d\n",
						ni, totalExpr, sv, rhs))
				}
			}
			if c.MaximumNumberOfAssignments > 0 {
				sv := fmt.Sprintf("s7o_%d", ni)
				rhs := c.MaximumNumberOfAssignments - histAssign
				conIdx++
				b.WriteString(fmt.Sprintf(" s7oc_%d: %s - %s <= %d\n",
					ni, totalExpr, sv, rhs))
			}
		}
	}

	// === S8: Total working weekends (final week only) ===
	if isFinalWeek {
		for ni := 0; ni < numNurses; ni++ {
			c := nurseContracts[ni]
			if c.MaximumNumberOfWorkingWeekends <= 0 { continue }
			histWk := nurseHists[ni].NumberOfWorkingWeekends
			sv := fmt.Sprintf("s8_%d", ni)
			var wkVars []string
			for w := 0; w < weeks; w++ {
				wkv := fmt.Sprintf("wkend_%d_%d", ni, w)
				wkVars = append(wkVars, wkv)
				binaryVars = append(binaryVars, wkv)
				satDay := w*7 + 5
				sunDay := w*7 + 6
				conIdx++
				b.WriteString(fmt.Sprintf(" wk8a_%d_%d: %s - %s >= 0\n", ni, w, wkv, wVar(ni, satDay)))
				conIdx++
				b.WriteString(fmt.Sprintf(" wk8b_%d_%d: %s - %s >= 0\n", ni, w, wkv, wVar(ni, sunDay)))
				conIdx++
				b.WriteString(fmt.Sprintf(" wk8c_%d_%d: %s - %s - %s <= 0\n",
					ni, w, wkv, wVar(ni, satDay), wVar(ni, sunDay)))
			}
			rhs := c.MaximumNumberOfWorkingWeekends - histWk
			conIdx++
			b.WriteString(fmt.Sprintf(" s8c_%d: %s - %s <= %d\n",
				ni, strings.Join(wkVars, " + "), sv, rhs))
		}
	}

	// === TERMINAL S4: Penalise incomplete shift-type streaks at end of horizon ===
	// Constraint: sum(y[D-P+1..D-1]) + p >= (P - 1 - histStreak) * y[D]
	// No duplicate variables: day D only appears on RHS as trigger.
	for ni := 0; ni < numNurses; ni++ {
		for _, s := range shiftTypes {
			minC := shiftMinConsec[s]
			if minC <= 1 {
				continue
			}
			histStreak := 0
			if nurseHists[ni].LastAssignedShiftType == s && nurseHists[ni].NumberOfConsecutiveDaysOff == 0 {
				histStreak = nurseHists[ni].NumberOfConsecutiveAssignments
			}

			rhs := minC - 1 - histStreak
			if rhs <= 0 {
				continue // Already met from history, no penalty possible.
			}

			sv := fmt.Sprintf("ts4_%d_%s", ni, s)
			contVars = append(contVars, sv)
			objTerms = append(objTerms, fmt.Sprintf("15 %s", sv))

			lastDay := numDays - 1
			startDay := lastDay - minC + 1
			if startDay < 0 {
				startDay = 0
			}

			// Window sum: days startDay to lastDay-1 (EXCLUDING lastDay).
			var windowTerms []string
			for d := startDay; d < lastDay; d++ {
				for _, k := range skills {
					if nurseSkills[ni][k] {
						windowTerms = append(windowTerms, xVar(ni, d, s, k))
					}
				}
			}

			// RHS trigger: (minC-1-histStreak) * y[lastDay]
			var triggerTerms []string
			for _, k := range skills {
				if nurseSkills[ni][k] {
					triggerTerms = append(triggerTerms, fmt.Sprintf("%d %s", rhs, xVar(ni, lastDay, s, k)))
				}
			}

			conIdx++
			lhs := strings.Join(windowTerms, " + ")
			if lhs == "" {
				lhs = "0"
			}
			b.WriteString(fmt.Sprintf(" ts4c_%d_%s: %s + %s - %s >= 0\n",
				ni, s, lhs, sv, strings.Join(triggerTerms, " - ")))
		}
	}

	// Same terminal penalty at each internal week boundary (for rolling horizon awareness).
	if weeks > 1 {
		for ni := 0; ni < numNurses; ni++ {
			for _, s := range shiftTypes {
				minC := shiftMinConsec[s]
				if minC <= 1 {
					continue
				}
				for wk := 0; wk < weeks-1; wk++ {
					wkLastDay := (wk+1)*7 - 1
					wkStartDay := wkLastDay - minC + 1
					if wkStartDay < wk*7 {
						wkStartDay = wk * 7
					}

					rhs := minC - 1 // No history at internal boundaries.
					svW := fmt.Sprintf("ts4_%d_%d_%s", ni, wk, s)
					contVars = append(contVars, svW)
					objTerms = append(objTerms, fmt.Sprintf("15 %s", svW))

					var windowTerms []string
					for d := wkStartDay; d < wkLastDay; d++ {
						for _, k := range skills {
							if nurseSkills[ni][k] {
								windowTerms = append(windowTerms, xVar(ni, d, s, k))
							}
						}
					}

					var triggerTerms []string
					for _, k := range skills {
						if nurseSkills[ni][k] {
							triggerTerms = append(triggerTerms, fmt.Sprintf("%d %s", rhs, xVar(ni, wkLastDay, s, k)))
						}
					}

					conIdx++
					lhs := strings.Join(windowTerms, " + ")
					if lhs == "" {
						lhs = "0"
					}
					b.WriteString(fmt.Sprintf(" ts4wc_%d_%d_%s: %s + %s - %s >= 0\n",
						ni, wk, s, lhs, svW, strings.Join(triggerTerms, " - ")))
				}
			}
		}
	}

	// === BOUNDS ===
	b.WriteString("\nBounds\n")
	for _, sv := range slackVars {
		b.WriteString(fmt.Sprintf(" 0 <= %s <= %d\n", sv.name, sv.ub))
	}
	for _, sv := range contVars {
		b.WriteString(fmt.Sprintf(" 0 <= %s\n", sv))
	}
	for _, bv := range binaryVars {
		b.WriteString(fmt.Sprintf(" 0 <= %s <= 1\n", bv))
	}
	// sts continuous [0,1]
	for ni := 0; ni < numNurses; ni++ {
		for d := 0; d < numDays; d++ {
			for _, s := range shiftTypes {
				b.WriteString(fmt.Sprintf(" 0 <= %s <= 1\n", stsVar(ni, d, s)))
				numVars++
			}
		}
	}

	// === BINARY ===
	b.WriteString("\nBinary\n")
	// x variables
	for ni := 0; ni < numNurses; ni++ {
		for d := 0; d < numDays; d++ {
			for _, s := range shiftTypes {
				for _, k := range skills {
					if nurseSkills[ni][k] {
						b.WriteString(fmt.Sprintf(" %s\n", xVar(ni, d, s, k)))
						numVars++
					}
				}
			}
		}
	}
	// w variables
	for ni := 0; ni < numNurses; ni++ {
		for d := 0; d < numDays; d++ {
			b.WriteString(fmt.Sprintf(" %s\n", wVar(ni, d)))
			numVars++
		}
	}
	// st, en transition variables
	for ni := 0; ni < numNurses; ni++ {
		for d := 0; d < numDays; d++ {
			b.WriteString(fmt.Sprintf(" %s\n", stVar(ni, d)))
			b.WriteString(fmt.Sprintf(" %s\n", enVar(ni, d)))
			numVars += 2
		}
	}
	// sts (shift-type start) variables — continuous [0,1] (not binary)
	// Structural forcing via window constraint makes them behave as binary.
	// Keeping them continuous massively reduces branch-and-bound tree.
	// (Added to Bounds section instead.)

	// slack vars (binary bounded)
	for _, sv := range slackVars {
		b.WriteString(fmt.Sprintf(" %s\n", sv.name))
		numVars++
	}
	// other binary vars
	for _, bv := range binaryVars {
		b.WriteString(fmt.Sprintf(" %s\n", bv))
		numVars++
	}

	// Continuous penalty variables (General integer section not needed — they're continuous).
	// HiGHS LP format: continuous vars don't need to be in Binary or General.
	for _, sv := range contVars {
		numVars++
		_ = sv
	}

	b.WriteString("\nEnd\n")

	if err := os.WriteFile(outputPath, []byte(b.String()), 0644); err != nil {
		return ModelInfo{}, err
	}

	return ModelInfo{
		NumVariables:   numVars,
		NumConstraints: conIdx,
		NumNurses:      numNurses,
		NumDays:        numDays,
		NumShiftTypes:  len(shiftTypes),
		SupportedConstraints: []string{
			"H1", "H2", "H3", "H4",
			"S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8",
		},
		UnsupportedConstraints: []string{},
	}, nil
}

func xVar(ni, d int, s, k string) string { return fmt.Sprintf("x_%d_%d_%s_%s", ni, d, s, k) }
func wVar(ni, d int) string              { return fmt.Sprintf("w_%d_%d", ni, d) }
func stVar(ni, d int) string             { return fmt.Sprintf("st_%d_%d", ni, d) }
func enVar(ni, d int) string             { return fmt.Sprintf("en_%d_%d", ni, d) }
func stsVar(ni, d int, s string) string  { return fmt.Sprintf("sts_%d_%d_%s", ni, d, s) }
