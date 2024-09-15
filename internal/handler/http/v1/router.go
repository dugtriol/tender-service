package v1

import (
	"context"
	"log/slog"
	"net/http"

	"tender-service/internal/service"
	mwLogger "tender-service/pkg/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const api = "/api"

func NewRouter(ctx context.Context, log *slog.Logger, route *chi.Mux, services *service.Services) {
	route.Use(middleware.Logger)
	route.Use(middleware.RequestID)
	route.Use(middleware.Recoverer)
	route.Use(middleware.URLFormat)
	route.Use(mwLogger.New(log))
	route.Use(render.SetContentType(render.ContentTypeJSON))

	route.Route(
		api, func(r chi.Router) {
			r.Get("/ping", Ping())
			newUserRoutes(ctx, log, r, services.User)
			newOrgRoutes(ctx, log, r, services.Organization)
			newOrgRespRoutes(ctx, log, r, services.OrgResponsible)
			newTenderRoutes(ctx, log, r, services.User, services.Tender, services.OrgResponsible)
			newBidRoutes(ctx, log, r, services.User, services.Tender, services.OrgResponsible, services.Bid)
		},
	)
}

func Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok"))

	}
}
