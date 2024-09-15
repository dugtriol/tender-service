package v1

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"tender-service/internal/service"
)

const (
	orgString = "/org"
)

type orgRoutes struct {
	orgService service.Organization
}

func newOrgRoutes(ctx context.Context, log *slog.Logger, route chi.Router, orgService service.Organization) {
	o := orgRoutes{orgService: orgService}
	route.Route(
		orgString, func(r chi.Router) {
			r.Post("/create", o.create(ctx, log))
			r.Get("/{id}", o.get(ctx, log))
		},
	)
}

type inputOrgCreate struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	OrganizationType string `json:"type" validate:"oneof=IE LLC JSC"`
}

type outputOrgCreate struct {
	Id               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	OrganizationType string `json:"type"`
}

func (o *orgRoutes) create(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input inputOrgCreate
		var err error

		if err = render.DecodeJSON(r.Body, &input); err != nil {
			newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgFailedParsing)
			return
		}
		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		result, err := o.orgService.Create(
			ctx, log, service.OrganizationCreateInput{
				Name:             input.Name,
				Description:      input.Description,
				OrganizationType: input.OrganizationType,
			},
		)
		if err != nil {
			if err == service.ErrOrgAlreadyExists {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgOrgAlreadyExists)
				return
			}
			newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
			return
		}
		output := &outputOrgCreate{
			Id:               result.Id,
			Name:             result.Name,
			Description:      result.Description,
			OrganizationType: result.OrganizationType,
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}

type inputOrgGet struct {
	Id string `validate:"uuid"`
}

type outputOrgGet struct {
	Id               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	OrganizationType string    `json:"type"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (o *orgRoutes) get(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id string
		var err error

		id = chi.URLParam(r, "id")
		if err = validator.New().Struct(inputOrgGet{Id: id}); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		result, err := o.orgService.Get(ctx, log, service.OrganizationGetInput{Id: id})
		if err != nil {
			if err == service.ErrOrgNotFound {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgOrgNotFound)
				return
			}
			newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
			return
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(
			w, r, &outputOrgGet{
				Id:               result.Id,
				Name:             result.Name,
				Description:      result.Description,
				OrganizationType: result.OrganizationType,
				CreatedAt:        result.CreatedAt,
				UpdatedAt:        result.UpdatedAt,
			},
		)
	}
}
