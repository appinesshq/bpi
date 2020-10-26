// Package profile contains profile related CRUD functionality.
package profile

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/foundation/database"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/trace"
)

var (
	// ErrNotFound is used when a specific Profile is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

// Profile manages the set of API's for profile access.
type Profile struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a Profile for api access.
func New(log *log.Logger, db *sqlx.DB) Profile {
	return Profile{
		log: log,
		db:  db,
	}
}

// Create adds a Profile to the database. It returns the created Profile with
// fields like ID and DateCreated populated.
func (p Profile) Create(ctx context.Context, traceID string, claims auth.Claims, np NewProfile, now time.Time) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.profile.create")
	defer span.End()

	t, err := TypeFromString(np.Type)
	if err != nil {
		return Info{}, errors.Wrap(err, "getting profile type")
	}

	n := Info{
		Name:        np.Name,
		DisplayName: np.DisplayName,
		Type:        *t,
		UserID:      claims.Subject,
		DateCreated: now.UTC(),
		DateUpdated: now.UTC(),
	}

	const q = `
	INSERT INTO profiles
		(name, user_id, display_name, type, date_created, date_updated)
	VALUES
		($1, $2, $3, $4, $5, $6)`

	p.log.Printf("%s: %s: %s", traceID, "profile.Create",
		database.Log(q, n.Name, n.UserID, n.DisplayName, n.Type, n.DateCreated, n.DateUpdated),
	)

	if _, err := p.db.ExecContext(ctx, q, n.Name, n.UserID, n.DisplayName, n.Type, n.DateCreated, n.DateUpdated); err != nil {
		return Info{}, errors.Wrap(err, "inserting profile")
	}

	return n, nil
}

// Update modifies data about a Profile. It will error if the specified name is
// invalid or does not reference an existing Profile.
func (p Profile) Update(ctx context.Context, traceID string, claims auth.Claims, name string, up UpdateProfile, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.profile.update")
	defer span.End()

	o, err := p.QueryByName(ctx, traceID, name)
	if err != nil {
		return err
	}

	// If you are not an admin and looking to retrieve someone elses profile.
	if !claims.Authorized(auth.RoleAdmin) && o.UserID != claims.Subject {
		return ErrForbidden
	}

	if up.Name != nil {
		o.Name = *up.Name
	}
	if up.DisplayName != nil {
		o.DisplayName = *up.DisplayName
	}
	o.DateUpdated = now

	const q = `
	UPDATE
		profiles
	SET
		"name" = $2,
		"display_name" = $3,
		"date_updated" = $4
	WHERE
		name = $1`

	p.log.Printf("%s: %s: %s", traceID, "profile.Update",
		database.Log(q, o.Name, o.Name, o.DisplayName, o.DateUpdated),
	)

	if _, err = p.db.ExecContext(ctx, q, o.Name, o.Name, o.DisplayName, o.DateUpdated); err != nil {
		return errors.Wrap(err, "updating profile")
	}

	return nil
}

// Delete removes the profile identified by a given ID.
func (p Profile) Delete(ctx context.Context, traceID string, name string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.profile.delete")
	defer span.End()

	const q = `
	DELETE FROM
		profiles
	WHERE
		name = $1`

	p.log.Printf("%s: %s: %s", traceID, "profile.Delete",
		database.Log(q, name),
	)

	if _, err := p.db.ExecContext(ctx, q, name); err != nil {
		return errors.Wrapf(err, "deleting profile %s", name)
	}

	return nil
}

// Query gets all Profiles from the database.
func (p Profile) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.profile.query")
	defer span.End()

	const q = `
	SELECT * FROM profiles
	OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY`
	offset := (pageNumber - 1) * rowsPerPage

	p.log.Printf("%s: %s: %s", traceID, "profile.Query",
		database.Log(q, offset, rowsPerPage),
	)

	o := []Info{}
	if err := p.db.SelectContext(ctx, &o, q, offset, rowsPerPage); err != nil {
		return nil, errors.Wrap(err, "selecting profiles")
	}

	return o, nil
}

// QueryByName finds the profile identified by a given name.
func (p Profile) QueryByName(ctx context.Context, traceID string, name string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.profile.querybyname")
	defer span.End()

	const q = `
	SELECT 
		* FROM profiles
	WHERE
		name = $1
	`

	p.log.Printf("%s: %s: %s", traceID, "profile.QueryByName",
		database.Log(q, name),
	)

	var o Info
	if err := p.db.GetContext(ctx, &o, q, name); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrap(err, "selecting single profile")
	}

	return o, nil
}

// QueryUserProfile finds the profile identified by a given user ID.
func (p Profile) QueryUserProfile(ctx context.Context, traceID string, userID string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.profile.QueryUserProfile")
	defer span.End()

	const q = `
	SELECT 
		* FROM profiles
	WHERE
		user_id = $1 AND type='USR'
	`

	p.log.Printf("%s: %s: %s", traceID, "profile.QueryUserProfile",
		database.Log(q, userID),
	)

	var o Info
	if err := p.db.GetContext(ctx, &o, q, userID); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrap(err, "selecting single profile")
	}

	return o, nil
}
