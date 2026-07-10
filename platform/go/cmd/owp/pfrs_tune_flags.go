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
	InstanceName        string
	MaxConcurrent       int
	ShowInvalid         bool
	ProgressEnabled     bool
	ProgressIntervalSec int
	Seeds               []int64
	AuditCSVPath        string
	TreeCSVPath         string
	BeamWidth           int
	BeamSeeds           []int64
	OverrideIter        int
	OverrideWorkers     int
	OverrideTemp        float64
	OverrideRate        float64
	CoolingMode         string
	ReheatThreshold     int
	ReheatFactor        float64
	ReheatMinFraction   float64
	NoReheat            bool
	FinalWindowWeeks    int
	FinalWindowIter     int
	LookaheadWeight     float64
	DiversitySlotsPct   int
	BeamStrategy        string
	RefinementMode      string
	RefinementIter      int
	RefinementTemp      float64
	WorkerMode          string
	Portfolio           []string
	LAHCBufferLength    int
	WorkerDecisionMode  string
	PolicyMode          string
	PolicyDir           string
	RunLabel            string
	Storage             storageConfig
}

func (o TunePFRSOptions) SingleConfig() bool {
	return o.OverrideIter > 0 || o.OverrideWorkers > 0 || o.OverrideTemp > 0 || o.OverrideRate > 0 ||
		o.BeamWidth > 1 || len(o.BeamSeeds) > 0
}

func (o TunePFRSOptions) UseBeamSearch() bool {
	return o.BeamWidth > 1 || len(o.BeamSeeds) > 0
}

func (o TunePFRSOptions) BuildGrid() []inrc2.TuningGridEntry {
	if !o.SingleConfig() {
		return inrc2.GenerateGrid(
			[]int{30000, 60000, 100000},
			[]int{16, 32},
			[]float64{1.0, 2.0, 5.0},
			[]float64{0.0009, 0.0005, 0.0001},
		)
	}

	defaults := inrc2.DefaultPFRSConfig()
	iter := o.OverrideIter
	if iter <= 0 {
		iter = defaults.IterationsPerWorker
	}
	workers := o.OverrideWorkers
	if workers <= 0 {
		workers = defaults.MaxTotalWorkers
	}
	temp := o.OverrideTemp
	if temp <= 0 {
		temp = defaults.InitialTemperature
	}
	rate := o.OverrideRate
	if rate <= 0 {
		rate = defaults.CoolingRate
	}
	return []inrc2.TuningGridEntry{{
		IterationsPerWorker: iter,
		MaxTotalWorkers:     workers,
		InitialTemperature:  temp,
		CoolingRate:         rate,
	}}
}

func parseTunePFRSOptions(args []string) TunePFRSOptions {
	opts := TunePFRSOptions{
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
	}

	if v := parseStringFlag(args, "--instance"); v != "" {
		opts.InstanceName = v
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
	if workerMode != "sa" && workerMode != "lahc" && workerMode != "tabu" && workerMode != "portfolio" {
		fmt.Fprintf(os.Stderr, "Invalid --pfrs-mode: %s (must be sa, lahc, tabu, or portfolio)\n", workerMode)
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
