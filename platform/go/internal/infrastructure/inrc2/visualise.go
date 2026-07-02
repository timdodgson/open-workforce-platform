package inrc2

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// --- PFRS Visualisation Support ---
// Reads audit CSV files and generates a self-contained HTML research dashboard.
// Independent of the optimiser — works on historical data only.

// AuditCSVRecord holds one parsed row from the audit CSV.
type AuditCSVRecord struct {
	Instance             string
	Seed                 int64
	Mode                 string
	IterationsPerWorker  int
	MaxTotalWorkers      int
	MaxConcurrent        int
	InitialTemperature   float64
	CoolingRate          float64
	CoolingMode          string
	EffectiveCoolingRate float64
	MinTemperature       float64
	LateAcceptanceLen    int
	Week                 int
	StartPenalty         int
	FinalPenalty         int
	Improvement          int
	HardViolations       int
	SoftViolations       int
	Candidates           int
	Accepted             int
	Rejected             int
	AcceptanceRate       float64
	BestIteration        int
	BestWorkerID         int
	WorkersStarted       int
	BranchesCreated      int
	BranchesDropped      int
	MaxQueueDepth        int
	MaxConcurrentSeen    int
	DurationMs           int64
	SAFinalTemp          float64
	SATempAtBest         float64
	SAAcceptedBetter     int
	SAAcceptedWorse      int
	SARejectedByProb     int
	LAHCAcceptedByCurrent int
	LAHCAcceptedByLate   int
	LAHCRejectedByLate   int
	BranchesQueued       int
	BranchesStartedCSV   int
	BranchesCompleted    int
	WinningBranchDepth   int
	WorkersImproved      int
	WorkersProducedBest  int
	RejectedNoop         int
	RejectedSkill        int
	RejectedSuccession   int
	RejectedHistory      int
}

// ReadAuditCSV parses an audit CSV file into records.
func ReadAuditCSV(path string) ([]AuditCSVRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var records []AuditCSVRecord
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if lineNum == 1 {
			continue // skip header
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 20 {
			continue
		}
		r := parseAuditCSVFields(fields)
		records = append(records, r)
	}
	return records, scanner.Err()
}

func parseAuditCSVFields(fields []string) AuditCSVRecord {
	r := AuditCSVRecord{}
	r.Instance = fields[0]
	r.Seed, _ = strconv.ParseInt(fields[1], 10, 64)
	r.Mode = fields[2]
	r.IterationsPerWorker, _ = strconv.Atoi(fields[3])
	r.MaxTotalWorkers, _ = strconv.Atoi(fields[4])
	r.MaxConcurrent, _ = strconv.Atoi(fields[5])
	r.InitialTemperature, _ = strconv.ParseFloat(fields[6], 64)
	r.CoolingRate, _ = strconv.ParseFloat(fields[7], 64)
	idx := 8
	if len(fields) > idx { r.CoolingMode = fields[idx]; idx++ }
	if len(fields) > idx { r.EffectiveCoolingRate, _ = strconv.ParseFloat(fields[idx], 64); idx++ }
	if len(fields) > idx { r.MinTemperature, _ = strconv.ParseFloat(fields[idx], 64); idx++ }
	if len(fields) > idx { r.LateAcceptanceLen, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.Week, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.StartPenalty, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.FinalPenalty, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.Improvement, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.HardViolations, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.SoftViolations, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.Candidates, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.Accepted, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.Rejected, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.AcceptanceRate, _ = strconv.ParseFloat(fields[idx], 64); idx++ }
	if len(fields) > idx { r.BestIteration, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.BestWorkerID, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.WorkersStarted, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.BranchesCreated, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.BranchesDropped, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.MaxQueueDepth, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.MaxConcurrentSeen, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.DurationMs, _ = strconv.ParseInt(fields[idx], 10, 64); idx++ }
	if len(fields) > idx { r.SAFinalTemp, _ = strconv.ParseFloat(fields[idx], 64); idx++ }
	if len(fields) > idx { r.SATempAtBest, _ = strconv.ParseFloat(fields[idx], 64); idx++ }
	if len(fields) > idx { r.SAAcceptedBetter, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.SAAcceptedWorse, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.SARejectedByProb, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.LAHCAcceptedByCurrent, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.LAHCAcceptedByLate, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.LAHCRejectedByLate, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.BranchesQueued, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.BranchesStartedCSV, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.BranchesCompleted, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.WinningBranchDepth, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.WorkersImproved, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.WorkersProducedBest, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.RejectedNoop, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.RejectedSkill, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.RejectedSuccession, _ = strconv.Atoi(fields[idx]); idx++ }
	if len(fields) > idx { r.RejectedHistory, _ = strconv.Atoi(fields[idx]); idx++ }
	return r
}

// reportData holds pre-computed aggregates for all dashboard tabs.
type reportData struct {
	Records         []AuditCSVRecord
	First           AuditCSVRecord
	TotalPenalty    int
	TotalCandidates int
	TotalAccepted   int
	TotalRejected   int
	TotalSABetter   int
	TotalSAWorse    int
	TotalSARejProb  int
	TotalWorkers    int
	TotalBranches   int
	TotalDuration   int64
	MaxWeekPenalty  int
	MaxWeekNum      int
	HardRejectRate  float64
	AcceptWorseRate float64
	CumPenalties    []int
	NumWeeks        int
}

func computeReportData(records []AuditCSVRecord) reportData {
	d := reportData{Records: records, First: records[0], NumWeeks: len(records)}
	cum := 0
	for _, r := range records {
		d.TotalPenalty += r.FinalPenalty
		d.TotalCandidates += r.Candidates
		d.TotalAccepted += r.Accepted
		d.TotalRejected += r.Rejected
		d.TotalSABetter += r.SAAcceptedBetter
		d.TotalSAWorse += r.SAAcceptedWorse
		d.TotalSARejProb += r.SARejectedByProb
		d.TotalWorkers += r.WorkersStarted
		d.TotalBranches += r.BranchesCreated
		d.TotalDuration += r.DurationMs
		if r.FinalPenalty > d.MaxWeekPenalty {
			d.MaxWeekPenalty = r.FinalPenalty
			d.MaxWeekNum = r.Week
		}
		cum += r.FinalPenalty
		d.CumPenalties = append(d.CumPenalties, cum)
	}
	if d.TotalCandidates+d.TotalRejected > 0 {
		d.HardRejectRate = float64(d.TotalRejected) / float64(d.TotalCandidates+d.TotalRejected) * 100
	}
	if d.TotalCandidates > 0 {
		d.AcceptWorseRate = float64(d.TotalSAWorse) / float64(d.TotalCandidates) * 100
	}
	return d
}

// GenerateReport generates a self-contained HTML research dashboard.
func GenerateReport(records []AuditCSVRecord, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	if len(records) == 0 {
		return fmt.Errorf("no records to visualise")
	}
	d := computeReportData(records)
	var b strings.Builder
	writeHTMLHead(&b)
	writeLayout(&b, d)
	writeHTMLFooter(&b)
	return os.WriteFile(outputDir+"/summary.html", []byte(b.String()), 0644)
}

// --- HTML Layout ---

func writeHTMLHead(b *strings.Builder) {
	b.WriteString(`<!DOCTYPE html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>PFRS Research Dashboard</title>
`)
	writeCSS(b)
	b.WriteString(`</head><body>`)
}

func writeCSS(b *strings.Builder) {
	b.WriteString(`<style>
:root{--bg:#0f1419;--surface:#1a1f2e;--card:#1e2533;--border:#2d3748;--text:#e7e9ea;
--muted:#9ca3af;--accent:#60a5fa;--green:#34d399;--amber:#fbbf24;--red:#f87171;--purple:#a78bfa}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:system-ui,-apple-system,sans-serif;background:var(--bg);color:var(--text);display:flex;min-height:100vh}
/* Sidebar navigation */
.sidebar{width:240px;background:var(--surface);border-right:1px solid var(--border);
  position:fixed;top:0;left:0;bottom:0;display:flex;flex-direction:column;z-index:10}
.sidebar-header{padding:20px 16px 12px;border-bottom:1px solid var(--border)}
.sidebar-header h1{font-size:16px;font-weight:700;color:var(--accent)}
.sidebar-header p{font-size:11px;color:var(--muted);margin-top:4px}
.nav-list{list-style:none;padding:12px 0;flex:1}
.nav-item{display:block;padding:10px 16px 10px 20px;color:var(--muted);font-size:13px;
  cursor:pointer;border-left:3px solid transparent;transition:all .12s}
.nav-item:hover{color:var(--text);background:rgba(96,165,250,.05)}
.nav-item.active{color:var(--accent);border-left-color:var(--accent);background:rgba(96,165,250,.08)}
.nav-icon{margin-right:8px}
/* Main content */
.main{margin-left:240px;flex:1;padding:24px 32px;max-width:1200px}
.section{display:none}
.section.active{display:block}
.card{background:var(--card);border:1px solid var(--border);border-radius:8px;padding:20px;margin:0 0 16px}
.card h3{font-size:14px;font-weight:600;color:var(--accent);margin-bottom:12px}
.metrics{display:grid;grid-template-columns:repeat(auto-fill,minmax(170px,1fr));gap:12px;margin:12px 0}
.metric-box{background:rgba(35,43,59,.8);border-radius:6px;padding:12px 14px}
.metric-box .lbl{font-size:11px;color:var(--muted);text-transform:uppercase;letter-spacing:.4px}
.metric-box .val{font-size:20px;font-weight:700;margin-top:3px}
.val-green{color:var(--green)}.val-blue{color:var(--accent)}.val-amber{color:var(--amber)}.val-red{color:var(--red)}
table{border-collapse:collapse;width:100%;font-size:12px;margin:8px 0}
th,td{border:1px solid var(--border);padding:7px 10px;text-align:right}
th{background:rgba(35,43,59,.8);color:var(--muted);font-weight:600;text-transform:uppercase;font-size:10px;letter-spacing:.3px}
td{color:var(--text)}
tr:nth-child(even) td{background:rgba(15,20,25,.4)}
.finding{padding:8px 14px;margin:6px 0;border-radius:4px;font-size:13px;border-left:3px solid var(--green);background:rgba(26,46,26,.6)}
.finding.warn{border-left-color:var(--amber);background:rgba(46,42,26,.6)}
.finding.info{border-left-color:var(--accent);background:rgba(26,32,48,.6)}
svg{display:block;margin:8px 0}
.svg-label{fill:var(--muted);font-size:11px}
.svg-title{fill:var(--text);font-size:13px;font-weight:600}
.placeholder-box{border:2px dashed var(--border);border-radius:8px;padding:32px;text-align:center;color:var(--muted)}
.node-detail{display:none;background:var(--card);border:1px solid var(--accent);border-radius:6px;
  padding:16px;margin-top:12px;font-size:12px}
.node-detail.visible{display:block}
.node-detail td,.node-detail th{border-color:var(--border);text-align:left}
@media(max-width:900px){
  .sidebar{width:200px}
  .main{margin-left:200px;padding:16px}
}
@media(max-width:700px){
  .sidebar{position:static;width:100%}
  .main{margin-left:0}
  body{flex-direction:column}
}
</style>
`)
}

func writeLayout(b *strings.Builder, d reportData) {
	// Sidebar.
	b.WriteString(`<nav class="sidebar"><div class="sidebar-header">
<h1>PFRS Research</h1><p>Performance Analysis</p></div><ul class="nav-list">`)
	navItems := []struct{ id, icon, label string }{
		{"summary", "📋", "Summary"},
		{"search", "📈", "Search Progress"},
		{"tree", "🌳", "Search Tree"},
		{"sa", "🔥", "Simulated Annealing"},
		{"workers", "👷", "Workers"},
		{"diversity", "🌍", "Diversity"},
		{"compare", "⚖", "Compare Runs"},
	}
	for _, n := range navItems {
		b.WriteString(fmt.Sprintf(`<li class="nav-item" data-section="%s"><span class="nav-icon">%s</span>%s</li>`,
			n.id, n.icon, n.label))
	}
	b.WriteString(`</ul></nav>`)

	// Main content area.
	b.WriteString(`<main class="main">`)
	writeSummarySection(b, d)
	writeSearchSection(b, d)
	writeTreeSection(b, d)
	writeSASection(b, d)
	writeWorkersSection(b, d)
	writeDiversitySection(b, d)
	writeCompareSection(b, d)
	b.WriteString(`</main>`)
}

// --- Section: Summary ---

func writeSummarySection(b *strings.Builder, d reportData) {
	b.WriteString(`<div class="section" id="sec-summary">`)

	// Experiment configuration.
	b.WriteString(`<div class="card"><h3>Experiment</h3><div class="metrics">`)
	b.WriteString(metricBox("Algorithm", d.First.Mode, ""))
	b.WriteString(metricBox("Cooling Mode", d.First.CoolingMode, ""))
	b.WriteString(metricBox("Beam Width", fmt.Sprintf("%d", d.First.MaxConcurrent), ""))
	b.WriteString(metricBox("Beam Seeds", fmt.Sprintf("%d", d.First.Seed), ""))
	b.WriteString(metricBox("Candidate Budget", fmtComma(d.First.IterationsPerWorker), ""))
	b.WriteString(metricBox("Initial Temperature", fmt.Sprintf("%.1f", d.First.InitialTemperature), ""))
	b.WriteString(metricBox("Effective Cooling", fmt.Sprintf("%.8f", d.First.EffectiveCoolingRate), ""))
	b.WriteString(metricBox("Max Workers", fmt.Sprintf("%d", d.First.MaxTotalWorkers), ""))
	b.WriteString(metricBox("Instance", d.First.Instance, ""))
	b.WriteString(`</div></div>`)

	// Current result.
	b.WriteString(`<div class="card"><h3>Current Result</h3><div class="metrics">`)
	b.WriteString(metricBox("Total Penalty", fmtComma(d.TotalPenalty), "val-green"))
	b.WriteString(metricBox("Previous Best", "—", "val-blue"))
	b.WriteString(metricBox("Improvement", "—", "val-blue"))
	b.WriteString(metricBox("Runtime", fmt.Sprintf("%.1fs", float64(d.TotalDuration)/1000.0), "val-blue"))
	b.WriteString(metricBox("Total Candidates", fmtComma(d.TotalCandidates), "val-blue"))
	b.WriteString(metricBox("Workers Used", fmtComma(d.TotalWorkers), "val-blue"))
	b.WriteString(metricBox("Branches", fmtComma(d.TotalBranches), "val-blue"))
	b.WriteString(metricBox("Weeks", fmt.Sprintf("%d", d.NumWeeks), "val-blue"))
	b.WriteString(`</div></div>`)

	// Research findings.
	b.WriteString(`<div class="card"><h3>Research Findings</h3>`)
	writeFindings(b, d)
	b.WriteString(`</div>`)

	b.WriteString(`</div>`)
}

func writeFindings(b *strings.Builder, d reportData) {
	// Dominant week contribution.
	weekPct := 0.0
	if d.TotalPenalty > 0 {
		weekPct = float64(d.MaxWeekPenalty) / float64(d.TotalPenalty) * 100
	}
	b.WriteString(fmt.Sprintf(`<div class="finding warn">✓ Week %d contributes %.0f%% of total penalty</div>`, d.MaxWeekNum, weekPct))

	// Beam retention.
	maxRetained := 0
	for _, r := range d.Records {
		if r.WorkersStarted > maxRetained {
			maxRetained = r.WorkersStarted
		}
	}
	b.WriteString(fmt.Sprintf(`<div class="finding">✓ Beam retained up to %d histories</div>`, maxRetained))

	// Accepted worse moves.
	b.WriteString(fmt.Sprintf(`<div class="finding">✓ Accepted worse moves %.1f%%</div>`, d.AcceptWorseRate))

	// Hard reject rate.
	b.WriteString(fmt.Sprintf(`<div class="finding">✓ Hard reject rate %.0f%%</div>`, d.HardRejectRate))

	// Adaptive cooling.
	if d.First.CoolingMode == "adaptive" {
		b.WriteString(`<div class="finding">✓ Adaptive cooling active</div>`)
	}

	// Beam search completion.
	allValid := true
	for _, r := range d.Records {
		if r.HardViolations > 0 {
			allValid = false
			break
		}
	}
	if allValid {
		b.WriteString(`<div class="finding">✓ Beam search completed successfully</div>`)
	} else {
		b.WriteString(`<div class="finding warn">⚠ Some weeks have hard violations</div>`)
	}
}

// --- Section: Search Progress ---

func writeSearchSection(b *strings.Builder, d reportData) {
	b.WriteString(`<div class="section" id="sec-search">`)

	// Penalty by week.
	b.WriteString(`<div class="card"><h3>Penalty by Week</h3>`)
	writeSVGBarChart(b, d.Records, "Final Penalty per Week",
		func(r AuditCSVRecord) int { return r.FinalPenalty }, "#60a5fa")
	b.WriteString(`</div>`)

	// Cumulative penalty.
	b.WriteString(`<div class="card"><h3>Cumulative Penalty</h3>`)
	writeSVGLineChart(b, d.Records, d.CumPenalties, "Cumulative Penalty", "#34d399")
	b.WriteString(`</div>`)

	// Week contribution (percentage).
	b.WriteString(`<div class="card"><h3>Week Contribution (%)</h3>`)
	writeSVGBarChart(b, d.Records, "Week Share of Total (%)",
		func(r AuditCSVRecord) int {
			if d.TotalPenalty > 0 {
				return r.FinalPenalty * 100 / d.TotalPenalty
			}
			return 0
		}, "#a78bfa")
	b.WriteString(`</div>`)

	// Candidate efficiency.
	b.WriteString(`<div class="card"><h3>Candidate Efficiency</h3>`)
	effValues := make([]int, len(d.Records))
	for i, r := range d.Records {
		if r.Candidates > 0 {
			effValues[i] = r.Improvement * 1000 / r.Candidates
		}
	}
	writeSVGLineChart(b, d.Records, effValues, "Improvement per 1K Candidates", "#fbbf24")
	b.WriteString(`</div>`)

	// Workers vs candidates.
	b.WriteString(`<div class="card"><h3>Workers vs Candidates</h3><table>`)
	b.WriteString(`<tr><th>Week</th><th>Workers</th><th>Candidates</th><th>Candidates/Worker</th><th>Penalty</th><th>Improvement</th></tr>`)
	for _, r := range d.Records {
		cPerW := 0
		if r.WorkersStarted > 0 {
			cPerW = r.Candidates / r.WorkersStarted
		}
		b.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%d</td></tr>",
			r.Week, r.WorkersStarted, fmtComma(r.Candidates), fmtComma(cPerW),
			fmtComma(r.FinalPenalty), r.Improvement))
	}
	b.WriteString(`</table></div>`)

	b.WriteString(`</div>`)
}

// --- Section: Search Tree ---

func writeTreeSection(b *strings.Builder, d reportData) {
	b.WriteString(`<div class="section" id="sec-tree">`)

	// Winning lineage.
	b.WriteString(`<div class="card"><h3>Winning Lineage</h3><table>`)
	b.WriteString(`<tr><th>Week</th><th>Path ID</th><th>Seed</th><th>Penalty</th>`)
	b.WriteString(`<th>Cumulative</th><th>Workers</th><th>Branches</th><th>Depth</th></tr>`)
	for i, r := range d.Records {
		b.WriteString(fmt.Sprintf("<tr><td>%d</td><td>W%d</td><td>%d</td><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%d</td></tr>",
			r.Week, r.BestWorkerID, r.Seed, fmtComma(r.FinalPenalty), fmtComma(d.CumPenalties[i]),
			r.WorkersStarted, r.BranchesCreated, r.WinningBranchDepth))
	}
	b.WriteString(`</table></div>`)

	// Beam evolution chart.
	b.WriteString(`<div class="card"><h3>Beam Evolution</h3>`)
	writeSVGBarChart(b, d.Records, "Branches per Week",
		func(r AuditCSVRecord) int { return r.BranchesCreated }, "#a78bfa")
	b.WriteString(`</div>`)

	// Interactive SVG tree with clickable nodes.
	b.WriteString(`<div class="card"><h3>Search Tree</h3>`)
	writeInteractiveTree(b, d)
	b.WriteString(`<div class="node-detail" id="node-detail"><table id="node-detail-table">`)
	b.WriteString(`<tr><th>Property</th><th>Value</th></tr>`)
	b.WriteString(`<tr><td>Week</td><td id="nd-week">—</td></tr>`)
	b.WriteString(`<tr><td>Path ID</td><td id="nd-path">—</td></tr>`)
	b.WriteString(`<tr><td>Parent</td><td id="nd-parent">—</td></tr>`)
	b.WriteString(`<tr><td>Seed</td><td id="nd-seed">—</td></tr>`)
	b.WriteString(`<tr><td>Penalty</td><td id="nd-penalty">—</td></tr>`)
	b.WriteString(`<tr><td>Cumulative</td><td id="nd-cumulative">—</td></tr>`)
	b.WriteString(`<tr><td>Workers</td><td id="nd-workers">—</td></tr>`)
	b.WriteString(`<tr><td>Candidates</td><td id="nd-candidates">—</td></tr>`)
	b.WriteString(`<tr><td>Temperature</td><td id="nd-temp">—</td></tr>`)
	b.WriteString(`<tr><td>SA Accepted Worse</td><td id="nd-saworse">—</td></tr>`)
	b.WriteString(`</table></div></div>`)

	b.WriteString(`</div>`)
}

func writeInteractiveTree(b *strings.Builder, d reportData) {
	numWeeks := len(d.Records)
	colW := 110
	w := 60 + numWeeks*colW
	h := 220

	b.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg" id="tree-svg">`, w, h))

	// Horizontal winning-path connectors.
	for i := 0; i < numWeeks-1; i++ {
		x1 := 50 + i*colW + 50
		x2 := 50 + (i+1)*colW + 10
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="50" x2="%d" y2="50" stroke="#34d399" stroke-width="2" stroke-dasharray="4,2"/>`, x1, x2))
	}

	// Per-week nodes.
	for i, r := range d.Records {
		x := 50 + i*colW
		maxD := r.WinningBranchDepth
		if maxD < 1 {
			maxD = 1
		}
		// Draw branch nodes (up to 4 visible).
		for depth := 0; depth < maxD && depth < 4; depth++ {
			cy := 50 + depth*35
			isWinner := depth == maxD-1
			color := "#4b5563"
			strokeColor := "none"
			if isWinner {
				color = "#34d399"
				strokeColor = "#34d399"
			}
			// Clickable circle with data attributes.
			b.WriteString(fmt.Sprintf(
				`<circle cx="%d" cy="%d" r="10" fill="%s" stroke="%s" stroke-width="2" class="tree-node" `+
					`data-week="%d" data-path="W%d-D%d" data-parent="%s" data-seed="%d" `+
					`data-penalty="%d" data-cum="%d" data-workers="%d" data-cands="%d" `+
					`data-temp="%.1f" data-saworse="%d" style="cursor:pointer"/>`,
				x+30, cy, color, strokeColor,
				r.Week, r.BestWorkerID, depth, func() string { if depth == 0 { return "root" }; return fmt.Sprintf("W%d-D%d", r.BestWorkerID, depth-1) }(),
				r.Seed, r.FinalPenalty, d.CumPenalties[i], r.WorkersStarted, r.Candidates,
				r.InitialTemperature, r.SAAcceptedWorse))
			// Vertical connector.
			if depth > 0 {
				b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#4b5563" stroke-width="1"/>`,
					x+30, cy-25, x+30, cy-10))
			}
		}
		// Discarded paths indicator.
		if r.BranchesDropped > 0 {
			cy := 50 + maxD*35
			if cy < h-40 {
				b.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="6" fill="none" stroke="#f87171" stroke-width="1.5" stroke-dasharray="2,2"/>`,
					x+30, cy))
				b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="svg-label" fill="#f87171">×%d</text>`,
					x+30, cy+16, r.BranchesDropped))
			}
		}
		// Week label.
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="svg-label">W%d</text>`, x+30, h-8, r.Week))
	}

	b.WriteString(`</svg>`)
}

// --- Section: Simulated Annealing ---

func writeSASection(b *strings.Builder, d reportData) {
	b.WriteString(`<div class="section" id="sec-sa">`)

	// Acceptance chart.
	b.WriteString(`<div class="card"><h3>SA Acceptance by Week</h3>`)
	writeSVGStackedBar(b, d.Records, "SA Move Acceptance", []barSeries{
		{name: "Improving", color: "#34d399", fn: func(r AuditCSVRecord) int { return r.SAAcceptedBetter }},
		{name: "Worse (accepted)", color: "#fbbf24", fn: func(r AuditCSVRecord) int { return r.SAAcceptedWorse }},
		{name: "Rejected by prob", color: "#f87171", fn: func(r AuditCSVRecord) int { return r.SARejectedByProb }},
	})
	b.WriteString(`</div>`)

	// Temperature data table.
	b.WriteString(`<div class="card"><h3>Temperature Profile</h3><table>`)
	b.WriteString(`<tr><th>Week</th><th>Initial</th><th>At Best</th><th>Final</th>`)
	b.WriteString(`<th>Better</th><th>Worse</th><th>Rejected</th><th>Worse %</th></tr>`)
	for _, r := range d.Records {
		worsePct := 0.0
		if r.Candidates > 0 {
			worsePct = float64(r.SAAcceptedWorse) / float64(r.Candidates) * 100
		}
		b.WriteString(fmt.Sprintf(
			"<tr><td>%d</td><td>%.1f</td><td>%.4f</td><td>%.6f</td><td>%s</td><td>%s</td><td>%s</td><td>%.1f%%</td></tr>",
			r.Week, r.InitialTemperature, r.SATempAtBest, r.SAFinalTemp,
			fmtComma(r.SAAcceptedBetter), fmtComma(r.SAAcceptedWorse),
			fmtComma(r.SARejectedByProb), worsePct))
	}
	b.WriteString(`</table></div>`)

	// Temperature curve placeholder.
	b.WriteString(`<div class="card"><h3>Temperature Curve</h3>`)
	b.WriteString(`<div class="placeholder-box">Temperature decay curve will be plotted when per-iteration temperature samples are available.</div>`)
	b.WriteString(`</div>`)

	// Cooling schedule placeholder.
	b.WriteString(`<div class="card"><h3>Cooling Schedule</h3>`)
	b.WriteString(`<div class="placeholder-box">Cooling schedule analysis — adaptive rate adjustments over time.</div>`)
	b.WriteString(`</div>`)

	// Plateau detection placeholder.
	b.WriteString(`<div class="card"><h3>Plateau Detection</h3>`)
	b.WriteString(`<div class="placeholder-box">Plateau detection will identify regions where no improvement occurred despite continued search.</div>`)
	b.WriteString(`</div>`)

	b.WriteString(`</div>`)
}

// --- Section: Workers ---

func writeWorkersSection(b *strings.Builder, d reportData) {
	b.WriteString(`<div class="section" id="sec-workers">`)

	// Workers started.
	b.WriteString(`<div class="card"><h3>Workers Started</h3>`)
	writeSVGBarChart(b, d.Records, "Workers Started per Week",
		func(r AuditCSVRecord) int { return r.WorkersStarted }, "#a78bfa")
	b.WriteString(`</div>`)

	// Workers completed (all started workers run to completion in current PFRS).
	b.WriteString(`<div class="card"><h3>Workers Completed</h3>`)
	writeSVGBarChart(b, d.Records, "Workers Completed per Week",
		func(r AuditCSVRecord) int { return r.WorkersStarted }, "#60a5fa")
	b.WriteString(`</div>`)

	// Branches.
	b.WriteString(`<div class="card"><h3>Branches</h3>`)
	writeSVGStackedBar(b, d.Records, "Branching Activity", []barSeries{
		{name: "Created", color: "#34d399", fn: func(r AuditCSVRecord) int { return r.BranchesCreated }},
		{name: "Dropped", color: "#f87171", fn: func(r AuditCSVRecord) int { return r.BranchesDropped }},
	})
	b.WriteString(`</div>`)

	// Branch depth.
	b.WriteString(`<div class="card"><h3>Branch Depth</h3>`)
	writeSVGBarChart(b, d.Records, "Winning Branch Depth",
		func(r AuditCSVRecord) int { return r.WinningBranchDepth }, "#34d399")
	b.WriteString(`</div>`)

	// Worker lifetime (duration per week).
	b.WriteString(`<div class="card"><h3>Worker Lifetime</h3><table>`)
	b.WriteString(`<tr><th>Week</th><th>Duration</th><th>Workers</th><th>Avg/Worker</th>`)
	b.WriteString(`<th>Candidates</th><th>Cands/Worker</th></tr>`)
	for _, r := range d.Records {
		avgDur := float64(0)
		candsPerW := 0
		if r.WorkersStarted > 0 {
			avgDur = float64(r.DurationMs) / float64(r.WorkersStarted) / 1000.0
			candsPerW = r.Candidates / r.WorkersStarted
		}
		b.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%.1fs</td><td>%d</td><td>%.1fs</td><td>%s</td><td>%s</td></tr>",
			r.Week, float64(r.DurationMs)/1000.0, r.WorkersStarted, avgDur,
			fmtComma(r.Candidates), fmtComma(candsPerW)))
	}
	b.WriteString(`</table></div>`)

	// Worker contribution.
	b.WriteString(`<div class="card"><h3>Worker Contribution</h3><table>`)
	b.WriteString(`<tr><th>Week</th><th>Improved</th><th>Produced Best</th><th>Max Concurrent</th><th>Max Queue</th></tr>`)
	for _, r := range d.Records {
		b.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td></tr>",
			r.Week, r.WorkersImproved, r.WorkersProducedBest, r.MaxConcurrentSeen, r.MaxQueueDepth))
	}
	b.WriteString(`</table></div>`)

	b.WriteString(`</div>`)
}

// --- Section: Diversity ---

func writeDiversitySection(b *strings.Builder, d reportData) {
	b.WriteString(`<div class="section" id="sec-diversity">`)

	// Hard rejection breakdown (real data).
	b.WriteString(`<div class="card"><h3>Hard Rejection Breakdown</h3>`)
	writeSVGStackedBar(b, d.Records, "Rejection Categories by Week", []barSeries{
		{name: "No-op", color: "#6b7280", fn: func(r AuditCSVRecord) int { return r.RejectedNoop }},
		{name: "Skill", color: "#f87171", fn: func(r AuditCSVRecord) int { return r.RejectedSkill }},
		{name: "Succession", color: "#fbbf24", fn: func(r AuditCSVRecord) int { return r.RejectedSuccession }},
		{name: "History", color: "#a78bfa", fn: func(r AuditCSVRecord) int { return r.RejectedHistory }},
	})
	b.WriteString(`</div>`)

	// Beam spread placeholder.
	b.WriteString(`<div class="card"><h3>Beam Spread</h3>`)
	b.WriteString(`<div class="placeholder-box">Beam spread analysis — how retained paths diverge over weeks.</div>`)
	b.WriteString(`</div>`)

	// Hamming distance placeholder.
	b.WriteString(`<div class="card"><h3>Hamming Distance</h3>`)
	b.WriteString(`<div class="placeholder-box">Parent→child Hamming distance analysis will be available when tree CSV is wired into the visualiser.</div>`)
	b.WriteString(`</div>`)

	// Structural diversity placeholder.
	b.WriteString(`<div class="card"><h3>Structural Diversity</h3>`)
	b.WriteString(`<div class="placeholder-box">Roster structural diversity metrics across beam paths.</div>`)
	b.WriteString(`</div>`)

	// Near-duplicate detection placeholder.
	b.WriteString(`<div class="card"><h3>Near-Duplicate Detection</h3>`)
	b.WriteString(`<div class="placeholder-box">Detection of near-duplicate rosters (Hamming distance &lt; 5%) within the beam.</div>`)
	b.WriteString(`</div>`)

	b.WriteString(`</div>`)
}

// --- Section: Compare Runs ---

func writeCompareSection(b *strings.Builder, d reportData) {
	b.WriteString(`<div class="section" id="sec-compare">`)
	b.WriteString(`<div class="card"><h3>Compare Runs</h3>`)
	b.WriteString(`<div class="placeholder-box">Run comparison will be added in a future iteration.</div>`)
	b.WriteString(`</div>`)
	b.WriteString(`</div>`)
}

// --- JavaScript ---

func writeHTMLFooter(b *strings.Builder) {
	b.WriteString(`
<script>
(function(){
  // --- Navigation ---
  var navItems = document.querySelectorAll('.nav-item');
  var sections = document.querySelectorAll('.section');

  function activate(id) {
    navItems.forEach(function(n){ n.classList.toggle('active', n.getAttribute('data-section')===id); });
    sections.forEach(function(s){ s.classList.toggle('active', s.id==='sec-'+id); });
    history.replaceState(null,'','#'+id);
  }

  navItems.forEach(function(n){
    n.addEventListener('click', function(){ activate(this.getAttribute('data-section')); });
  });

  // Restore from fragment or default.
  var hash = location.hash.replace('#','');
  activate(hash || 'summary');

  // --- Tree node click ---
  var nodes = document.querySelectorAll('.tree-node');
  var detailPanel = document.getElementById('node-detail');
  nodes.forEach(function(node){
    node.addEventListener('click', function(){
      document.getElementById('nd-week').textContent = this.getAttribute('data-week');
      document.getElementById('nd-path').textContent = this.getAttribute('data-path');
      document.getElementById('nd-parent').textContent = this.getAttribute('data-parent');
      document.getElementById('nd-seed').textContent = this.getAttribute('data-seed');
      document.getElementById('nd-penalty').textContent = this.getAttribute('data-penalty');
      document.getElementById('nd-cumulative').textContent = this.getAttribute('data-cum');
      document.getElementById('nd-workers').textContent = this.getAttribute('data-workers');
      document.getElementById('nd-candidates').textContent = this.getAttribute('data-cands');
      document.getElementById('nd-temp').textContent = this.getAttribute('data-temp');
      document.getElementById('nd-saworse').textContent = this.getAttribute('data-saworse');
      detailPanel.classList.add('visible');
    });
  });
})();
</script>
</body></html>`)
}

// --- Chart Generation ---

type barSeries struct {
	name  string
	color string
	fn    func(AuditCSVRecord) int
}

func writeSVGBarChart(b *strings.Builder, records []AuditCSVRecord, title string,
	valueFn func(AuditCSVRecord) int, color string) {

	maxVal := 1
	for _, r := range records {
		if v := valueFn(r); v > maxVal {
			maxVal = v
		}
	}

	barW := 48
	gap := 18
	w := 50 + len(records)*(barW+gap)
	h := 220
	chartH := 155.0
	topPad := 28

	b.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, w, h))
	b.WriteString(fmt.Sprintf(`<text x="%d" y="16" text-anchor="middle" class="svg-title">%s</text>`, w/2, title))

	for i, r := range records {
		v := valueFn(r)
		x := 40 + i*(barW+gap)
		barH := float64(v) / float64(maxVal) * chartH
		y := topPad + int(chartH-barH)
		b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%.0f" fill="%s" rx="3"/>`,
			x, y, barW, barH, color))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="svg-label">%s</text>`,
			x+barW/2, y-3, fmtComma(v)))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="svg-label">W%d</text>`,
			x+barW/2, h-4, r.Week))
	}
	b.WriteString(`</svg>`)
}

func writeSVGLineChart(b *strings.Builder, records []AuditCSVRecord, values []int, title string, color string) {
	maxVal := 1
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	barW := 48
	gap := 18
	w := 50 + len(records)*(barW+gap)
	h := 220
	chartH := 155.0
	topPad := 28

	b.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, w, h))
	b.WriteString(fmt.Sprintf(`<text x="%d" y="16" text-anchor="middle" class="svg-title">%s</text>`, w/2, title))

	// Polyline.
	var points []string
	for i, v := range values {
		x := 40 + i*(barW+gap) + barW/2
		y := topPad + int(chartH-float64(v)/float64(maxVal)*chartH)
		points = append(points, fmt.Sprintf("%d,%d", x, y))
	}
	b.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="%s" stroke-width="2.5"/>`,
		strings.Join(points, " "), color))

	// Dots and labels.
	for i, v := range values {
		x := 40 + i*(barW+gap) + barW/2
		y := topPad + int(chartH-float64(v)/float64(maxVal)*chartH)
		b.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="4" fill="%s"/>`, x, y, color))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="svg-label">%s</text>`,
			x, y-7, fmtComma(v)))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="svg-label">W%d</text>`,
			x, h-4, records[i].Week))
	}
	b.WriteString(`</svg>`)
}

func writeSVGStackedBar(b *strings.Builder, records []AuditCSVRecord, title string, series []barSeries) {
	maxVal := 1
	for _, r := range records {
		total := 0
		for _, s := range series {
			total += s.fn(r)
		}
		if total > maxVal {
			maxVal = total
		}
	}

	barW := 48
	gap := 18
	w := 50 + len(records)*(barW+gap)
	h := 260
	chartH := 170.0
	topPad := 48

	b.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, w, h))
	b.WriteString(fmt.Sprintf(`<text x="%d" y="16" text-anchor="middle" class="svg-title">%s</text>`, w/2, title))

	// Legend.
	for i, s := range series {
		lx := 40 + i*130
		b.WriteString(fmt.Sprintf(`<rect x="%d" y="26" width="10" height="10" fill="%s"/>`, lx, s.color))
		b.WriteString(fmt.Sprintf(`<text x="%d" y="35" class="svg-label">%s</text>`, lx+14, s.name))
	}

	for ri, r := range records {
		x := 40 + ri*(barW+gap)
		yBase := topPad + int(chartH)
		for _, s := range series {
			v := s.fn(r)
			barH := float64(v) / float64(maxVal) * chartH
			yBase -= int(barH)
			b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%.0f" fill="%s" rx="2"/>`,
				x, yBase, barW, barH, s.color))
		}
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" class="svg-label">W%d</text>`,
			x+barW/2, h-4, r.Week))
	}
	b.WriteString(`</svg>`)
}

// --- Data Utilities ---

func metricBox(label, value, cls string) string {
	valCls := ""
	if cls != "" {
		valCls = " " + cls
	}
	return fmt.Sprintf(`<div class="metric-box"><div class="lbl">%s</div><div class="val%s">%s</div></div>`,
		label, valCls, value)
}

func fmtComma(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(ch))
	}
	return string(result)
}
