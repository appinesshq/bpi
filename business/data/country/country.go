// Package country contains country related CRUD functionality.
package country

import (
	"context"
	"database/sql"
	"log"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/foundation/database"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/trace"
)

var (
	// ErrNotFound is used when a specific country is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

// Country manages the set of API's for country access.
type Country struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a Country for api access.
func New(log *log.Logger, db *sqlx.DB) Country {
	return Country{
		log: log,
		db:  db,
	}
}

// Activate activates a country in the database.
func (c Country) ToggleActivate(ctx context.Context, traceID string, claims auth.Claims, countryCode string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.country.activate")
	defer span.End()

	// Only admins should be able to do this.
	if !claims.Authorized(auth.RoleAdmin) {
		return ErrForbidden
	}

	q := `
	SELECT
		*
	FROM
		countries
	WHERE 
		code = $1`

	c.log.Printf("%s: %s: %s", traceID, "country.ToggleActivate",
		database.Log(q, countryCode),
	)

	var country Info
	if err := c.db.GetContext(ctx, &country, q, countryCode); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return errors.Wrapf(err, "selecting country %q", countryCode)
	}

	q = `
	UPDATE
		countries
	SET 
		"active" = $2
	WHERE
		code = $1`

	c.log.Printf("%s: %s: %s", traceID, "country.Activate",
		database.Log(q, countryCode, !country.Active),
	)

	if _, err := c.db.ExecContext(ctx, q, countryCode, !country.Active); err != nil {
		return errors.Wrap(err, "updating country")
	}

	return nil
}

// Query retrieves a list of active countries from the database.
func (c Country) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.country.query")
	defer span.End()

	const q = `
	SELECT
		*
	FROM
		countries
	WHERE 
		active
	ORDER BY
		code
	OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY`

	offset := (pageNumber - 1) * rowsPerPage

	c.log.Printf("%s: %s: %s", traceID, "country.Query",
		database.Log(q, offset, rowsPerPage),
	)

	countries := []Info{}
	if err := c.db.SelectContext(ctx, &countries, q, offset, rowsPerPage); err != nil {
		return nil, errors.Wrap(err, "selecting countries")
	}

	return countries, nil
}

// QueryByID gets the specified country from the database.
func (c Country) QueryByID(ctx context.Context, traceID string, claims auth.Claims, countryID int) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.country.querybyid")
	defer span.End()

	const q = `
	SELECT
		*
	FROM
		countries
	WHERE 
		country_id = $1 AND active`

	c.log.Printf("%s: %s: %s", traceID, "country.QueryByID",
		database.Log(q, countryID),
	)

	var country Info
	if err := c.db.GetContext(ctx, &country, q, countryID); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting country %q", countryID)
	}

	return country, nil
}

// QueryByCountryCode gets the specified country from the database.
func (c Country) QueryByCode(ctx context.Context, traceID string, claims auth.Claims, countryCode string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.country.querybycode")
	defer span.End()

	const q = `
	SELECT
		*
	FROM
		countries
	WHERE 
		code = $1 AND active`

	c.log.Printf("%s: %s: %s", traceID, "country.QueryByCode",
		database.Log(q, countryCode),
	)

	var country Info
	if err := c.db.GetContext(ctx, &country, q, countryCode); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting country %q", countryCode)
	}

	return country, nil
}
