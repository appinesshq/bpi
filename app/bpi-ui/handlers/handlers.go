// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/appinesshq/bpi/business/mid"
	"github.com/appinesshq/bpi/foundation/web"
	"github.com/pkg/errors"
)

// Auth holds authentication details
type Auth struct {
	Token   string
	Expires time.Time
}

// API constructs an http.Handler with all application routes defined.
func API(build string, index string, apiHost string, shutdown chan os.Signal, log *log.Logger) (http.Handler, error) {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	auth := Auth{}

	root, err := newIndex("assets/views/"+index, apiHost, &auth)
	if err != nil {
		return nil, errors.Wrap(err, "setup root")
	}
	app.Handle(http.MethodGet, "/", root.handler)

	// Register the assets.
	fs := http.FileServer(http.Dir("assets"))
	fs = http.StripPrefix("/assets/", fs)
	f := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		fs.ServeHTTP(w, r)
		return nil
	}
	app.Handle(http.MethodGet, "/assets/*", f)

	// Register debug check endpoints.
	cg := checkGroup{
		build: build,
	}
	app.HandleDebug(http.MethodGet, "/readiness", cg.readiness)
	app.HandleDebug(http.MethodGet, "/liveness", cg.liveness)

	admin, err := newIndex("assets/views/admin.tmpl", apiHost, &auth)
	if err != nil {
		return nil, errors.Wrap(err, "setup admin")
	}
	app.Handle(http.MethodGet, "/admin", admin.handler)

	return app, nil
}
