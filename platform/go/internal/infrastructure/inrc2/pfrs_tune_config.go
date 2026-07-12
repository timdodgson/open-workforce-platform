package inrc2

// TuneConfigParams configures PFRS solver settings for a tuning run.
type TuneConfigParams struct {
	Entry            TuningGridEntry
	Mode             string
	Portfolio        []string
	MaxConcurrent    int
	CoolingMode      string
	Seed             int64
	LAHCBufferLength int
	ReheatThreshold  int
	ReheatFactor     float64
	ReheatMinFraction float64
	NoReheat         bool
	Beam             bool // beam runs enable reheat/tabu/branch defaults

	DecisionEngine   WorkerDecisionEngine
	DecisionRecorder *ShadowRecorder
	AssistMode       bool
	AssistRecorder   *AssistRecorder

	OnProgress         ProgressFunc
	ProgressIntervalMs int64
	OnAudit            AuditFunc
}

// BuildTunePFRSConfig assembles a PFRSConfig for standard or beam tuning.
func BuildTunePFRSConfig(p TuneConfigParams) PFRSConfig {
	cfg := PFRSConfig{
		Mode:                 p.Mode,
		Portfolio:            p.Portfolio,
		IterationsPerWorker:  p.Entry.IterationsPerWorker,
		MaxConcurrentWorkers: p.MaxConcurrent,
		MaxTotalWorkers:      p.Entry.MaxTotalWorkers,
		BranchOnGlobalBest:   true,
		InitialTemperature:   p.Entry.InitialTemperature,
		CoolingRate:          p.Entry.CoolingRate,
		CoolingMode:          p.CoolingMode,
		MinTemperature:       0.0001,
		LateAcceptanceLength: 1000,
		Seed:                 p.Seed,
		Deterministic:        true,
		ScoringMode:          "official-penalty",
		OnProgress:           p.OnProgress,
		ProgressIntervalMs:   p.ProgressIntervalMs,
		OnAudit:              p.OnAudit,
		DecisionEngine:       p.DecisionEngine,
		DecisionRecorder:     p.DecisionRecorder,
		AssistMode:           p.AssistMode,
		AssistRecorder:       p.AssistRecorder,
	}

	if p.Beam {
		cfg.TabuTenure = 7
		cfg.BranchCooldown = 25000
		cfg.ReheatEnabled = !p.NoReheat
		cfg.ReheatThreshold = 50000
		cfg.ReheatFactor = 1.0
		cfg.ReheatMinCandidateFraction = 0.20
		if p.ReheatThreshold > 0 {
			cfg.ReheatThreshold = p.ReheatThreshold
		}
		if p.ReheatFactor > 0 {
			cfg.ReheatFactor = p.ReheatFactor
		}
		if p.ReheatMinFraction > 0 {
			cfg.ReheatMinCandidateFraction = p.ReheatMinFraction
		}
	}

	if p.Mode == "lahc" {
		cfg.LateAcceptanceLength = int(float64(cfg.IterationsPerWorker) * 0.03)
		if cfg.LateAcceptanceLength < 1000 {
			cfg.LateAcceptanceLength = 1000
		}
	}
	if p.LAHCBufferLength > 0 {
		cfg.LateAcceptanceLength = p.LAHCBufferLength
	}
	if len(p.Portfolio) > 0 {
		cfg.Portfolio = p.Portfolio
		cfg.Mode = "portfolio"
	} else if p.Mode == "portfolio" {
		cfg.Portfolio = DefaultPortfolioStrategies
	}

	return cfg
}
