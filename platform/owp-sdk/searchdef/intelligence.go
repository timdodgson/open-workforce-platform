package searchdef

type Confidence float64

type SafetyStatus string

const (
	SafetyPassed   SafetyStatus = "passed"
	SafetyRejected SafetyStatus = "rejected"
)

type SearchAction string

const (
	SearchContinue     SearchAction = "continue"
	SearchEarlyStop    SearchAction = "early_stop"
	SearchRestart      SearchAction = "restart"
	SearchAdjustTemp   SearchAction = "adjust_temp"
	SearchAdjustLAHC   SearchAction = "adjust_lahc"
	SearchAdjustTabu   SearchAction = "adjust_tabu"
	SearchAdjustBudget SearchAction = "adjust_budget"
)

type SearchProgress struct {
	Algorithm          string
	IterationsComplete int
	IterationsTotal    int
	CurrentPenalty     int
	BestPenalty        int
	InitialPenalty     int
	ImprovementRate    float64
	Temperature        float64
	PlateauLength      int
	Accepted           int
	Rejected           int
	CandidatesEval     int
}

type SearchRecommendation struct {
	Action         SearchAction
	Confidence     Confidence
	Reasons        []string
	NewTemperature float64
	NewLAHCLength  int
	NewTabuTenure  int
	NewBudget      int
}

type SearchAssist interface {
	Checkpoint(progress SearchProgress) *SearchRecommendation
}
