// Package user contains user related CRUD functionality.
package user

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/foundation/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/trace"
	"golang.org/x/crypto/bcrypt"
)

var Issuer = "MB Appiness Solutions"

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("authentication failed")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")

	// PasswordSalt is the salt value which will be addedd to passwords during the hashing process for extra security.
	PasswordSalt = "joireu98ytu98grHROIHGWJREOIJOIJroJ5Y09JRATHJOIHJj5y09aeoirjroiejjrtjhJROIJIJyjHJroisjh509e5e0jte0jhreoijtkjrtrej9yg"

	// EmailSalt is the salt value which will be addedd to emails during the hashing process for extra security.
	// Emails are hashed for GDPR compliance.
	EmailSalt = "nbkjvnKJNBKJNFNKFbnkfnte80bnfdb5e5090hetaoijknbnjvNKSFBfnkjneinF8I*H$%IHIGRiuhgIUNGEibus8b8s9rnbnrwengiubi4w9898U8H"
)

// HashEmail hashes the provided email address for GDPR compliance.
func hashEmail(email string) string {
	sum := sha256.Sum256([]byte(email+EmailSalt))
	return fmt.Sprintf("%x", sum)
}

// User manages the set of API's for user access.
type User struct {
	log *log.Logger
	db  *sqlx.DB
}

// New constructs a User for api access.
func New(log *log.Logger, db *sqlx.DB) User {
	return User{
		log: log,
		db:  db,
	}
}

// Create inserts a new user into the database.
func (u User) Create(ctx context.Context, traceID string, nu NewUser, now time.Time) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "internal.data.user.create")
	defer span.End()

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password+PasswordSalt), bcrypt.DefaultCost)
	if err != nil {
		return Info{}, errors.Wrap(err, "generating password hash")
	}

	usr := Info{
		ID:           uuid.New().String(),
		Email:        hashEmail(nu.Email),
		PasswordHash: hash,
		Roles:        nu.Roles,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `
	INSERT INTO users
		(user_id, email, password_hash, roles, date_created, date_updated)
	VALUES
		($1, $2, $3, $4, $5, $6)`

	u.log.Printf("%s: %s: %s", traceID, "user.Create",
		database.Log(q, usr.ID, usr.Email, usr.PasswordHash, usr.Roles, usr.DateCreated, usr.DateUpdated),
	)

	if _, err = u.db.ExecContext(ctx, q, usr.ID, usr.Email, usr.PasswordHash, usr.Roles, usr.DateCreated, usr.DateUpdated); err != nil {
		return Info{}, errors.Wrap(err, "inserting user")
	}

	return usr, nil
}

// Update replaces a user document in the database.
func (u User) Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uu UpdateUser, now time.Time) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.update")
	defer span.End()

	usr, err := u.QueryByID(ctx, traceID, claims, userID)
	if err != nil {
		return err
	}

	if uu.Email != nil {
		usr.Email = hashEmail(*uu.Email)
	}
	if uu.Roles != nil {
		usr.Roles = uu.Roles
	}
	if uu.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		usr.PasswordHash = pw
	}
	usr.DateUpdated = now

	const q = `
	UPDATE
		users
	SET 
		"email" = $2,
		"roles" = $3,
		"password_hash" = $4,
		"date_updated" = $5
	WHERE
		user_id = $1`

	u.log.Printf("%s: %s: %s", traceID, "user.Update",
		database.Log(q, usr.ID, usr.Email, usr.Roles, usr.PasswordHash, usr.DateCreated, usr.DateUpdated),
	)

	if _, err = u.db.ExecContext(ctx, q, userID, usr.Email, usr.Roles, usr.PasswordHash, usr.DateUpdated); err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

// Delete removes a user from the database.
func (u User) Delete(ctx context.Context, traceID string, userID string) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.delete")
	defer span.End()

	if _, err := uuid.Parse(userID); err != nil {
		return ErrInvalidID
	}

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = $1`

	u.log.Printf("%s: %s: %s", traceID, "user.Delete",
		database.Log(q, userID),
	)

	if _, err := u.db.ExecContext(ctx, q, userID); err != nil {
		return errors.Wrapf(err, "deleting user %s", userID)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (u User) Query(ctx context.Context, traceID string, pageNumber int, rowsPerPage int) ([]Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.query")
	defer span.End()

	const q = `
	SELECT
		*
	FROM
		users
	ORDER BY
		user_id
	OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY`

	offset := (pageNumber - 1) * rowsPerPage

	u.log.Printf("%s: %s: %s", traceID, "user.Query",
		database.Log(q, offset, rowsPerPage),
	)

	users := []Info{}
	if err := u.db.SelectContext(ctx, &users, q, offset, rowsPerPage); err != nil {
		return nil, errors.Wrap(err, "selecting users")
	}

	return users, nil
}

// QueryByID gets the specified user from the database.
func (u User) QueryByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.querybyid")
	defer span.End()

	if strings.ToLower(userID) == "me" {
		userID = claims.Subject
	}

	if _, err := uuid.Parse(userID); err != nil {
		return Info{}, ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return Info{}, ErrForbidden
	}

	const q = `
	SELECT
		*
	FROM
		users
	WHERE 
		user_id = $1`

	u.log.Printf("%s: %s: %s", traceID, "user.QueryByID",
		database.Log(q, userID),
	)

	var usr Info
	if err := u.db.GetContext(ctx, &usr, q, userID); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting user %q", userID)
	}

	return usr, nil
}

// QueryByEmail gets the specified user from the database by email.
func (u User) QueryByEmail(ctx context.Context, traceID string, claims auth.Claims, email string) (Info, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.querybyemail")
	defer span.End()

	email = hashEmail(email)

	const q = `
	SELECT
		*
	FROM
		users
	WHERE
		email = $1`

	u.log.Printf("%s: %s: %s", traceID, "user.QueryByEmail",
		database.Log(q, email),
	)

	var usr Info
	if err := u.db.GetContext(ctx, &usr, q, email); err != nil {
		if err == sql.ErrNoRows {
			return Info{}, ErrNotFound
		}
		return Info{}, errors.Wrapf(err, "selecting user %q", email)
	}

	// If you are not an admin and looking to retrieve someone other than yourself.
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != usr.ID {
		return Info{}, ErrForbidden
	}

	return usr, nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims Info representing this user. The claims can be
// used to generate a token for future authentication.
func (u User) Authenticate(ctx context.Context, traceID string, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "business.data.user.authenticate")
	defer span.End()

	email = hashEmail(email)

	const q = `
	SELECT
		*
	FROM
		users
	WHERE
		email = $1`

	u.log.Printf("%s: %s: %s", traceID, "user.Authenticate",
		database.Log(q, email),
	)

	var usr Info
	if err := u.db.GetContext(ctx, &usr, q, email); err != nil {

		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated user which emails are in the system.
		if err == sql.ErrNoRows {
			return auth.Claims{}, ErrAuthenticationFailure
		}

		return auth.Claims{}, errors.Wrap(err, "selecting single user")
	}

	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password+PasswordSalt)); err != nil {
		return auth.Claims{}, ErrAuthenticationFailure
	}

	// If we are this far the request is valid. Create some claims for the user
	// and generate their token.
	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    Issuer,
			Subject:   usr.ID,
			Audience:  "users",
			ExpiresAt: now.Add(time.Hour).Unix(),
			IssuedAt:  now.Unix(),
		},
		Roles: usr.Roles,
	}

	return claims, nil
}
