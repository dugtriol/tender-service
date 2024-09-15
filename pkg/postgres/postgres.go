package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxPoolSize  = 1
	defaultConnAttempts = 10
	defaultConnTimeout  = time.Second
)

type PgxPool interface {
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (
		int64, error,
	)
	Ping(ctx context.Context) error
}

type Database struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Cluster PgxPool
	Builder squirrel.StatementBuilderType
}

func New(ctx context.Context, connString string, opts ...Option) (*Database, error) {
	db := &Database{
		maxPoolSize:  defaultMaxPoolSize,
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(db)
	}

	db.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatal(err.Error())
	}

	config.MaxConns = int32(db.maxPoolSize)
	for db.connAttempts > 0 {
		db.Cluster, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", db.connAttempts)
		time.Sleep(db.connTimeout)
		db.connAttempts--
	}
	if err != nil {
		fmt.Errorf("database - New - pgxpool.NewWithConfig: %w", err)
		return nil, err
	}
	return db, nil
}

func (db *Database) Close() {
	if db.Cluster != nil {
		db.Cluster.Close()
	}
}
