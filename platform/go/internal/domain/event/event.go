// Package event provides the BusinessEvent domain object.
//
// A BusinessEvent represents something that has happened in the real world
// that creates, changes or removes work. It is the starting point of the
// optimisation pipeline.
package event

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// BusinessEvent represents a historical business fact that has entered the platform.
//
// Business Events are immutable once created. They contain opaque business
// details owned by the producing system and are validated during construction.
type BusinessEvent struct {
	id         string
	eventType  string
	occurredAt time.Time
	details    json.RawMessage
}

// New creates a validated BusinessEvent.
//
// It returns a domain error if any invariant is violated.
// A successfully constructed BusinessEvent is always in a valid state.
func New(id string, eventType string, occurredAt time.Time, details json.RawMessage) (BusinessEvent, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return BusinessEvent{}, errors.New("business event id must not be empty")
	}

	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		return BusinessEvent{}, errors.New("business event type must not be empty")
	}

	if occurredAt.IsZero() {
		return BusinessEvent{}, errors.New("business event occurrence time must not be zero")
	}

	if len(details) == 0 {
		return BusinessEvent{}, errors.New("business event details must not be empty")
	}

	if !json.Valid(details) {
		return BusinessEvent{}, errors.New("business event details must be valid JSON")
	}

	detailsCopy := make(json.RawMessage, len(details))
	copy(detailsCopy, details)

	return BusinessEvent{
		id:         id,
		eventType:  eventType,
		occurredAt: occurredAt,
		details:    detailsCopy,
	}, nil
}

// ID returns the unique identifier of the business event.
func (e BusinessEvent) ID() string {
	return e.id
}

// Type returns the business-defined event type.
func (e BusinessEvent) Type() string {
	return e.eventType
}

// OccurredAt returns the time the business event occurred.
func (e BusinessEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// Details returns the opaque business details as raw JSON.
//
// The meaning of the details is owned by the producing system.
func (e BusinessEvent) Details() json.RawMessage {
	cp := make(json.RawMessage, len(e.details))
	copy(cp, e.details)
	return cp
}

// Equal returns true if two BusinessEvents have the same identity.
//
// Equality is determined by ID, not by comparing all fields.
func (e BusinessEvent) Equal(other BusinessEvent) bool {
	return e.id == other.id
}

// MarshalJSON implements json.Marshaler for serialisation.
func (e BusinessEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID         string          `json:"id"`
		Type       string          `json:"type"`
		OccurredAt time.Time       `json:"occurredAt"`
		Details    json.RawMessage `json:"details"`
	}{
		ID:         e.id,
		Type:       e.eventType,
		OccurredAt: e.occurredAt,
		Details:    e.details,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// It applies the same validation as the constructor so that deserialised
// events are subject to the same domain invariants.
func (e *BusinessEvent) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID         string          `json:"id"`
		Type       string          `json:"type"`
		OccurredAt time.Time       `json:"occurredAt"`
		Details    json.RawMessage `json:"details"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsed, err := New(raw.ID, raw.Type, raw.OccurredAt, raw.Details)
	if err != nil {
		return err
	}

	*e = parsed
	return nil
}
