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
	organization = "organization"
)

type OrganizationRepo struct {
	*postgres.Database
}

func NewOrganizationRepo(db *postgres.Database) *OrganizationRepo {
	return &OrganizationRepo{db}
}

func (r *OrganizationRepo) Create(ctx context.Context, input entity.Organization) (entity.Organization, error) {
	sql, args, _ := r.Builder.Insert(organization).Columns(
		"name",
		"description",
		"type",
	).Values(
		input.Name,
		input.Description,
		input.OrganizationType,
	).Suffix("RETURNING id, name, description, type, created_at, updated_at").ToSql()

	var output entity.Organization
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&output.Id,
		&output.Name,
		&output.Description,
		&output.OrganizationType,
		&output.CreatedAt,
		&output.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return entity.Organization{}, repoerrs.ErrAlreadyExists
			}
		}
		return entity.Organization{}, fmt.Errorf("OrganizationRepo - Create - r.Cluster.QueryRow: %v", err)
	}
	return output, nil
}

func (r *OrganizationRepo) GetById(ctx context.Context, id string) (entity.Organization, error) {
	sql, args, _ := r.Builder.
		Select("*").
		From(organization).
		Where("id = ?", id).
		ToSql()

	var output entity.Organization
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&output.Id,
		&output.Name,
		&output.Description,
		&output.OrganizationType,
		&output.CreatedAt,
		&output.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Organization{}, repoerrs.ErrNotFound
		}
		return entity.Organization{}, fmt.Errorf("OrganizationRepo - GetById - r.Cluster.QueryRow: %v", err)
	}

	return output, nil
}
