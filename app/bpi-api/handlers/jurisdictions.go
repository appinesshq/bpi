package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/business/data/jurisdiction"
	"github.com/appinesshq/bpi/foundation/web"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/trace"
)

type jurisdictionGroup struct {
	jurisdiction jurisdiction.Jurisdiction
}

func (jg jurisdictionGroup) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.jurisdictionGroup.query")
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

	Jurisdictions, err := jg.jurisdiction.Query(ctx, v.TraceID, pageNumber, rowsPerPage)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, Jurisdictions, http.StatusOK)
}

func (jg jurisdictionGroup) queryByCode(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.jurisdictionGroup.queryByCode")
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
	c, err := jg.jurisdiction.QueryByCode(ctx, v.TraceID, claims, params["cc"])
	if err != nil {
		switch err {
		case jurisdiction.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case jurisdiction.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "CC: %s", params["cc"])
		}
	}

	return web.Respond(ctx, w, c, http.StatusOK)
}

func (jg jurisdictionGroup) toggleActive(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.jurisdictionGroup.toggleActive")
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
	if err := jg.jurisdiction.ToggleActive(ctx, v.TraceID, claims, params["cc"]); err != nil {
		switch err {
		case jurisdiction.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case jurisdiction.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case jurisdiction.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "CC: %s", params["cc"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
