package legacysearch

import (
	"encoding/json"
	"errors"
	"strings"
)

// Assignment represents the pairing of a Resource to a Work Item.
//
// Assignments are immutable once created and are validated during construction.
type Assignment struct {
	resourceID string
	workItemID string
}

// NewAssignment creates a validated Assignment.
func NewAssignment(resourceID string, workItemID string) (Assignment, error) {
	resourceID = strings.TrimSpace(resourceID)
	if resourceID == "" {
		return Assignment{}, errors.New("assignment resource id must not be empty")
	}

	workItemID = strings.TrimSpace(workItemID)
	if workItemID == "" {
		return Assignment{}, errors.New("assignment work item id must not be empty")
	}

	return Assignment{
		resourceID: resourceID,
		workItemID: workItemID,
	}, nil
}

// ResourceID returns the ID of the assigned resource.
func (a Assignment) ResourceID() string {
	return a.resourceID
}

// WorkItemID returns the ID of the assigned work item.
func (a Assignment) WorkItemID() string {
	return a.workItemID
}

// MarshalJSON implements json.Marshaler for serialisation.
func (a Assignment) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ResourceID string `json:"resourceId"`
		WorkItemID string `json:"workItemId"`
	}{
		ResourceID: a.resourceID,
		WorkItemID: a.workItemID,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// It applies the same validation as the constructor.
func (a *Assignment) UnmarshalJSON(data []byte) error {
	var raw struct {
		ResourceID string `json:"resourceId"`
		WorkItemID string `json:"workItemId"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	parsed, err := NewAssignment(raw.ResourceID, raw.WorkItemID)
	if err != nil {
		return err
	}

	*a = parsed
	return nil
}
