package profile

import "time"

// Info represents an individual profile.
type Info struct {
	// ID          string    `db:"profile_id" json:"id"`             // Unique identifier.
	Name        string    `db:"name" json:"name"`                 // Unique username
	DisplayName string    `db:"display_name" json:"display_name"` // Display name of the profile.
	UserID      string    `db:"user_id" json:"user_id"`           // ID of the user who created the profile.
	DateCreated time.Time `db:"date_created" json:"date_created"` // When the profile was added.
	DateUpdated time.Time `db:"date_updated" json:"date_updated"` // When the profile record was last modified.
}

// NewProfile is what we require from clients when adding a Profile.
type NewProfile struct {
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"display_name" validate:"required"`
}

// UpdateProfile defines what information may be provided to modify an
// existing Profile. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateProfile struct {
	Name        *string `json:"name"`
	DisplayName *string `json:"display_name"`
}
