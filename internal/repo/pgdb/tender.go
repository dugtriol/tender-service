package pgdb

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"tender-service/internal/entity"
	"tender-service/internal/repo/repoerrs"
	"tender-service/pkg/postgres"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	tender = "tender"

	maxPaginationLimit     = 50
	defaultPaginationLimit = 5
)

type TenderRepo struct {
	*postgres.Database
}

func NewTenderRepo(db *postgres.Database) *TenderRepo {
	return &TenderRepo{db}
}

func (r *TenderRepo) Create(ctx context.Context, input entity.Tender) (entity.Tender, error) {
	sql, args, _ := r.Builder.Insert(tender).Columns(
		"name",
		"description",
		"type",
		"organization_id",
		"creator_username",
	).Values(
		input.Name,
		input.Description,
		input.ServiceType,
		input.OrganizationId,
		input.CreatorUsername,
	).Suffix(
		"RETURNING id, name, description, type, status, " +
			"organization_id, version, created_at, creator_username",
	).ToSql()

	var output entity.Tender
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&output.Id,
		&output.Name,
		&output.Description,
		&output.ServiceType,
		&output.Status,
		&output.OrganizationId,
		&output.Version,
		&output.CreatedAt,
		&output.CreatorUsername,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return entity.Tender{}, repoerrs.ErrAlreadyExists
			}
		}
		return entity.Tender{}, fmt.Errorf("TenderRepo - Create - r.Cluster.QueryRow: %v", err)
	}
	return output, nil
}

func (r *TenderRepo) GetById(ctx context.Context, id string) (entity.Tender, error) {
	sql, args, _ := r.Builder.
		Select("*").
		From(tender).
		Where("id = ?", id).
		ToSql()

	var output entity.Tender
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&output.Id,
		&output.Name,
		&output.Description,
		&output.ServiceType,
		&output.Status,
		&output.OrganizationId,
		&output.Version,
		&output.CreatedAt,
		&output.CreatorUsername,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Tender{}, repoerrs.ErrNotFound
		}
		return entity.Tender{}, fmt.Errorf("TenderRepo - GetById - r.Cluster.QueryRow: %v", err)
	}
	return output, nil
}

func (r *TenderRepo) GetByTypePagination(ctx context.Context, limit, offset int, serviceType []string) (
	[]entity.Tender, error,
) {
	type sqlData struct {
		sql  string
		args []interface{}
	}
	var (
		sqlArray []sqlData
	)
	if limit > maxPaginationLimit {
		limit = maxPaginationLimit
	}
	if limit == 0 {
		limit = defaultPaginationLimit
	}

	orderBySql := "name"
	if len(serviceType) == 0 {
		sql, args, err := r.Builder.
			Select("*").
			From(tender).
			OrderBy(orderBySql).
			Limit(uint64(limit)).
			Offset(uint64(offset)).
			ToSql()
		if err != nil {
			return nil, fmt.Errorf("TenderRepo - GetByTypePagination - r.Builder: %v", err)
		}
		sqlArray = append(
			sqlArray, sqlData{
				sql:  sql,
				args: args,
			},
		)
	} else {
		for _, stype := range serviceType {
			sql, args, err := r.Builder.
				Select("*").
				From(tender).
				Where("type = ?", stype).
				OrderBy(orderBySql).
				Limit(uint64(limit)).
				Offset(uint64(offset)).
				ToSql()
			if err != nil {
				return nil, fmt.Errorf("TenderRepo - GetByTypePagination - r.Builder: %v", err)
			}
			sqlArray = append(
				sqlArray, sqlData{
					sql:  sql,
					args: args,
				},
			)
		}
	}

	var output []entity.Tender
	for _, data := range sqlArray {
		rows, err := r.Cluster.Query(ctx, data.sql, data.args...)
		if err != nil {
			return nil, fmt.Errorf("TenderRepo - GetByTypePagination - r.Cluster.Query: %v", err)
		}
		for rows.Next() {
			var t entity.Tender
			if err = rows.Scan(
				&t.Id,
				&t.Name,
				&t.Description,
				&t.ServiceType,
				&t.Status,
				&t.OrganizationId,
				&t.Version,
				&t.CreatedAt,
				&t.CreatorUsername,
			); err != nil {
				return nil, fmt.Errorf("TenderRepo - GetByTypePagination - rows.Scan: %v", err)
			}
			output = append(output, t)
		}
		rows.Close()
	}

	return output, nil
}

func (r *TenderRepo) GetMyPagination(ctx context.Context, limit, offset int, username string) (
	[]entity.Tender, error,
) {
	if limit > maxPaginationLimit {
		limit = maxPaginationLimit
	}
	if limit == 0 {
		limit = defaultPaginationLimit
	}

	orderBySql := "name"
	sql, args, err := r.Builder.
		Select("*").
		From(tender).
		Where("creator_username = ?", username).
		OrderBy(orderBySql).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("TenderRepo - GetByTenderID - r.Builder: %v", err)
	}

	var output []entity.Tender
	rows, err := r.Cluster.Query(ctx, sql, args...)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("TenderRepo - GetByTenderID - r.Cluster.Query: %v", err)
	}
	for rows.Next() {
		var t entity.Tender
		if err = rows.Scan(
			&t.Id,
			&t.Name,
			&t.Description,
			&t.ServiceType,
			&t.Status,
			&t.OrganizationId,
			&t.Version,
			&t.CreatedAt,
			&t.CreatorUsername,
		); err != nil {
			return nil, fmt.Errorf("TenderRepo - GetByTenderID - rows.Scan: %v", err)
		}
		output = append(output, t)
	}

	return output, nil
}

type SqlData struct {
	Sql  string
	Args []interface{}
}

func (r *TenderRepo) GetSqlData(tenderId, column, field string) (SqlData, error) {
	var err error

	sql, args, err := r.
		Builder.
		Update(tender).
		Set(column, field).
		Where("id = ?", tenderId).
		ToSql()
	if err != nil {
		return SqlData{}, fmt.Errorf("TenderRepo.GetSqlData - r.Builder: %v", err)
	}

	return SqlData{
		Sql:  sql,
		Args: args,
	}, nil
}

func (r *TenderRepo) PutStatus(ctx context.Context, tenderId, status string) error {
	var (
		err error
		tx  pgx.Tx
	)
	tx, err = r.Cluster.Begin(ctx)
	if err != nil {
		return fmt.Errorf("TenderRepo.PutStatus - r.Cluster.Begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	statusSql, err := r.GetSqlData(tenderId, "status", status)

	sql, args, err := r.
		Builder.
		Update(tender).
		Set("status", status).
		Where("id = ?", tenderId).
		ToSql()

	_, err = tx.Exec(ctx, statusSql.Sql, statusSql.Args...)
	if err != nil {
		return fmt.Errorf("TenderRepo.PutStatus - tx.Exec.statusSql: %v", err)
	}
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("TenderRepo.PutStatus - tx.Exec.version: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("TenderRepo.PutStatus - tx.Commit: %v", err)
	}
	return nil
}

func (r *TenderRepo) EditTender(ctx context.Context, input entity.Tender, tenderId string) error {
	var err error

	type inputShort struct {
		Name        string
		Description string
		ServiceType string
	}
	val := reflect.ValueOf(
		inputShort{
			Name:        input.Name,
			Description: input.Description,
			ServiceType: input.ServiceType,
		},
	)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {

		var (
			tx   pgx.Tx
			sql  string
			args []interface{}
		)

		field := val.Field(i)
		fieldType := typ.Field(i)

		if len(field.String()) == 0 || len(fieldType.Name) == 0 {
			continue
		}

		tx, err = r.Cluster.Begin(ctx)
		if err != nil {
			return fmt.Errorf("TenderRepo.EditBid - r.Cluster.Begin: %v", err)
		}

		var column string
		switch fieldType.Name {
		case "Name":
			column = "name"
		case "Description":
			column = "description"
		case "ServiceType":
			column = "type"
		default:
			break
		}

		fieldVal := fmt.Sprintf("%s", field.String())

		if len(fieldVal) == 0 {
			continue
		}
		sql, args, err = r.
			Builder.
			Update(tender).
			Set(column, fieldVal).
			Where("id = ?", tenderId).
			ToSql()

		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("TenderRepo.EditBid - r.Builder: %v", err)
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("TenderRepo.EditBid - tx.Exec: %v - %s - %s - %s", err, sql, column, fieldType.Name)
		}

		err = tx.Commit(ctx)
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("TenderRepo.EditBid - tx.Commit: %v", err)
		}
	}
	return nil
}

func (r *TenderRepo) IncrementVersion(ctx context.Context, tenderId string) error {
	var (
		err error
		t   entity.Tender
		tx  pgx.Tx
	)

	t, err = r.GetById(ctx, tenderId)
	if err != nil {
		return fmt.Errorf("TenderRepo.IncrementVersion - r.GetById: %v", err)
	}

	tx, err = r.Cluster.Begin(ctx)
	if err != nil {
		return fmt.Errorf("TenderRepo.IncrementVersion - r.Cluster.Begin: %v", err)
	}

	sql, args, err := r.
		Builder.
		Update(tender).
		Set("version", t.Version+1).
		Where("id = ?", t.Id).
		ToSql()
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("TenderRepo.IncrementVersion - tx.Exec: %v", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("TenderRepo.IncrementVersion - tx.Exec: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("TenderRepo.IncrementVersion - tx.Commit: %v", err)
	}
	return nil
}
