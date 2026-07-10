package workerlearning

// SearchOutcome captures search statistics for a single-search worker learning record.
// Domain-agnostic: callers map from optimisation.SearchResult or other engines.
type SearchOutcome struct {
	InitialPenalty int
	BestPenalty    int
	DurationMs     int64
	Candidates     int
	Accepted       int
	Rejected       int
}
