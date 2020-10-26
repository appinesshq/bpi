package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/business/data/profile"
	"github.com/appinesshq/bpi/foundation/web"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/trace"
)

type profileGroup struct {
	profile profile.Profile
}

func (pg profileGroup) query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.profileGroup.query")
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

	profiles, err := pg.profile.Query(ctx, v.TraceID, pageNumber, rowsPerPage)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, profiles, http.StatusOK)
}

func (pg profileGroup) queryByName(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.profileGroup.queryByID")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	params := web.Params(r)
	prod, err := pg.profile.QueryByName(ctx, v.TraceID, params["name"])
	if err != nil {
		switch err {
		case profile.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case profile.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "ID: %s", params["name"])
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}

func (pg profileGroup) queryByUserID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.profileGroup.queryByID")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	params := web.Params(r)
	prod, err := pg.profile.QueryByUserID(ctx, v.TraceID, params["id"])
	if err != nil {
		switch err {
		case profile.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case profile.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		default:
			return errors.Wrapf(err, "user ID: %s", params["id"])
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}

func (pg profileGroup) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.profileGroup.create")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	var np profile.NewProfile
	if err := web.Decode(r, &np); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	prod, err := pg.profile.Create(ctx, v.TraceID, claims, np, v.Now)
	if err != nil {
		return errors.Wrapf(err, "creating new profile: %+v", np)
	}

	return web.Respond(ctx, w, prod, http.StatusCreated)
}

func (pg profileGroup) update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.profileGroup.update")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return web.NewShutdownError("claims missing from context")
	}

	var upd profile.UpdateProfile
	if err := web.Decode(r, &upd); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	params := web.Params(r)
	if err := pg.profile.Update(ctx, v.TraceID, claims, params["name"], upd, v.Now); err != nil {
		switch err {
		case profile.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		case profile.ErrNotFound:
			return web.NewRequestError(err, http.StatusNotFound)
		case profile.ErrForbidden:
			return web.NewRequestError(err, http.StatusForbidden)
		default:
			return errors.Wrapf(err, "Name: %s  User: %+v", params["name"], &upd)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (pg profileGroup) delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := trace.SpanFromContext(ctx).Tracer().Start(ctx, "handlers.profileGroup.delete")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	params := web.Params(r)
	if err := pg.profile.Delete(ctx, v.TraceID, params["name"]); err != nil {
		switch err {
		case profile.ErrInvalidID:
			return web.NewRequestError(err, http.StatusBadRequest)
		default:
			return errors.Wrapf(err, "Name: %s", params["name"])
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
