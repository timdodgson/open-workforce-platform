package assist

import (
	"sync"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
)

// SearchAssistRecorder collects all assist decisions for a search run.
type SearchAssistRecorder struct {
	mu      sync.Mutex
	records []searchdef.SearchAssistRecord
}

// NewSearchAssistRecorder creates a new recorder.
func NewSearchAssistRecorder() *SearchAssistRecorder {
	return &SearchAssistRecorder{}
}

// Record adds a checkpoint record.
func (r *SearchAssistRecorder) Record(rec searchdef.SearchAssistRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec)
}

// FinaliseAll updates all records with the final search outcome.
func (r *SearchAssistRecorder) FinaliseAll(finalBest int, totalCandidates int, runtimeMs int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.records {
		r.records[i].FinalBestPenalty = finalBest
		r.records[i].TotalCandidates = totalCandidates
		r.records[i].RuntimeMs = runtimeMs
	}
}

// Records returns all collected records.
func (r *SearchAssistRecorder) Records() []searchdef.SearchAssistRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]searchdef.SearchAssistRecord{}, r.records...)
}
