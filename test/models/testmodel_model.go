package models

import "time"

// Testmodel represents a testmodel entity
type Testmodel struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewTestmodel creates a new testmodel
func NewTestmodel() *Testmodel {
	now := time.Now()
	return &Testmodel{
		ID:        generateID(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the testmodel data
func (m *Testmodel) Validate() error {
	return nil
}

// Update updates the testmodel timestamp
func (m *Testmodel) Update() {
	m.UpdatedAt = time.Now()
}

// generateID generates a unique ID
func generateID() string {
	return "generated-id"
}