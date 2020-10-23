// Package jurisdiction contains jurisdiction related CRUD functionality.
package jurisdiction

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/foundation/database"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/trace"
)

var (
	// ErrNotFound is used when a specifij Jurisdiction is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")
)

// Jurisdiction manages the set of API's for jurisdiction access.
type Jurisdiction struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a Jurisdiction for api access.
func New(log *log.Logger, db *sqlx.DB) Jurisdiction {
	return Jurisdiction{
		log: log,
		db:  db,
	}
}

// ToggleActive activates a jurisdiction in the database.
func (j Jurisdiction) ToggleActive(ctx context.Context, traceID string, claims auth.Claims, jurisdictionCode string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.jurisdiction.activate")
	defer span.End()

	// Only admins should be able to do this.
	if !claims.Authorized(auth.RoleAdmin) {
		return ErrForbidden
	}

	q := `
	SELECT
		j.*
	FROM
		jurisdictions as j
	INNER JOIN 
		countries as c
	ON
		j.country_code = c.code
	WHERE 
		j.code = $1 AND c.active`

	j.log.Printf("%s: %s: %s", traceID, "jurisdiction.ToggleActivate",
		database.Log(q, jurisdictionCode),
	)

	var jurisdiction Info
	if err := j.db.GetContext(ctx, &jurisdiction, q, jurisdictionCode); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return errors.Wrapf(err, "selecting jurisdiction %q", jurisdictionCode)
	}

	q = `
	UPDATE
		jurisdictions
	SET 
		"active" = $2
	WHERE
		code = $1`

	j.log.Printf("%s: %s: %s", traceID, "jurisdiction.Activate",
		database.Log(q, jurisdictionCode, !jurisdiction.Active),
	)

	if _, err := j.db.ExecContext(ctx, q, jurisdictionCode, !jurisdiction.Active); err != nil {
		return errors.Wrap(err, "updating jurisdiction")
	}

	return nil
}

// Query retrieves a list of active jurisdictions from the database.
func (j Jurisdiction) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.jurisdiction.query")
	defer span.End()

	const q = `
	SELECT
		*
	FROM
		jurisdictions
	WHERE 
		active
	ORDER BY
		code
	OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY`

	offset := (pageNumber - 1) * rowsPerPage

	j.log.Printf("%s: %s: %s", traceID, "jurisdiction.Query",
		database.Log(q, offset, rowsPerPage),
	)

	jurisdictions := []Info{}
	if err := j.db.SelectContext(ctx, &jurisdictions, q, offset, rowsPerPage); err != nil {
		return nil, errors.Wrap(err, "selecting jurisdictions")
	}

	return jurisdictions, nil
}

// QueryByjurisdictionCode gets the specified jurisdiction from the database.
func (j Jurisdiction) QueryByCode(ctx context.Context, traceID string, claims auth.Claims, jurisdictionCode string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.jurisdiction.querybycode")
	defer span.End()

	// Check whether the ID is valid.
	p := strings.Split(jurisdictionCode, ".")
	if len(p) != 2 || len(p[0]) != 2 {
		return Info{}, ErrInvalidID
	}

	const q = `
	SELECT
		*
	FROM
		jurisdictions
	WHERE 
		code = $1 AND active`

	j.log.Printf("%s: %s: %s", traceID, "jurisdiction.QueryByCode",
		database.Log(q, jurisdictionCode),
	)

	var jurisdiction Info
	if err := j.db.GetContext(ctx, &jurisdiction, q, jurisdictionCode); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting jurisdiction %q", jurisdictionCode)
	}

	return jurisdiction, nil
}
