package models

import "time"

// Model is a struct containing the basic fields for a persisted entity defined
// in the API.
type Model struct {
	// Id of the entity
	ID uint `json:"id"`
	// Created timestamp. Populated when the object is saved to the db.
	CreatedAt time.Time `json:"created_at"`
	// Last updated timestamp. Updated when the object is updated in the db.
	UpdatedAt time.Time `json:"updated_at"`
}
