package models

import (
	"errors"
	"time"
)

// Testmodel represents a testmodel in the system
type Testmodel struct {
	ID        string    `json:"id"`
	
	
	
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewTestmodel creates a new Testmodel
func NewTestmodel() *Testmodel {
	return &Testmodel{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Validate validates the testmodel
func (m *Testmodel) Validate() error {
	if m.ID == "" {
		return errors.New("id is required")
	}
	
	
	
	return nil
}

// Update updates the testmodel timestamp
func (m *Testmodel) Update() {
	m.UpdatedAt = time.Now()
}