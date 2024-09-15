package v1

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"tender-service/internal/service"
)

const (
	orgRespString = "/orgresp"
)

type orgRespRoutes struct {
	orgRespService service.OrgResponsible
}

func newOrgRespRoutes(
	ctx context.Context, log *slog.Logger, route chi.Router, orgRespService service.OrgResponsible,
) {
	o := orgRespRoutes{orgRespService: orgRespService}
	route.Route(
		orgRespString, func(r chi.Router) {
			r.Post("/create", o.create(ctx, log))
			r.Get("/{id}", o.get(ctx, log))
		},
	)
}

type inputOrgRespCreate struct {
	OrganizationId string `json:"organization_id" validate:"uuid"`
	UserId         string `json:"user_id" validate:"uuid"`
}

type outputOrgRespCreate struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organization_id"`
	UserId         string `json:"user_id"`
}

func (o *orgRespRoutes) create(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input inputOrgRespCreate
		var err error

		if err = render.DecodeJSON(r.Body, &input); err != nil {
			newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgFailedParsing)
			return
		}
		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		result, err := o.orgRespService.Create(
			ctx, log, service.OrgResponsibleCreateInput{
				OrganizationId: input.OrganizationId,
				UserId:         input.UserId,
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
		output := outputOrgRespCreate{
			Id:             result.Id,
			OrganizationId: result.OrganizationId,
			UserId:         result.UserId,
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}

type inputOrgRespGet struct {
	Id string `validate:"uuid"`
}

type outputOrgRespGet struct {
	Id             string `json:"id"`
	OrganizationId string `json:"organization_id"`
	UserId         string `json:"user_id"`
}

func (o *orgRespRoutes) get(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id string
		var err error

		id = chi.URLParam(r, "id")
		if err = validator.New().Struct(inputOrgRespGet{Id: id}); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		result, err := o.orgRespService.Get(ctx, log, service.OrgResponsibleGetInput{Id: id})
		if err != nil {
			if err == service.ErrOrgNotFound {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgOrgRespNotFound)
				return
			}
			newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
			return
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(
			w, r, &outputOrgRespGet{
				Id:             result.Id,
				OrganizationId: result.OrganizationId,
				UserId:         result.UserId,
			},
		)
	}
}
