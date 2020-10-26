// Package category contains category related CRUD functionality.
package category

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/foundation/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/trace"
)

var (
	// ErrNotFound is used when a specific Category is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

// Category manages the set of API's for category access.
type Category struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a Category for api access.
func New(log *log.Logger, db *sqlx.DB) Category {
	return Category{
		log: log,
		db:  db,
	}
}

// Create adds a Category to the database. It returns the created Category with
// fields like ID and DateCreated populated.
func (p Category) Create(ctx context.Context, traceID string, claims auth.Claims, n NewCategory, now time.Time) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.category.create")
	defer span.End()

	pid := sql.NullString{}
	if n.ParentID != "" {
		pid.String = n.ParentID
		pid.Valid = true
	}
	cat := Info{
		ID:          uuid.New().String(),
		Slug:        n.Slug,
		Name:        n.Name,
		ParentID:    pid,
		UserID:      claims.Subject,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}
	// if cat.ParentID == "" {
	// 	cat.ParentID = "00000000-0000-0000-0000-000000000000"
	// }

	const q = `
	INSERT INTO categories
		(category_id, slug, name, user_id, parent_id, date_created, date_updated)
	VALUES
		($1, $2, $3, $4, $5, $6, $7)`

	p.log.Printf("%s: %s: %s", traceID, "category.Create",
		database.Log(q, cat.ID, cat.Slug, cat.Name, cat.UserID, cat.ParentID, cat.DateCreated, cat.DateUpdated),
	)

	if _, err := p.db.ExecContext(ctx, q, cat.ID, cat.Slug, cat.Name, cat.UserID, cat.ParentID, cat.DateCreated, cat.DateUpdated); err != nil {
		return Info{}, errors.Wrap(err, "inserting category")
	}

	return cat, nil
}

// Update modifies data about a Category. It will error if the specified ID is
// invalid or does not reference an existing Category.
func (p Category) Update(ctx context.Context, traceID string, claims auth.Claims, id string, up UpdateCategory, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.category.update")
	defer span.End()

	cat, err := p.QueryByID(ctx, traceID, id)
	if err != nil {
		return err
	}

	// If you are not an admin and looking to retrieve someone elses category.
	if !claims.Authorized(auth.RoleAdmin) && cat.UserID != claims.Subject {
		return ErrForbidden
	}

	if up.Slug != nil {
		cat.Slug = *up.Slug
	}
	if up.Name != nil {
		cat.Name = *up.Name
	}
	if up.ParentID != nil {
		pid := sql.NullString{}
		if *up.ParentID != "" {
			pid.String = *up.ParentID
			pid.Valid = true
		}
		cat.ParentID = pid
		// if cat.ParentID == "" {
		// 	cat.ParentID = "00000000-0000-0000-0000-000000000000"
		// }
	}
	cat.DateUpdated = now

	const q = `
	UPDATE
		categories
	SET
		"slug" = $2,
		"name" = $3,
		"parent_id" = $4,
		"date_updated" = $5
	WHERE
		category_id = $1`

	p.log.Printf("%s: %s: %s", traceID, "category.Update",
		database.Log(q, id, cat.Slug, cat.Name, cat.ParentID, cat.DateUpdated),
	)

	if _, err = p.db.ExecContext(ctx, q, id, cat.Slug, cat.Name, cat.ParentID, cat.DateUpdated); err != nil {
		return errors.Wrap(err, "updating category")
	}

	return nil
}

// Delete removes the category identified by a given ID.
func (p Category) Delete(ctx context.Context, traceID string, id string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.category.delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `
	DELETE FROM
		categories
	WHERE
		category_id = $1`

	p.log.Printf("%s: %s: %s", traceID, "category.Delete",
		database.Log(q, id),
	)

	if _, err := p.db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting category %s", id)
	}

	return nil
}

// Query gets all Categories from the database.
func (p Category) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.category.query")
	defer span.End()

	const q = `
	SELECT
		*
	FROM
		categories
	OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY`
	offset := (pageNumber - 1) * rowsPerPage

	p.log.Printf("%s: %s: %s", traceID, "category.Query",
		database.Log(q, offset, rowsPerPage),
	)

	categories := []Info{}
	if err := p.db.SelectContext(ctx, &categories, q, offset, rowsPerPage); err != nil {
		return nil, errors.Wrap(err, "selecting categories")
	}

	return categories, nil
}

// QueryByID finds the category identified by a given ID.
func (p Category) QueryByID(ctx context.Context, traceID string, id string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.category.querybyid")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return Info{}, ErrInvalidID
	}

	const q = `
	SELECT
		*
	FROM
		categories
	WHERE
		category_id = $1
	GROUP BY
		category_id`

	p.log.Printf("%s: %s: %s", traceID, "category.QueryByID",
		database.Log(q, id),
	)

	var cat Info
	if err := p.db.GetContext(ctx, &cat, q, id); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrap(err, "selecting single category")
	}

	return cat, nil
}
