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
	employee = "employee"
)

type UserRepo struct {
	*postgres.Database
}

func NewUserRepo(db *postgres.Database) *UserRepo {
	return &UserRepo{db}
}

func (r *UserRepo) Create(ctx context.Context, user entity.User) (string, error) {
	sql, args, _ := r.Builder.Insert(employee).Columns("username", "first_name", "last_name").Values(
		user.Username,
		user.FirstName,
		user.LastName,
	).Suffix("RETURNING id").ToSql()

	var id string
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return "", repoerrs.ErrAlreadyExists
			}
		}
		return "", fmt.Errorf("UserRepo - Create - r.Cluster.QueryRow: %v", err)
	}
	return id, nil
}

func (r *UserRepo) GetById(ctx context.Context, id string) (entity.User, error) {
	sql, args, _ := r.Builder.
		Select("*").
		From(employee).
		Where("id = ?", id).
		ToSql()

	var user entity.User
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&user.Id,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, repoerrs.ErrNotFound
		}
		return entity.User{}, fmt.Errorf("UserRepo - GetById - r.Cluster.QueryRow: %v", err)
	}

	return user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	sql, args, _ := r.Builder.
		Select("*").
		From(employee).
		Where("username = ?", username).
		ToSql()

	var user entity.User
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&user.Id,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, repoerrs.ErrNotFound
		}
		return entity.User{}, fmt.Errorf("UserRepo - GetByUsername - r.Cluster.QueryRow: %v", err)
	}

	return user, nil
}
