package repo

import (
	"context"

	"tender-service/internal/entity"
	"tender-service/internal/repo/pgdb"
	"tender-service/pkg/postgres"
)

type User interface {
	Create(ctx context.Context, input entity.User) (string, error)
	GetById(ctx context.Context, id string) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.User, error)
}

type Organization interface {
	Create(ctx context.Context, input entity.Organization) (entity.Organization, error)
	GetById(ctx context.Context, id string) (entity.Organization, error)
}

type OrgResponsible interface {
	Create(
		ctx context.Context, input entity.OrgResponsible,
	) (entity.OrgResponsible, error)
	GetById(ctx context.Context, id string) (entity.OrgResponsible, error)
	GetByIds(ctx context.Context, input entity.OrgResponsible) (
		entity.OrgResponsible, error,
	)
}

type Tender interface {
	Create(ctx context.Context, input entity.Tender) (entity.Tender, error)
	GetById(ctx context.Context, id string) (entity.Tender, error)
	GetByTypePagination(ctx context.Context, limit, offset int, serviceType []string) (
		[]entity.Tender, error,
	)
	GetMyPagination(ctx context.Context, limit, offset int, username string) (
		[]entity.Tender, error,
	)
	PutStatus(ctx context.Context, tenderId, status string) error
	EditTender(ctx context.Context, input entity.Tender, tenderId string) error
	IncrementVersion(ctx context.Context, tenderId string) error
}

type Bid interface {
	Create(ctx context.Context, input entity.Bid) (entity.Bid, error)
	GetById(ctx context.Context, bidId string) (entity.Bid, error)
	GetMyPagination(ctx context.Context, limit, offset int, authorId string) ([]entity.Bid, error)
	GetByTenderID(ctx context.Context, limit, offset int, authorId, tenderId string) ([]entity.Bid, error)
	PutStatus(ctx context.Context, bidId, status string) error
	EditBid(ctx context.Context, input entity.Bid, bidId string) error
	IncrementVersion(ctx context.Context, bidId string) error
}

type Repositories struct {
	User
	Organization
	OrgResponsible
	Tender
	Bid
}

func NewRepositories(db *postgres.Database) *Repositories {
	return &Repositories{
		User:           pgdb.NewUserRepo(db),
		Organization:   pgdb.NewOrganizationRepo(db),
		OrgResponsible: pgdb.NewOrgResponsibleRepo(db),
		Tender:         pgdb.NewTenderRepo(db),
		Bid:            pgdb.NewBidRepo(db),
	}
}
