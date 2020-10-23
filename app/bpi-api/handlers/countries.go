package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/business/data/country"
	"github.com/appinesshq/bpi/foundation/web"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/trace"
)

type countryGroup struct {
	country country.Country
}

func (cg countryGroup) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.countryGroup.query")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	params := web.Params(r)
	pageNumber, err := strconv.Atoi(params["page"])
	if err != nil {
		return web.NewRequestError(fmt.Errorf("invalid page format: %s", params["page"]), http.StatusBadRequest)
	}
	rowsPerPage, err := strconv.Atoi(params["rows"])
	if err != nil {
		return web.NewRequestError(fmt.Errorf("invalid rows format: %s", params["rows"]), http.StatusBadRequest)
	}

	countries, err := cg.country.Query(ctx, v.TraceID, pageNumber, rowsPerPage)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, countries, http.StatusOK)
}

// func (cg countryGroup) queryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
// 	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.countryGroup.queryByID")
// 	defer span.End()

// 	params := web.Params(r)
// 	id, err := strconv.Atoi(params["id"])
// 	if err != nil {
// 		return web.NewRequestError(err, http.StatusBadRequest)
// 	}

// 	v, ok := ctx.Value(web.KeyValues).(*web.Values)
// 	if !ok {
// 		return web.NewShutdownError("web value missing from context")
// 	}

// 	claims, ok := ctx.Value(auth.Key).(auth.Claims)
// 	if !ok {
// 		return web.NewShutdownError("claims missing from context")
// 	}

// 	c, err := cg.country.QueryByID(ctx, v.TraceID, claims, id)
// 	if err != nil {
// 		switch err {
// 		case country.ErrInvalidID:
// 			return web.NewRequestError(err, http.StatusBadRequest)
// 		case country.ErrNotFound:
// 			return web.NewRequestError(err, http.StatusNotFound)
// 		default:
// 			return errors.Wrapf(err, "ID: %s", params["id"])
// 		}
// 	}

// 	return web.Respond(ctx, w, c, http.StatusOK)
// }

func (cg countryGroup) queryByCode(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.countryGroup.queryByCode")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	params := web.Params(r)
	c, err := cg.country.QueryByCode(ctx, v.TraceID, claims, params["cc"])
	if err != nil {
		switch err {
		case country.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case country.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "CC: %s", params["cc"])
		}
	}

	return web.Respond(ctx, w, c, http.StatusOK)
}

func (cg countryGroup) toggleActive(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.countryGroup.toggleActive")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	params := web.Params(r)
	if err := cg.country.ToggleActive(ctx, v.TraceID, claims, params["cc"]); err != nil {
		switch err {
		case country.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case country.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case country.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "CC: %s", params["cc"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
