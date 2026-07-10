package main

import (
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func printPFRSAuditSummary(disp cli.Options, auditRows []inrc2.WeekAuditRow) {
	if len(auditRows) == 0 {
		return
	}

	fmt.Println()
	fmt.Println(disp.Heading(cli.EmojiConfig, "Audit Summary"))
	fmt.Println()

	totalCands := 0
	totalAccepted := 0
	totalRejected := 0
	totalRejNoop := 0
	totalRejSkill := 0
	totalRejSucc := 0
	totalRejHist := 0
	totalSABetter := 0
	totalSAWorse := 0
	totalSARejProb := 0
	totalLAHCCurrent := 0
	totalLAHCLate := 0
	totalLAHCRejLate := 0
	totalBranches := 0
	totalDropped := 0
	totalWorkers := 0
	totalImproved := 0
	totalProducedBest := 0
	maxDepth := 0
	var totalDurationMs int64

	for _, r := range auditRows {
		totalCands += r.Candidates
		totalAccepted += r.Accepted
		totalRejected += r.Rejected
		totalRejNoop += r.RejectedNoop
		totalRejSkill += r.RejectedSkill
		totalRejSucc += r.RejectedSuccession
		totalRejHist += r.RejectedHistory
		totalSABetter += r.SAAcceptedBetter
		totalSAWorse += r.SAAcceptedWorse
		totalSARejProb += r.SARejectedByProb
		totalLAHCCurrent += r.LAHCAcceptedByCurrent
		totalLAHCLate += r.LAHCAcceptedByLate
		totalLAHCRejLate += r.LAHCRejectedByLate
		totalBranches += r.BranchesCreated
		totalDropped += r.BranchesDropped
		totalWorkers += r.WorkersStarted
		totalImproved += r.WorkersImproved
		totalProducedBest += r.WorkersProducedBest
		totalDurationMs += r.DurationMs
		if r.WinningBranchDepth > maxDepth {
			maxDepth = r.WinningBranchDepth
		}
	}

	acceptRate := 0.0
	if totalCands > 0 {
		acceptRate = float64(totalAccepted) / float64(totalCands) * 100
	}

	fmt.Printf("  Candidates:  %s\n", cli.FormatInt(totalCands))
	fmt.Printf("  Accepted:    %s (%.1f%%)\n", cli.FormatInt(totalAccepted), acceptRate)
	fmt.Printf("  Rejected:    %s\n", cli.FormatInt(totalRejected))
	fmt.Println()
	fmt.Println("  Rejection Breakdown:")
	fmt.Printf("    No-op (same assignment): %s\n", cli.FormatInt(totalRejNoop))
	fmt.Printf("    Skill mismatch:          %s\n", cli.FormatInt(totalRejSkill))
	fmt.Printf("    Forbidden succession:    %s\n", cli.FormatInt(totalRejSucc))
	fmt.Printf("    History succession:       %s\n", cli.FormatInt(totalRejHist))
	fmt.Println()

	if auditRows[0].Mode == "sa" {
		fmt.Println("  SA Acceptance:")
		fmt.Printf("    Accepted (improving):    %s\n", cli.FormatInt(totalSABetter))
		fmt.Printf("    Accepted (worse, prob):  %s\n", cli.FormatInt(totalSAWorse))
		fmt.Printf("    Rejected (by prob):      %s\n", cli.FormatInt(totalSARejProb))
		fmt.Println()
	} else {
		fmt.Println("  LAHC Acceptance:")
		fmt.Printf("    Accepted (current):      %s\n", cli.FormatInt(totalLAHCCurrent))
		fmt.Printf("    Accepted (late score):   %s\n", cli.FormatInt(totalLAHCLate))
		fmt.Printf("    Rejected (late score):   %s\n", cli.FormatInt(totalLAHCRejLate))
		fmt.Println()
	}

	fmt.Println("  Branching:")
	fmt.Printf("    Total workers:           %s\n", cli.FormatInt(totalWorkers))
	fmt.Printf("    Branches created:        %s\n", cli.FormatInt(totalBranches))
	fmt.Printf("    Branches dropped:        %s\n", cli.FormatInt(totalDropped))
	fmt.Printf("    Max winning depth:       %d\n", maxDepth)
	fmt.Printf("    Workers improved parent: %s\n", cli.FormatInt(totalImproved))
	fmt.Printf("    Workers produced best:   %s\n", cli.FormatInt(totalProducedBest))
	fmt.Println()

	fmt.Println("  Per-Week:")
	fmt.Printf("    %-5s %8s %8s %8s %10s %8s %14s %10s\n",
		"Week", "Start", "Final", "Δ", "Workers", "Branches", "Candidates", "Time")
	for _, r := range auditRows {
		fmt.Printf("    %-5d %8d %8d %8d %10d %8d %14s %10s\n",
			r.Week, r.StartPenalty, r.FinalPenalty, r.Improvement,
			r.WorkersStarted, r.BranchesCreated,
			cli.FormatInt(r.Candidates), cli.FormatMs(r.DurationMs))
	}
	fmt.Println()
}
