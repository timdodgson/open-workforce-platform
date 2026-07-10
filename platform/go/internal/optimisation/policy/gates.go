package policy

// GateResult captures whether a pipeline gate passed or failed.
type GateResult struct {
	Gate      string
	Passed    bool
	Reason    string
	Value     float64
	Threshold float64
}
