package v1

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"tender-service/internal/entity"
	"tender-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const (
	bidPath        = "/bids"
	usernameMethod = "user"
	idMethod       = "id"
)

type bidRoutes struct {
	userService    service.User
	tenderService  service.Tender
	orgResponsible service.OrgResponsible
	bidService     service.Bid
}

func newBidRoutes(
	ctx context.Context, log *slog.Logger, route chi.Router, userService service.User, tenderService service.Tender,
	orgResponsible service.OrgResponsible, bidService service.Bid,
) {
	u := bidRoutes{
		userService: userService, tenderService: tenderService, orgResponsible: orgResponsible, bidService: bidService,
	}
	route.Route(
		bidPath, func(r chi.Router) {
			r.Post("/new", u.create(ctx, log))
			r.Get("/my", u.getMy(ctx, log))
			r.Get("/{tenderId}/list", u.getList(ctx, log))
			r.Get("/{bidId}/status", u.getStatus(ctx, log))
			r.Put("/{bidId}/status", u.setStatus(ctx, log))
			r.Patch("/{bidId}/edit", u.edit(ctx, log))
		},
	)
}

// IsUserOrgResponsible проверка является ли пользователь ответственный за организацию
func (u *bidRoutes) IsUserOrgResponsible(
	w http.ResponseWriter, r *http.Request, err error, ctx context.Context, log *slog.Logger,
	organizationId, userId string,
) (error, bool) {
	if _, err = u.orgResponsible.GetByIds(
		ctx, log, service.OrgResponsibleGetByIdsInput{
			OrganizationId: organizationId,
			UserId:         userId,
		},
	); err != nil {
		newErrorResponse(w, r, log, err, http.StatusForbidden, MsgForbidden)
		return nil, true
	}
	return err, false
}

// IsExistUser существует ли пользователь
func (u *bidRoutes) IsExistUser(
	w http.ResponseWriter, r *http.Request, err error, ctx context.Context, log *slog.Logger, data string,
	method string,
) (entity.User, error, bool) {
	var user entity.User
	switch method {
	case idMethod:
		user, err = u.userService.GetById(ctx, log, service.UserGetByIdInput{Id: data})
	case usernameMethod:
		user, err = u.userService.GetByUsername(
			ctx,
			log,
			service.UserGetByUsernameInput{Username: data},
		)
	}
	if err != nil {
		if err == service.ErrUserNotFound {
			newErrorResponse(w, r, log, err, http.StatusUnauthorized, MsgUserNotFound)
			return entity.User{}, nil, true
		}
		newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
		return entity.User{}, nil, true
	}
	return user, err, false
}

type inputBidCreate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	TenderId    string `json:"tenderId" validate:"required,uuid"`
	AuthorType  string `json:"authorType" validate:"required,oneof=User Organization"`
	AuthorId    string `json:"authorId" validate:"required,uuid"`
}

type bidOutput struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	TenderId    string    `json:"tenderId"`
	AuthorType  string    `json:"authorType"`
	AuthorId    string    `json:"authorId"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (u *bidRoutes) create(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input inputBidCreate
		var err error
		var done bool

		if err = render.DecodeJSON(r.Body, &input); err != nil {
			newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgFailedParsing)
			return
		}
		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		_, err, done = u.IsExistUser(w, r, err, ctx, log, input.AuthorId, idMethod)
		if done {
			return
		}

		// создание предложения
		var res entity.Bid
		if res, err = u.bidService.Create(
			ctx, log, service.BidCreateInput{
				Name:        input.Name,
				Description: input.Description,
				TenderId:    input.TenderId,
				AuthorType:  input.AuthorType,
				AuthorId:    input.AuthorId,
			},
		); err != nil {
			if err == service.ErrBidAlreadyExists {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgBidAlreadyExists)
				return
			}
			newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
			return
		}

		if _, err = u.tenderService.GetById(ctx, log, input.TenderId); err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}

		output := bidOutput{
			Id:          res.AuthorId,
			Name:        res.Name,
			Description: res.Description,
			Status:      res.Status,
			TenderId:    res.TenderId,
			AuthorType:  res.AuthorType,
			AuthorId:    res.AuthorId,
			Version:     res.Version,
			CreatedAt:   res.CreatedAt,
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}

type inputBidGetMy struct {
	Limit    int    `validate:"omitempty,number,gte=0,lte=50"`
	Offset   int    `validate:"omitempty,number,gte=0"`
	Username string `validate:"required"`
}

func (u *bidRoutes) getMy(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			limit    int
			offset   int
			username string
			err      error
		)
		l := r.URL.Query().Get("limit")
		if len(l) == 0 {
			limit = 0
		} else {
			if limit, err = strconv.Atoi(l); err != nil {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgInvalidReq)
				return
			}
		}

		off := r.URL.Query().Get("offset")
		if len(off) == 0 {
			offset = 0
		} else {
			if offset, err = strconv.Atoi(off); err != nil {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgInvalidReq)
				return
			}
		}

		username = r.URL.Query().Get("username")
		if len(username) == 0 {
			username = ""
		}

		input := inputBidGetMy{
			Limit:    limit,
			Offset:   offset,
			Username: username,
		}

		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		user, err, done := u.IsExistUser(w, r, err, ctx, log, input.Username, usernameMethod)
		if done {
			return
		}

		var bids []entity.Bid
		if bids, err = u.bidService.GetMy(
			ctx, log, service.BidGetMyInput{
				Limit:  input.Limit,
				Offset: input.Offset,
				UserId: user.Id,
			},
		); err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}
		var output []bidOutput
		for _, t := range bids {
			out := bidOutput{
				Id:          t.Id,
				Name:        t.Name,
				Description: t.Description,
				Status:      t.Status,
				TenderId:    t.TenderId,
				AuthorType:  t.AuthorType,
				AuthorId:    t.AuthorId,
				Version:     t.Version,
				CreatedAt:   t.CreatedAt,
			}
			output = append(output, out)
		}
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}

type inputBidGetList struct {
	TenderId string `validate:"required,uuid"`
	Limit    int    `validate:"omitempty,number,gte=0,lte=50"`
	Offset   int    `validate:"omitempty,number,gte=0"`
	Username string `validate:"required"`
}

func (u *bidRoutes) getList(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			limit    int
			offset   int
			username string
			tenderId string
			err      error
		)
		tenderId = chi.URLParam(r, "tenderId")

		l := r.URL.Query().Get("limit")
		if len(l) == 0 {
			limit = 0
		} else {
			if limit, err = strconv.Atoi(l); err != nil {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgInvalidReq)
				return
			}
		}

		off := r.URL.Query().Get("offset")
		if len(off) == 0 {
			offset = 0
		} else {
			if offset, err = strconv.Atoi(off); err != nil {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgInvalidReq)
				return
			}
		}

		username = r.URL.Query().Get("username")
		if len(username) == 0 {
			username = ""
		}

		input := inputBidGetList{
			TenderId: tenderId,
			Limit:    limit,
			Offset:   offset,
			Username: username,
		}

		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		user, err, done := u.IsExistUser(w, r, err, ctx, log, input.Username, usernameMethod)
		if done {
			return
		}

		var bids []entity.Bid
		if bids, err = u.bidService.GetByTenderId(
			ctx, log, service.BidGetByTenderIdInput{
				Limit:    input.Limit,
				Offset:   input.Offset,
				UserId:   user.Id,
				TenderId: tenderId,
			},
		); err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}
		var output []bidOutput
		for _, t := range bids {
			out := bidOutput{
				Id:          t.Id,
				Name:        t.Name,
				Description: t.Description,
				Status:      t.Status,
				TenderId:    t.TenderId,
				AuthorType:  t.AuthorType,
				AuthorId:    t.AuthorId,
				Version:     t.Version,
				CreatedAt:   t.CreatedAt,
			}
			output = append(output, out)
		}
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}

type inputBidGetStatus struct {
	BidId    string `validate:"required,uuid"`
	Username string `validate:"required"`
}

func (u *bidRoutes) getStatus(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err    error
			output entity.Bid
			user   entity.User
			done   bool
		)

		input := inputBidGetStatus{
			BidId:    chi.URLParam(r, "bidId"),
			Username: r.URL.Query().Get("username"),
		}

		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		user, err, done = u.IsExistUser(w, r, err, ctx, log, input.Username, usernameMethod)
		if done {
			return
		}

		output, err = u.bidService.GetById(ctx, log, input.BidId)
		if err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgBidNotFound)
			return
		}

		t, err := u.tenderService.GetById(ctx, log, output.TenderId)
		if err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}

		err, done = u.IsUserOrgResponsible(w, r, err, ctx, log, t.OrganizationId, user.Id)
		if done {
			return
		}

		type status struct {
			Status string `json:"status"`
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, status{Status: output.Status})
	}
}

type bidSetStatusInput struct {
	BidId    string `validate:"required,uuid"`
	Status   string `validate:"required,oneof=Created Published Closed"`
	Username string `validate:"required"`
}

func (u *bidRoutes) setStatus(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err  error
			out  entity.Bid
			user entity.User
			done bool
		)

		input := bidSetStatusInput{
			BidId:    chi.URLParam(r, "bidId"),
			Status:   r.URL.Query().Get("status"),
			Username: r.URL.Query().Get("username"),
		}

		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		user, err, done = u.IsExistUser(w, r, err, ctx, log, input.Username, usernameMethod)
		if done {
			return
		}

		var b entity.Bid
		if b, err = u.bidService.GetById(ctx, log, input.BidId); err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}

		t, err := u.tenderService.GetById(ctx, log, b.TenderId)
		if err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}

		err, done = u.IsUserOrgResponsible(w, r, err, ctx, log, t.OrganizationId, user.Id)
		if done {
			return
		}

		if out, err = u.bidService.PutStatus(ctx, log, input.BidId, input.Status); err != nil {
			newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
			return
		}

		output := bidOutput{
			Id:          out.Id,
			Name:        out.Name,
			Description: out.Description,
			Status:      out.Status,
			TenderId:    out.TenderId,
			AuthorType:  out.AuthorType,
			AuthorId:    out.AuthorId,
			Version:     out.Version,
			CreatedAt:   out.CreatedAt,
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}

type editParamsInputBid struct {
	BidId    string `validate:"required,uuid"`
	Username string `validate:"required"`
}

type editBodyInputBid struct {
	Name        string `json:"name" validate:"omitempty"`
	Description string `json:"description" validate:"omitempty"`
}

func (u *bidRoutes) edit(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info(fmt.Sprintf("edit(ctx context.Context, log *slog.Logger) http.HandlerFunc"))
		var (
			err  error
			out  entity.Bid
			done bool
		)

		inputParams := editParamsInputBid{
			BidId:    chi.URLParam(r, "bidId"),
			Username: r.URL.Query().Get("username"),
		}

		if err = validator.New().Struct(inputParams); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		var inputBody editBodyInputBid
		if err = render.DecodeJSON(r.Body, &inputBody); err != nil {
			newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgFailedParsing)
			return
		}

		if err = validator.New().Struct(inputBody); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		_, err, done = u.IsExistUser(w, r, err, ctx, log, inputParams.Username, usernameMethod)
		if done {
			return
		}

		if out, err = u.bidService.EditBid(
			ctx,
			log, service.BidEditInput{
				Name:        inputBody.Name,
				Description: inputBody.Description,
			}, inputParams.BidId,
		); err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgBidNotFound)
			return
		}

		output := bidOutput{
			Id:          out.Id,
			Name:        out.Name,
			Description: out.Description,
			Status:      out.Status,
			TenderId:    out.TenderId,
			AuthorType:  out.AuthorType,
			AuthorId:    out.AuthorId,
			Version:     out.Version,
			CreatedAt:   out.CreatedAt,
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}
