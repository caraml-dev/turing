package models

import "time"

type ID uint

func NewID(id int) *ID {
	casted := ID(id)
	return &casted
}

// Model is a struct containing the basic fields for a persisted entity defined
// in the API.
type Model struct {
	// Id of the entity
	ID ID `json:"id"`
	// Created timestamp. Populated when the object is saved to the db.
	CreatedAt time.Time `json:"created_at"`
	// Last updated timestamp. Updated when the object is updated in the db.
	UpdatedAt time.Time `json:"updated_at"`
}

func (m Model) GetID() ID {
	return m.ID
}

func (m Model) GetCreatedAt() time.Time {
	return m.CreatedAt
}

func (m Model) GetUpdatedAt() time.Time {
	return m.UpdatedAt
}
