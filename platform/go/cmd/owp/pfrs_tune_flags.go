package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// TunePFRSOptions holds parsed tune-pfrs CLI configuration.
type TunePFRSOptions struct {
	inrc2.TuneOptions
	Storage storageConfig
}

func parseTunePFRSOptions(args []string) TunePFRSOptions {
	opts := TunePFRSOptions{
		TuneOptions: inrc2.TuneOptions{
			InstanceName:        "n012w8",
			MaxConcurrent:       runtime.NumCPU(),
			ProgressEnabled:     true,
			ProgressIntervalSec: 10,
			Seeds:               []int64{42},
			AuditCSVPath:        "../web/pfrs-lab/data/results.csv",
			TreeCSVPath:         "../web/pfrs-lab/data/tree.csv",
			BeamWidth:           1,
			CoolingMode:         "adaptive",
			FinalWindowWeeks:    1,
			RefinementMode:      "none",
			RefinementIter:      100000,
			RefinementTemp:      10.0,
			WorkerMode:          "sa",
			BeamStrategy:        "none",
			HistoryIndex:        -1,
		},
	}

	if v := parseStringFlag(args, "--instance"); v != "" {
		opts.InstanceName = v
	}
	// Competition instance slice: e.g. --history 1 --weeks 6-2-9-1 → n030w4_1_6-2-9-1
	if hasFlag(args, "--history") {
		opts.HistoryIndex = parseIntFlag(args, "--history")
	}
	if v := parseStringFlag(args, "--weeks"); v != "" {
		opts.WeekSequence = v
	}
	if v := parseIntFlag(args, "--pfrs-max-concurrent"); v > 0 {
		opts.MaxConcurrent = v
	}
	opts.ShowInvalid = parseShowInvalidFlag(args)

	if v := parseBoolFlag(args, "--progress"); v == "false" {
		opts.ProgressEnabled = false
	}
	if v := parseStringFlag(args, "--progress-interval"); v != "" {
		opts.ProgressIntervalSec = parseProgressIntervalSec(v)
	}

	if seedStr := parseStringFlag(args, "--seeds"); seedStr != "" {
		opts.Seeds = parseSeedList(seedStr)
	}
	if v := parseStringFlag(args, "--audit-csv"); v != "" {
		opts.AuditCSVPath = v
	}
	if v := parseStringFlag(args, "--tree-csv"); v != "" {
		opts.TreeCSVPath = v
	}

	if v := parseIntFlag(args, "--pfrs-beam-width"); v > 0 {
		opts.BeamWidth = v
	}
	if beamSeedStr := parseStringFlag(args, "--pfrs-beam-seeds"); beamSeedStr != "" {
		opts.BeamSeeds = parseSeedList(beamSeedStr)
	}

	opts.OverrideIter = parseIntFlag(args, "--pfrs-iterations-per-worker")
	if opts.OverrideIter == 0 {
		opts.OverrideIter = parseIntFlag(args, "--iterations")
	}
	opts.OverrideWorkers = parseIntFlag(args, "--pfrs-max-total-workers")
	opts.OverrideTemp = parseFloatFlag(args, "--pfrs-initial-temperature")
	if opts.OverrideTemp == 0 {
		opts.OverrideTemp = parseFloatFlag(args, "--temperature")
	}
	opts.OverrideRate = parseFloatFlag(args, "--pfrs-cooling-rate")

	coolingMode := parseStringFlag(args, "--pfrs-cooling-mode")
	if coolingMode == "" {
		coolingMode = parseStringFlag(args, "--cooling")
	}
	if coolingMode == "" {
		if opts.OverrideRate > 0 {
			coolingMode = "fixed-rate"
		} else {
			coolingMode = "adaptive"
		}
	}
	if coolingMode != "adaptive" && coolingMode != "fixed-rate" {
		fmt.Fprintf(os.Stderr, "Invalid cooling mode: %s (must be adaptive or fixed-rate)\n", coolingMode)
		os.Exit(1)
	}
	opts.CoolingMode = coolingMode

	opts.ReheatThreshold = parseIntFlag(args, "--pfrs-reheat-threshold")
	opts.ReheatFactor = parseFloatFlag(args, "--pfrs-reheat-factor")
	opts.ReheatMinFraction = parseFloatFlag(args, "--pfrs-reheat-min-fraction")
	for _, arg := range args {
		if arg == "--pfrs-no-reheat" {
			opts.NoReheat = true
		}
	}

	if v := parseIntFlag(args, "--pfrs-final-window-weeks"); v > 0 {
		opts.FinalWindowWeeks = v
	}
	opts.FinalWindowIter = parseIntFlag(args, "--pfrs-final-window-iterations")
	opts.LookaheadWeight = parseFloatFlag(args, "--pfrs-lookahead-weight")
	opts.DiversitySlotsPct = parseIntFlag(args, "--pfrs-diversity-slots")
	if v := parseIntFlag(args, "--pfrs-mid-horizon-week"); v > 0 {
		opts.MidHorizonWeek = v
	}
	opts.MidHorizonWeight = parseFloatFlag(args, "--pfrs-mid-horizon-weight")
	opts.MidHorizonSecondHalfIter = parseIntFlag(args, "--pfrs-mid-horizon-second-half-iterations")

	beamStrategy := parseStringFlag(args, "--pfrs-beam-strategy")
	if beamStrategy == "" {
		if opts.LookaheadWeight > 0 {
			beamStrategy = "lookahead"
		} else {
			beamStrategy = "none"
		}
	}
	if beamStrategy != "none" && beamStrategy != "lookahead" && beamStrategy != "budget" {
		fmt.Fprintf(os.Stderr, "Invalid --pfrs-beam-strategy: %s (must be none, lookahead, or budget)\n", beamStrategy)
		os.Exit(1)
	}
	opts.BeamStrategy = beamStrategy

	if v := parseStringFlag(args, "--pfrs-refinement"); v != "" {
		opts.RefinementMode = v
	}
	if v := parseIntFlag(args, "--pfrs-refinement-iterations"); v > 0 {
		opts.RefinementIter = v
	}
	if v := parseFloatFlag(args, "--pfrs-refinement-temperature"); v > 0 {
		opts.RefinementTemp = v
	}

	workerMode := parseStringFlag(args, "--pfrs-mode")
	if workerMode == "" {
		workerMode = "sa"
	}
	if workerMode != "sa" && workerMode != "lahc" && workerMode != "tabu" && workerMode != "ga" && workerMode != "portfolio" {
		fmt.Fprintf(os.Stderr, "Invalid --pfrs-mode: %s (must be sa, lahc, tabu, ga, or portfolio)\n", workerMode)
		os.Exit(1)
	}
	opts.WorkerMode = workerMode

	if portfolioStr := parseStringFlag(args, "--pfrs-portfolio"); portfolioStr != "" {
		opts.Portfolio = strings.Split(portfolioStr, ",")
		opts.WorkerMode = "portfolio"
	}
	opts.LAHCBufferLength = parseIntFlag(args, "--pfrs-late-acceptance-length")

	opts.WorkerDecisionMode, opts.PolicyMode, opts.PolicyDir = applyPFRSIntelligenceFlags(args)
	opts.RunLabel = parseRunLabelFlag(args, true)
	opts.Storage = parseStorageConfig(args, true)

	if opts.RunLabel != "" {
		labelDir := ensureRunOutputDir(opts.RunLabel)
		opts.AuditCSVPath = filepath.Join(labelDir, "results.csv")
		opts.TreeCSVPath = filepath.Join(labelDir, "tree.csv")
	}

	return opts
}

func parseProgressIntervalSec(v string) int {
	v = strings.TrimSuffix(v, "s")
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			fmt.Fprintf(os.Stderr, "Invalid --progress-interval: %s\n", v)
			os.Exit(1)
		}
		n = n*10 + int(ch-'0')
	}
	if n > 0 {
		return n
	}
	return 10
}
