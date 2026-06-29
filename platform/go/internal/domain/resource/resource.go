// Package resource provides the Resource domain object.
//
// A Resource represents something capable of completing Work Items.
// Resources may represent people, vehicles, teams, equipment, AI agents,
// robots or external providers.
package resource

import (
	"encoding/json"
	"errors"
	"strings"
)

// Resource represents available supply within the platform.
//
// Resources are immutable once created. They contain opaque business
// details and are validated during construction.
type Resource struct {
	id           string
	resourceType string
	details      json.RawMessage
}

// New creates a validated Resource.
//
// It returns a domain error if any invariant is violated.
// A successfully constructed Resource is always in a valid state.
func New(id string, resourceType string, details json.RawMessage) (Resource, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Resource{}, errors.New("resource id must not be empty")
	}

	resourceType = strings.TrimSpace(resourceType)
	if resourceType == "" {
		return Resource{}, errors.New("resource type must not be empty")
	}

	if len(details) == 0 {
		return Resource{}, errors.New("resource details must not be empty")
	}

	if !json.Valid(details) {
		return Resource{}, errors.New("resource details must be valid JSON")
	}

	detailsCopy := make(json.RawMessage, len(details))
	copy(detailsCopy, details)

	return Resource{
		id:           id,
		resourceType: resourceType,
		details:      detailsCopy,
	}, nil
}

// ID returns the unique identifier of the resource.
func (r Resource) ID() string {
	return r.id
}

// Type returns the resource type.
func (r Resource) Type() string {
	return r.resourceType
}

// Details returns the opaque business details as raw JSON.
//
// The returned slice is a defensive copy; callers cannot mutate internal state.
func (r Resource) Details() json.RawMessage {
	cp := make(json.RawMessage, len(r.details))
	copy(cp, r.details)
	return cp
}

// Equal returns true if two Resources have the same identity.
//
// Equality is determined by ID, not by comparing all fields.
func (r Resource) Equal(other Resource) bool {
	return r.id == other.id
}

// MarshalJSON implements json.Marshaler for serialisation.
func (r Resource) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID      string          `json:"id"`
		Type    string          `json:"type"`
		Details json.RawMessage `json:"details"`
	}{
		ID:      r.id,
		Type:    r.resourceType,
		Details: r.details,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// It applies the same validation as the constructor so that deserialised
// resources are subject to the same domain invariants.
func (r *Resource) UnmarshalJSON(data []byte) error {
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

	*r = parsed
	return nil
}
