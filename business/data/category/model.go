package category

import (
	"database/sql"
	"time"
)

// Info represents an individual category.
type Info struct {
	ID          string         `db:"category_id" json:"id"`                // Unique identifier.
	Slug        string         `db:"slug" json:"slug"`                     // Unique category name
	Name        string         `db:"name" json:"name"`                     // Display name of the category.
	UserID      string         `db:"user_id" json:"user_id"`               // User ID of the category owner.
	ParentID    sql.NullString `db:"parent_id" json:"parent_id,omitempty"` // Parent category ID.
	DateCreated time.Time      `db:"date_created" json:"date_created"`     // When the category was added.
	DateUpdated time.Time      `db:"date_updated" json:"date_updated"`     // When the category record was last modified.
}

// NewCategory is what we require from clients when adding a Category.
type NewCategory struct {
	Slug     string `json:"slug" validate:"required"` // Unique category name
	Name     string `json:"name" validate:"required"` // Display name of the category.
	ParentID string `json:"parent_id"`                // Parent category ID.
}

// UpdateCategory defines what information may be provided to modify an
// existing Category. All fields are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and a field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateCategory struct {
	Slug     *string `json:"slug"`      // Unique category slug
	Name     *string `json:"name"`      // Display name of the category.
	ParentID *string `json:"parent_id"` // Parent category ID.
}
