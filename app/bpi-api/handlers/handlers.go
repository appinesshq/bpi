// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/appinesshq/bpi/business/auth" // Import is removed in final PR
	"github.com/appinesshq/bpi/business/data/category"
	"github.com/appinesshq/bpi/business/data/country"
	"github.com/appinesshq/bpi/business/data/jurisdiction"
	"github.com/appinesshq/bpi/business/data/product"
	"github.com/appinesshq/bpi/business/data/profile"
	"github.com/appinesshq/bpi/business/data/user"
	"github.com/appinesshq/bpi/business/mid"
	"github.com/appinesshq/bpi/foundation/web"
	"github.com/jmoiron/sqlx"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, a *auth.Auth, db *sqlx.DB) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	// Register debug check endpoints.
	cg := checkGroup{
		build: build,
		db:    db,
	}
	app.HandleDebug(http.MethodGet, "/readiness", cg.readiness)
	app.HandleDebug(http.MethodGet, "/liveness", cg.liveness)

	// Register user management and authentication endpoints.
	ug := userGroup{
		user: user.New(log, db),
		auth: a,
	}
	app.Handle(http.MethodGet, "/v1/users/:page/:rows", ug.query, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodGet, "/v1/users/token/:kid", ug.token)
	app.Handle(http.MethodGet, "/v1/users/:id", ug.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/users", ug.create, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodPut, "/v1/users/:id", ug.update, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))
	app.Handle(http.MethodDelete, "/v1/users/:id", ug.delete, mid.Authenticate(a), mid.Authorize(auth.RoleAdmin))

	// Register product and sale endpoints.
	pg := productGroup{
		product: product.New(log, db),
	}
	app.Handle(http.MethodGet, "/v1/products/:page/:rows", pg.query, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/products/:id", pg.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/products", pg.create, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/products/:id", pg.update, mid.Authenticate(a))
	app.Handle(http.MethodDelete, "/v1/products/:id", pg.delete, mid.Authenticate(a))

	// Register country endpoints.
	cog := countryGroup{
		country: country.New(log, db),
	}
	app.Handle(http.MethodGet, "/v1/countries/:page/:rows", cog.query, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/countries/:cc", cog.queryByCode, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/countries/:cc", cog.toggleActive, mid.Authenticate(a))

	// Register jurisdiction endpoints.
	jg := jurisdictionGroup{
		jurisdiction: jurisdiction.New(log, db),
	}
	app.Handle(http.MethodGet, "/v1/jurisdictions/:page/:rows", jg.query, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/jurisdictions/:cc", jg.queryByCode, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/jurisdictions/:cc", jg.toggleActive, mid.Authenticate(a))

	// Register profile endpoints.
	prg := profileGroup{
		profile: profile.New(log, db),
	}
	app.Handle(http.MethodGet, "/v1/profiles/:page/:rows", prg.query, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/profiles/:name", prg.queryByName, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/profiles", prg.create, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/profiles/:name", prg.update, mid.Authenticate(a))
	app.Handle(http.MethodDelete, "/v1/profiles/:name", prg.delete, mid.Authenticate(a))

	// User profile
	app.Handle(http.MethodGet, "/v1/users/profile/user/:id", prg.QueryUserProfile, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/users/:id/profile", prg.QueryUserProfile, mid.Authenticate(a))

	// Register category endpoints.
	cag := categoryGroup{
		category: category.New(log, db),
	}
	app.Handle(http.MethodGet, "/v1/categories/:page/:rows", cag.query, mid.Authenticate(a))
	app.Handle(http.MethodGet, "/v1/categories/:id", cag.queryByID, mid.Authenticate(a))
	app.Handle(http.MethodPost, "/v1/categories", cag.create, mid.Authenticate(a))
	app.Handle(http.MethodPut, "/v1/categories/:id", cag.update, mid.Authenticate(a))
	app.Handle(http.MethodDelete, "/v1/categories/:id", cag.delete, mid.Authenticate(a))

	return app
}
