// Package workitem provides the WorkItem domain object.
//
// A WorkItem represents a unit of work that requires planning or optimisation.
// Work Items are created from one or more Business Events and become the
// primary input to the optimisation engine.
package workitem

import (
	"encoding/json"
	"errors"
	"strings"
)

// WorkItem represents a unit of work that may be scheduled or optimised.
//
// Work Items are immutable once created. They contain planning details
// required by the optimisation engine and are validated during construction.
type WorkItem struct {
	id       string
	workType string
	details  json.RawMessage
}

// New creates a validated WorkItem.
//
// It returns a domain error if any invariant is violated.
// A successfully constructed WorkItem is always in a valid state.
func New(id string, workType string, details json.RawMessage) (WorkItem, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return WorkItem{}, errors.New("work item id must not be empty")
	}

	workType = strings.TrimSpace(workType)
	if workType == "" {
		return WorkItem{}, errors.New("work item type must not be empty")
	}

	if len(details) == 0 {
		return WorkItem{}, errors.New("work item details must not be empty")
	}

	if !json.Valid(details) {
		return WorkItem{}, errors.New("work item details must be valid JSON")
	}

	detailsCopy := make(json.RawMessage, len(details))
	copy(detailsCopy, details)

	return WorkItem{
		id:       id,
		workType: workType,
		details:  detailsCopy,
	}, nil
}

// ID returns the unique identifier of the work item.
func (w WorkItem) ID() string {
	return w.id
}

// Type returns the work item type.
func (w WorkItem) Type() string {
	return w.workType
}

// Details returns the planning details as raw JSON.
//
// The returned slice is a defensive copy; callers cannot mutate internal state.
func (w WorkItem) Details() json.RawMessage {
	cp := make(json.RawMessage, len(w.details))
	copy(cp, w.details)
	return cp
}

// Equal returns true if two WorkItems have the same identity.
//
// Equality is determined by ID, not by comparing all fields.
func (w WorkItem) Equal(other WorkItem) bool {
	return w.id == other.id
}

// MarshalJSON implements json.Marshaler for serialisation.
func (w WorkItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID      string          `json:"id"`
		Type    string          `json:"type"`
		Details json.RawMessage `json:"details"`
	}{
		ID:      w.id,
		Type:    w.workType,
		Details: w.details,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// It applies the same validation as the constructor so that deserialised
// work items are subject to the same domain invariants.
func (w *WorkItem) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID      string          `json:"id"`
		Type    string          `json:"type"`
		Details json.RawMessage `json:"details"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsed, err := New(raw.ID, raw.Type, raw.Details)
	if err != nil {
		return err
	}

	*w = parsed
	return nil
}
