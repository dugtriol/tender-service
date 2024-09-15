package service

import (
	"context"
	"log/slog"

	"tender-service/internal/entity"
	"tender-service/internal/repo"
)

type UserCreateInput struct {
	Username  string
	FirstName string
	LastName  string
}

type UserGetByIdInput struct {
	Id string
}

type UserGetByUsernameInput struct {
	Username string
}

type User interface {
	Create(ctx context.Context, log *slog.Logger, input UserCreateInput) (string, error)
	GetById(ctx context.Context, log *slog.Logger, input UserGetByIdInput) (entity.User, error)
	GetByUsername(ctx context.Context, log *slog.Logger, input UserGetByUsernameInput) (
		entity.User, error,
	)
}

type OrganizationCreateInput struct {
	Name             string
	Description      string
	OrganizationType string
}

type OrganizationGetInput struct {
	Id string
}

type Organization interface {
	Create(ctx context.Context, log *slog.Logger, input OrganizationCreateInput) (entity.Organization, error)
	Get(ctx context.Context, log *slog.Logger, input OrganizationGetInput) (entity.Organization, error)
}

type OrgResponsibleCreateInput struct {
	OrganizationId string
	UserId         string
}

type OrgResponsibleGetInput struct {
	Id string
}

type OrgResponsibleGetByIdsInput struct {
	OrganizationId string
	UserId         string
}

type OrgResponsible interface {
	Create(
		ctx context.Context, log *slog.Logger, input OrgResponsibleCreateInput,
	) (entity.OrgResponsible, error)
	Get(
		ctx context.Context, log *slog.Logger, input OrgResponsibleGetInput,
	) (entity.OrgResponsible, error)
	GetByIds(
		ctx context.Context, log *slog.Logger, input OrgResponsibleGetByIdsInput,
	) (entity.OrgResponsible, error)
}

type TenderCreateInput struct {
	Name            string
	Description     string
	ServiceType     string
	OrganizationId  string
	CreatorUsername string
}

type TenderGetByTypeInput struct {
	Limit       int
	Offset      int
	ServiceType []string
}

type TenderGetMyInput struct {
	Limit    int
	Offset   int
	Username string
}

type TenderEditInput struct {
	Name        string
	Description string
	ServiceType string
}

type Tender interface {
	Create(
		ctx context.Context, log *slog.Logger, input TenderCreateInput,
	) (entity.Tender, error)
	GetByType(
		ctx context.Context, log *slog.Logger, input TenderGetByTypeInput,
	) ([]entity.Tender, error)
	GetMy(
		ctx context.Context, log *slog.Logger, input TenderGetMyInput,
	) ([]entity.Tender, error)
	GetById(
		ctx context.Context, log *slog.Logger, id string,
	) (entity.Tender, error)
	PutStatus(ctx context.Context, log *slog.Logger, tenderId, status string) (entity.Tender, error)
	EditTender(
		ctx context.Context, log *slog.Logger, input TenderEditInput, tenderId string,
	) (entity.Tender, error)
}

type BidCreateInput struct {
	Name        string
	Description string
	TenderId    string
	AuthorType  string
	AuthorId    string
}

type BidGetByTenderIdInput struct {
	Limit    int
	Offset   int
	TenderId string
	UserId   string
}

type BidGetMyInput struct {
	Limit  int
	Offset int
	UserId string
}

type BidEditInput struct {
	Name        string
	Description string
}

type Bid interface {
	Create(
		ctx context.Context, log *slog.Logger, input BidCreateInput,
	) (entity.Bid, error)
	GetByTenderId(
		ctx context.Context, log *slog.Logger, input BidGetByTenderIdInput,
	) ([]entity.Bid, error)
	GetById(
		ctx context.Context, log *slog.Logger, bidId string,
	) (entity.Bid, error)
	GetMy(
		ctx context.Context, log *slog.Logger, input BidGetMyInput,
	) ([]entity.Bid, error)
	PutStatus(ctx context.Context, log *slog.Logger, bidId, status string) (entity.Bid, error)
	EditBid(ctx context.Context, log *slog.Logger, input BidEditInput, bidId string) (
		entity.Bid, error,
	)
}

type Services struct {
	User           User
	Organization   Organization
	OrgResponsible OrgResponsible
	Tender         Tender
	Bid            Bid
}

type ServicesDependencies struct {
	Repos *repo.Repositories
}

func NewServices(dep ServicesDependencies) *Services {
	return &Services{
		User:           NewUserService(dep.Repos.User),
		Organization:   NewOrganizationService(dep.Repos.Organization),
		OrgResponsible: NewOrgResponsibleService(dep.Repos.OrgResponsible),
		Tender:         NewTenderService(dep.Repos.Tender),
		Bid:            NewBidService(dep.Repos.Bid),
	}
}
