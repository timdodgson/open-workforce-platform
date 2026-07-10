package searchdef

type SearchAssistConfig struct {
	Mode               string
	CheckpointInterval int
	MinBudgetFraction  float64
	StagnationWindow   int
	RecentImprovWindow int
}

func DefaultSearchAssistConfig() SearchAssistConfig {
	return SearchAssistConfig{
		Mode:               "off",
		CheckpointInterval: 10000,
		MinBudgetFraction:  0.20,
		StagnationWindow:   50000,
		RecentImprovWindow: 5000,
	}
}

type SearchAssistRecord struct {
	Algorithm       string
	Checkpoint      int
	Candidates      int
	IterationsTotal int
	CurrentPenalty  int
	BestPenalty     int
	InitialPenalty  int
	Temperature     float64
	PlateauLength   int
	ImprovementRate float64
	RecommendedAction SearchAction
	Confidence        Confidence
	Reasons           string
	SafetyTriggered bool
	SafetyRule      string
	Accepted        bool
	FinalAction     SearchAction
	FinalBestPenalty int
	TotalCandidates  int
	RuntimeMs        int64
}
