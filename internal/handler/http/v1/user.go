package v1

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"tender-service/internal/service"
)

const (
	userString = "/user"
)

type userRoutes struct {
	userService service.User
}

func newUserRoutes(ctx context.Context, log *slog.Logger, route chi.Router, userService service.User) {
	u := userRoutes{userService: userService}
	route.Route(
		userString, func(r chi.Router) {
			r.Post("/create", u.create(ctx, log))
			r.Get("/{id}", u.get(ctx, log))
		},
	)
}

type inputUserCreate struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (u *userRoutes) create(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input inputUserCreate
		var err error

		if err = render.DecodeJSON(r.Body, &input); err != nil {
			newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgFailedParsing)
			return
		}
		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}
		id, err := u.userService.Create(
			ctx, log, service.UserCreateInput{
				Username:  input.Username,
				FirstName: input.FirstName,
				LastName:  input.LastName,
			},
		)
		if err != nil {
			if err == service.ErrUserAlreadyExists {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgUserAlreadyExists)
				return
			}
			newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
			return
		}

		type response struct {
			Id string `json:"id"`
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response{Id: id})
	}
}

type inputUserGet struct {
	Id string `validate:"uuid"`
}

func (u *userRoutes) get(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id string
		var err error

		id = chi.URLParam(r, "id")
		if err = validator.New().Struct(inputUserGet{Id: id}); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}
		log.Info(fmt.Sprintf("Handler - User - Create - validate is ok"))
		user, err := u.userService.GetById(ctx, log, service.UserGetByIdInput{Id: id})
		if err != nil {
			if err == service.ErrUserNotFound {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgUserNotFound)
				return
			}
			newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
			return
		}

		type userResp struct {
			Id        string `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		w.WriteHeader(http.StatusOK)
		render.JSON(
			w, r, &userResp{
				Id:        user.Id,
				Username:  user.Username,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			},
		)
	}
}
