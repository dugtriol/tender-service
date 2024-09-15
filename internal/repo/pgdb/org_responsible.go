package pgdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"tender-service/internal/entity"
	"tender-service/internal/repo/repoerrs"
	"tender-service/pkg/postgres"
)

const (
	orgResponsible = "organization_responsible"
)

type OrgResponsibleRepo struct {
	*postgres.Database
}

func NewOrgResponsibleRepo(db *postgres.Database) *OrgResponsibleRepo {
	return &OrgResponsibleRepo{db}
}

func (r *OrgResponsibleRepo) Create(
	ctx context.Context, input entity.OrgResponsible,
) (entity.OrgResponsible, error) {
	sql, args, _ := r.Builder.Insert(orgResponsible).Columns(
		"organization_id",
		"user_id",
	).Values(
		input.OrganizationId,
		input.UserId,
	).Suffix("RETURNING id, organization_id, user_id").ToSql()

	var output entity.OrgResponsible
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&output.Id,
		&output.OrganizationId,
		&output.UserId,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return entity.OrgResponsible{}, repoerrs.ErrAlreadyExists
			}
		}
		return entity.OrgResponsible{}, fmt.Errorf("OrgResponsibleRepo - Create - r.Cluster.QueryRow: %v", err)
	}
	return output, nil
}

func (r *OrgResponsibleRepo) GetById(ctx context.Context, id string) (entity.OrgResponsible, error) {
	sql, args, _ := r.Builder.
		Select("*").
		From(orgResponsible).
		Where("id = ?", id).
		ToSql()

	var output entity.OrgResponsible
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&output.Id,
		&output.OrganizationId,
		&output.UserId,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.OrgResponsible{}, repoerrs.ErrNotFound
		}
		return entity.OrgResponsible{}, fmt.Errorf("OrgResponsibleRepo - GetById - r.Cluster.QueryRow: %v", err)
	}

	return output, nil
}

func (r *OrgResponsibleRepo) GetByIds(ctx context.Context, input entity.OrgResponsible) (
	entity.OrgResponsible, error,
) {
	sql, args, _ := r.Builder.
		Select("*").
		From(orgResponsible).
		Where("organization_id = ? and user_id = ?", input.OrganizationId, input.UserId).ToSql()

	var output entity.OrgResponsible
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&output.Id,
		&output.OrganizationId,
		&output.UserId,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.OrgResponsible{}, repoerrs.ErrNotFound
		}
		return entity.OrgResponsible{}, fmt.Errorf("OrgResponsibleRepo - GetByIds - r.Cluster.QueryRow: %v", err)
	}

	return output, nil
}
