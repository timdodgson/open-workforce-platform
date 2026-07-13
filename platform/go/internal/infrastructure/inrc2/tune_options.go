package inrc2

// TuneOptions holds domain configuration for a PFRS tuning run.
type TuneOptions struct {
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
	// MidHorizonWeek is 1-indexed checkpoint for S7/S8 exposure (0 = auto when MidHorizonWeight set).
	MidHorizonWeek int
	// MidHorizonWeight is λ for mid-horizon selection; 0 = telemetry only when week is set.
	MidHorizonWeight float64
	// MidHorizonSecondHalfIter boosts IterationsPerWorker for weeks after the checkpoint (0 = off).
	MidHorizonSecondHalfIter int
	RefinementMode           string
	RefinementIter           int
	RefinementTemp           float64
	WorkerMode               string
	Portfolio                []string
	LAHCBufferLength         int
	WorkerDecisionMode       string
	PolicyMode               string
	PolicyDir                string
	RunLabel                 string
	// HistoryIndex selects H0-*-<n>.json; -1 means default (first sorted history file).
	HistoryIndex int
	// WeekSequence is a competition week order like "6-2-9-1"; empty means first N sorted WD files.
	WeekSequence string
}

func (o TuneOptions) SingleConfig() bool {
	return o.OverrideIter > 0 || o.OverrideWorkers > 0 || o.OverrideTemp > 0 || o.OverrideRate > 0 ||
		o.BeamWidth > 1 || len(o.BeamSeeds) > 0
}

func (o TuneOptions) UseBeamSearch() bool {
	return o.BeamWidth > 1 || len(o.BeamSeeds) > 0
}

func (o TuneOptions) BuildGrid() []TuningGridEntry {
	if !o.SingleConfig() {
		return GenerateGrid(
			[]int{30000, 60000, 100000},
			[]int{16, 32},
			[]float64{1.0, 2.0, 5.0},
			[]float64{0.0009, 0.0005, 0.0001},
		)
	}

	defaults := DefaultPFRSConfig()
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
	return []TuningGridEntry{{
		IterationsPerWorker: iter,
		MaxTotalWorkers:     workers,
		InitialTemperature:  temp,
		CoolingRate:         rate,
	}}
}
