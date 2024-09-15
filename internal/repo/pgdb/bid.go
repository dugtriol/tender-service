package pgdb

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"tender-service/internal/entity"
	"tender-service/internal/repo/repoerrs"
	"tender-service/pkg/postgres"
)

const (
	bidTable = "bid"
)

type BidRepo struct {
	*postgres.Database
}

func NewBidRepo(db *postgres.Database) *BidRepo {
	return &BidRepo{db}
}

func (r *BidRepo) Create(ctx context.Context, input entity.Bid) (entity.Bid, error) {
	sql, args, _ := r.Builder.Insert(bidTable).Columns(
		"name",
		"description",
		"tender_id",
		"author_type",
		"author_id",
	).Values(
		input.Name,
		input.Description,
		input.TenderId,
		input.AuthorType,
		input.AuthorId,
	).Suffix(
		"RETURNING id, name, description, status, " +
			"tender_id, author_type, author_id, version, created_at",
	).ToSql()

	var output entity.Bid
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&output.Id,
		&output.Name,
		&output.Description,
		&output.Status,
		&output.TenderId,
		&output.AuthorType,
		&output.AuthorId,
		&output.Version,
		&output.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return entity.Bid{}, repoerrs.ErrAlreadyExists
			}
		}
		return entity.Bid{}, fmt.Errorf("BidRepo - Create - r.Cluster.QueryRow: %v", err)
	}
	return output, nil
}

func (r *BidRepo) GetById(ctx context.Context, bidId string) (entity.Bid, error) {
	sql, args, _ := r.Builder.
		Select("*").
		From(bidTable).
		Where("id = ?", bidId).
		ToSql()

	var output entity.Bid
	err := r.Cluster.QueryRow(ctx, sql, args...).Scan(
		&output.Id,
		&output.Name,
		&output.Description,
		&output.Status,
		&output.TenderId,
		&output.AuthorType,
		&output.AuthorId,
		&output.Version,
		&output.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Bid{}, repoerrs.ErrNotFound
		}
		return entity.Bid{}, fmt.Errorf("BidRepo - GetById - r.Cluster.QueryRow: %v", err)
	}
	return output, nil
}

// GetMyPagination change username на user id
func (r *BidRepo) GetMyPagination(ctx context.Context, limit, offset int, authorId string) ([]entity.Bid, error) {
	if limit > maxPaginationLimit {
		limit = maxPaginationLimit
	}
	if limit == 0 {
		limit = defaultPaginationLimit
	}

	orderBySql := "name"
	sql, args, err := r.Builder.
		Select("*").
		From(bidTable).
		Where("author_id = ?", authorId).
		OrderBy(orderBySql).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("BidRepo - GetByTenderID - r.Builder: %v", err)
	}

	var output []entity.Bid
	rows, err := r.Cluster.Query(ctx, sql, args...)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("BidRepo - GetByTenderID - r.Cluster.Query: %v", err)
	}
	for rows.Next() {
		var t entity.Bid
		if err = rows.Scan(
			&t.Id,
			&t.Name,
			&t.Description,
			&t.Status,
			&t.TenderId,
			&t.AuthorType,
			&t.AuthorId,
			&t.Version,
			&t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("BidRepo - GetByTenderID - rows.Scan: %v", err)
		}
		output = append(output, t)
	}

	return output, nil
}

// GetByTenderID change username на user id
// /bids/{tenderId}/list
func (r *BidRepo) GetByTenderID(ctx context.Context, limit, offset int, authorId, tenderId string) (
	[]entity.Bid, error,
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
		From(bidTable).
		Where("author_id = ?", authorId).
		Where("tender_id = ?", tenderId).
		OrderBy(orderBySql).
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("BidRepo - GetByTenderID - r.Builder: %v", err)
	}

	var output []entity.Bid
	rows, err := r.Cluster.Query(ctx, sql, args...)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("BidRepo - GetByTenderID - r.Cluster.Query: %v", err)
	}
	for rows.Next() {
		var t entity.Bid
		if err = rows.Scan(
			&t.Id,
			&t.Name,
			&t.Description,
			&t.Status,
			&t.TenderId,
			&t.AuthorType,
			&t.AuthorId,
			&t.Version,
			&t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("BidRepo - GetByTenderID - rows.Scan: %v", err)
		}
		output = append(output, t)
	}

	return output, nil
}

func (r *BidRepo) GetSqlData(bidId, column, field string) (SqlData, error) {
	var err error

	sql, args, err := r.
		Builder.
		Update(bidTable).
		Set(column, field).
		Where("id = ?", bidId).
		ToSql()
	if err != nil {
		return SqlData{}, fmt.Errorf("TenderRepo.GetSqlData - r.Builder: %v", err)
	}

	return SqlData{
		Sql:  sql,
		Args: args,
	}, nil
}

func (r *BidRepo) PutStatus(ctx context.Context, bidId, status string) error {
	var (
		err error
		tx  pgx.Tx
	)
	tx, err = r.Cluster.Begin(ctx)
	if err != nil {
		return fmt.Errorf("BidRepo.PutStatus - r.Cluster.Begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	statusSql, err := r.GetSqlData(bidId, "status", status)

	sql, args, err := r.
		Builder.
		Update(bidTable).
		Set("status", status).
		Where("id = ?", bidId).
		ToSql()

	_, err = tx.Exec(ctx, statusSql.Sql, statusSql.Args...)
	if err != nil {
		return fmt.Errorf("BidRepo.PutStatus - tx.Exec.statusSql: %v", err)
	}
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("BidRepo.PutStatus - tx.Exec.version: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("BidRepo.PutStatus - tx.Commit: %v", err)
	}
	return nil
}

func (r *BidRepo) EditBid(ctx context.Context, input entity.Bid, bidId string) error {
	var err error

	type inputShort struct {
		Name        string
		Description string
	}
	val := reflect.ValueOf(
		inputShort{
			Name:        input.Name,
			Description: input.Description,
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

		fmt.Println("Tx Begin")
		tx, err = r.Cluster.Begin(ctx)
		if err != nil {
			return fmt.Errorf("TenderRepo.EditBid - r.Cluster.Begin: %v", err)
		}

		fmt.Println("Switch")
		var column string
		switch fieldType.Name {
		case "Name":
			column = "name"
		case "Description":
			column = "description"
		default:
			break
		}

		fieldVal := fmt.Sprintf("%s", field.String())

		if len(fieldVal) == 0 {
			continue
		}

		sql, args, err = r.
			Builder.
			Update(bidTable).
			Set(column, fieldVal).
			Where("id = ?", bidId).
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

func (r *BidRepo) IncrementVersion(ctx context.Context, bidId string) error {
	var (
		err error
		t   entity.Bid
		tx  pgx.Tx
	)

	t, err = r.GetById(ctx, bidId)
	if err != nil {
		return fmt.Errorf("BidRepo.IncrementVersion - r.GetById: %v", err)
	}

	tx, err = r.Cluster.Begin(ctx)
	if err != nil {
		return fmt.Errorf("BidRepo.IncrementVersion - r.Cluster.Begin: %v", err)
	}

	sql, args, err := r.
		Builder.
		Update(bidTable).
		Set("version", t.Version+1).
		Where("id = ?", t.Id).
		ToSql()
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("BidRepo.IncrementVersion - tx.Exec: %v", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("BidRepo.IncrementVersion - tx.Exec: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("BidRepo.IncrementVersion - tx.Commit: %v", err)
	}
	return nil
}
