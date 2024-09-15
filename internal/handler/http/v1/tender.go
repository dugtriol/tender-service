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
	tender        = "/tenders"
	statusCreated = "Created"
	statusClosed  = "Closed"
)

type tenderRoutes struct {
	userService    service.User
	tenderService  service.Tender
	orgResponsible service.OrgResponsible
}

func newTenderRoutes(
	ctx context.Context, log *slog.Logger, route chi.Router, userService service.User, tenderService service.Tender,
	orgResponsible service.OrgResponsible,
) {
	u := tenderRoutes{userService: userService, tenderService: tenderService, orgResponsible: orgResponsible}
	route.Route(
		tender, func(r chi.Router) {
			r.Post("/new", u.create(ctx, log))
			r.Get("/", u.getByType(ctx, log))
			r.Get("/my", u.getMy(ctx, log))
			r.Get("/{tenderId}/status", u.getStatus(ctx, log))
			r.Put("/{tenderId}/status", u.setStatus(ctx, log))
			r.Patch("/{tenderId}/edit", u.edit(ctx, log))
		},
	)
}

// IsUserOrgResponsible проверка является ли пользователь ответственный за организацию
func (u *tenderRoutes) IsUserOrgResponsible(
	w http.ResponseWriter, r *http.Request, err error, ctx context.Context, log *slog.Logger,
	organizationId, userId string,
) (error, bool) {
	if _, err = u.orgResponsible.GetByIds(
		ctx, log, service.OrgResponsibleGetByIdsInput{
			OrganizationId: organizationId,
			UserId:         userId,
		},
	); err != nil {
		//if err == service.ErrOrgRespNotFound {
		//	newErrorResponse(w, r, log, err, http.StatusForbidden, MsgForbidden)
		//	return nil, true
		//}
		newErrorResponse(w, r, log, err, http.StatusForbidden, MsgForbidden)
		return nil, true
	}
	return err, false
}

// IsExistUser существует ли пользователь
func (u *tenderRoutes) IsExistUser(
	w http.ResponseWriter, r *http.Request, err error, ctx context.Context, log *slog.Logger,
	username string,
) (entity.User, error, bool) {
	log.Info(fmt.Sprintf("IsExistUser"))
	var user entity.User
	if user, err = u.userService.GetByUsername(
		ctx,
		log,
		service.UserGetByUsernameInput{Username: username},
	); err != nil {
		if err == service.ErrUserNotFound {
			newErrorResponse(w, r, log, err, http.StatusUnauthorized, MsgUserNotFound)
			return entity.User{}, nil, true
		}
		newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
		return entity.User{}, nil, true
	}
	return user, err, false
}

type inputTenderCreate struct {
	Name            string `json:"name" validate:"required"`
	Description     string `json:"description" validate:"required"`
	ServiceType     string `json:"serviceType" validate:"required,oneof=Construction Delivery Manufacture"`
	OrganizationId  string `json:"organizationId" validate:"required,uuid"`
	CreatorUsername string `json:"creatorUsername" validate:"required"`
}

type outputTenderCreate struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ServiceType string    `json:"serviceType"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (u *tenderRoutes) create(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input inputTenderCreate
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

		user, err, done := u.IsExistUser(w, r, err, ctx, log, input.CreatorUsername)
		if done {
			return
		}

		err, done = u.IsUserOrgResponsible(w, r, err, ctx, log, input.OrganizationId, user.Id)
		if done {
			return
		}

		// создание тендера
		var res entity.Tender
		if res, err = u.tenderService.Create(
			ctx, log, service.TenderCreateInput{
				Name:            input.Name,
				Description:     input.Description,
				ServiceType:     input.ServiceType,
				OrganizationId:  input.OrganizationId,
				CreatorUsername: input.CreatorUsername,
			},
		); err != nil {
			if err == service.ErrTenderAlreadyExists {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgTenderAlreadyExists)
				return
			}
			newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
			return
		}

		output := outputTenderCreate{
			Id:          res.Id,
			Name:        res.Name,
			Description: res.Description,
			Status:      res.Status,
			ServiceType: res.ServiceType,
			Version:     res.Version,
			CreatedAt:   res.CreatedAt,
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}

type inputGetByType struct {
	Limit       int      `validate:"omitempty,gte=0,lte=50"`
	Offset      int      `validate:"omitempty,gte=0"`
	ServiceType []string `validate:"omitempty,dive,required,oneof=Construction Delivery Manufacture"`
}

type outputGetByType struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ServiceType string    `json:"serviceType"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (u *tenderRoutes) getByType(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			limit  int
			offset int
			err    error
		)

		if len(r.URL.Query()["limit"]) == 0 {
			limit = 0
		} else {
			if limit, err = strconv.Atoi(r.URL.Query()["limit"][0]); err != nil {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgInvalidReq)
				return
			}
		}

		if len(r.URL.Query()["limit"]) == 0 {
			offset = 0
		} else {
			if offset, err = strconv.Atoi(r.URL.Query()["limit"][0]); err != nil {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgInvalidReq)
				return
			}
		}

		input := inputGetByType{
			Limit:       limit,
			Offset:      offset,
			ServiceType: r.URL.Query()["service_type"],
		}

		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}
		var tenders []entity.Tender
		if tenders, err = u.tenderService.GetByType(
			ctx, log, service.TenderGetByTypeInput{
				Limit:       input.Limit,
				Offset:      input.Offset,
				ServiceType: input.ServiceType,
			},
		); err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}
		var output []outputGetByType
		for _, t := range tenders {
			out := outputGetByType{
				Id:          t.Id,
				Name:        t.Name,
				Description: t.Description,
				Status:      t.Status,
				ServiceType: t.ServiceType,
				Version:     t.Version,
				CreatedAt:   t.CreatedAt,
			}
			output = append(output, out)
		}
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}

type inputTenderGetMy struct {
	Limit    int    `validate:"omitempty,number,gte=0,lte=50"`
	Offset   int    `validate:"omitempty,number,gte=0"`
	Username string `validate:"required"`
}

type outputTenderGetMy struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ServiceType string    `json:"serviceType"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (u *tenderRoutes) getMy(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			limit    int
			offset   int
			username string
			err      error
		)

		if len(r.URL.Query()["limit"]) == 0 {
			limit = 0
		} else {
			if limit, err = strconv.Atoi(r.URL.Query()["limit"][0]); err != nil {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgInvalidReq)
				return
			}
		}

		if len(r.URL.Query()["offset"]) == 0 {
			offset = 0
		} else {
			if offset, err = strconv.Atoi(r.URL.Query()["offset"][0]); err != nil {
				newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgInvalidReq)
				return
			}
		}

		if len(r.URL.Query()["username"]) == 0 {
			username = ""
		} else {
			username = r.URL.Query()["username"][0]
		}

		input := inputTenderGetMy{
			Limit:    limit,
			Offset:   offset,
			Username: username,
		}

		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		_, err, done := u.IsExistUser(w, r, err, ctx, log, input.Username)
		if done {
			return
		}

		var tenders []entity.Tender
		if tenders, err = u.tenderService.GetMy(
			ctx, log, service.TenderGetMyInput{
				Limit:    input.Limit,
				Offset:   input.Offset,
				Username: input.Username,
			},
		); err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}
		var output []outputTenderGetMy
		for _, t := range tenders {
			out := outputTenderGetMy{
				Id:          t.Id,
				Name:        t.Name,
				Description: t.Description,
				Status:      t.Status,
				ServiceType: t.ServiceType,
				Version:     t.Version,
				CreatedAt:   t.CreatedAt,
			}
			output = append(output, out)
		}
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}

type inputGetStatus struct {
	TenderId string `validate:"required,uuid"`
	Username string `validate:"required"`
}

func (u *tenderRoutes) getStatus(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err    error
			output entity.Tender
			user   entity.User
			done   bool
		)

		input := inputGetStatus{
			TenderId: chi.URLParam(r, "tenderId"),
			Username: r.URL.Query().Get("username"),
		}

		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		user, err, done = u.IsExistUser(w, r, err, ctx, log, input.Username)
		if done {
			return
		}

		output, err = u.tenderService.GetById(ctx, log, input.TenderId)
		if err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}

		switch output.Status {
		case statusCreated, statusClosed:
			err, done = u.IsUserOrgResponsible(w, r, err, ctx, log, output.OrganizationId, user.Id)
			if done {
				return
			}
		default:
			break
		}

		type status struct {
			Status string `json:"status"`
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, status{Status: output.Status})
	}
}

type inputSetStatus struct {
	TenderId string `validate:"required,uuid"`
	Status   string `validate:"required,oneof=Created Published Closed"`
	Username string `validate:"required"`
}

type outputSetStatus struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ServiceType string    `json:"serviceType"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (u *tenderRoutes) setStatus(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err  error
			out  entity.Tender
			user entity.User
			done bool
		)

		input := inputSetStatus{
			TenderId: chi.URLParam(r, "tenderId"),
			Status:   r.URL.Query().Get("status"),
			Username: r.URL.Query().Get("username"),
		}

		if err = validator.New().Struct(input); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		user, err, done = u.IsExistUser(w, r, err, ctx, log, input.Username)
		if done {
			return
		}

		var t entity.Tender
		if t, err = u.tenderService.GetById(ctx, log, input.TenderId); err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}

		err, done = u.IsUserOrgResponsible(w, r, err, ctx, log, t.OrganizationId, user.Id)
		if done {
			return
		}

		if out, err = u.tenderService.PutStatus(ctx, log, input.TenderId, input.Status); err != nil {
			newErrorResponse(w, r, log, err, http.StatusInternalServerError, MsgInternalServerErr)
			return
		}

		output := outputSetStatus{
			Id:          out.Id,
			Name:        out.Name,
			Description: out.Description,
			Status:      out.Status,
			ServiceType: out.ServiceType,
			Version:     out.Version,
			CreatedAt:   out.CreatedAt,
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}

type editParamsInput struct {
	TenderId string `validate:"required,uuid"`
	Username string `validate:"required"`
}

type editBodyInput struct {
	Name        string `json:"name" validate:"omitempty"`
	Description string `json:"description" validate:"omitempty"`
	ServiceType string `json:"serviceType" validate:"omitempty,oneof=Construction Delivery Manufacture"`
}

type outputEditTender struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	ServiceType string    `json:"serviceType"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (u *tenderRoutes) edit(ctx context.Context, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err  error
			out  entity.Tender
			done bool
		)

		inputParams := editParamsInput{
			TenderId: chi.URLParam(r, "tenderId"),
			Username: r.URL.Query().Get("username"),
		}

		if err = validator.New().Struct(inputParams); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		var inputBody editBodyInput
		if err = render.DecodeJSON(r.Body, &inputBody); err != nil {
			newErrorResponse(w, r, log, err, http.StatusBadRequest, MsgFailedParsing)
			return
		}

		if err = validator.New().Struct(inputBody); err != nil {
			newErrorValidateResponse(w, r, log, http.StatusBadRequest, MsgInvalidReq, err)
			return
		}

		_, err, done = u.IsExistUser(w, r, err, ctx, log, inputParams.Username)
		if done {
			return
		}

		if out, err = u.tenderService.EditTender(
			ctx,
			log, service.TenderEditInput{
				Name:        inputBody.Name,
				Description: inputBody.Description,
				ServiceType: inputBody.ServiceType,
			}, inputParams.TenderId,
		); err != nil {
			newErrorResponse(w, r, log, err, http.StatusNotFound, MsgTenderNotFound)
			return
		}

		output := outputEditTender{
			Id:          out.Id,
			Name:        out.Name,
			Description: out.Description,
			Status:      out.Status,
			ServiceType: out.ServiceType,
			Version:     out.Version,
			CreatedAt:   out.CreatedAt,
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, output)
	}
}
