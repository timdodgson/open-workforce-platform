package main

import (
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// SearchSolveOptions holds parsed flags common to metaheuristic solve commands.
type SearchSolveOptions struct {
	Mode                 string
	Iterations           int
	Temperature          float64
	Seed                 int64
	LateAcceptanceLength int
	TabuTenure           int
	TabuNeighbourhood    int
	Portfolio            []string
	RunLabel             string
	Storage              storageConfig
	PortfolioModelPath   string
}

func parseSearchSolveOptions(args []string, defaultMode string, defaultIter int, defaultTemp float64, defaultSeed int64) SearchSolveOptions {
	opts := SearchSolveOptions{
		Mode:                 parseSearchMode(args, defaultMode),
		Iterations:           parseSearchIterations(args, defaultIter),
		Temperature:          parseSearchTemperature(args, defaultTemp),
		Seed:                 parseSearchSeed(args, defaultSeed),
		LateAcceptanceLength: 1000,
		TabuTenure:           7,
		TabuNeighbourhood:    100,
		RunLabel:             parseRunLabelFlag(args, false),
		Storage:              parseStorageConfig(args, false),
		PortfolioModelPath:   parseStringFlag(args, "--portfolio-model"),
	}
	if v := parseIntFlag(args, "--late-acceptance-length"); v > 0 {
		opts.LateAcceptanceLength = v
	}
	if v := parseIntFlag(args, "--tabu-tenure"); v > 0 {
		opts.TabuTenure = v
	}
	if v := parseIntFlag(args, "--tabu-neighbourhood"); v > 0 {
		opts.TabuNeighbourhood = v
	}
	if portfolioStr := parseStringFlag(args, "--portfolio"); portfolioStr != "" {
		opts.Portfolio = strings.Split(portfolioStr, ",")
	}
	return opts
}

func (o SearchSolveOptions) BuildSearchConfig(domain, instance string, overrides func(*optimisation.SearchConfig)) optimisation.SearchConfig {
	cfg := optimisation.SearchConfig{
		Mode:                 o.Mode,
		Iterations:           o.Iterations,
		InitialTemperature:   o.Temperature,
		MinTemperature:       0.0001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: o.LateAcceptanceLength,
		TabuTenure:           o.TabuTenure,
		TabuNeighbourhood:    o.TabuNeighbourhood,
		Portfolio:            o.Portfolio,
		Seed:                 o.Seed,
		PolicyDomain:         domain,
		PolicyInstance:       instance,
	}
	if overrides != nil {
		overrides(&cfg)
	}
	return cfg
}
