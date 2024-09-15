package app

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	defaultAttempts = 20
	defaultTimeout  = time.Second
)

func init() {
	log.Printf("Migrate: start")
	conn, ok := os.LookupEnv("POSTGRES_CONN")
	if !ok {
		fmt.Printf("%s not set\n", "POSTGRES_CONN")
	}
	conn += "?sslmode=disable"

	attempts := defaultAttempts
	var m *migrate.Migrate
	var err error

	for attempts > 0 {
		m, err = migrate.New(
			"file://migrations",
			conn,
		)
		if err == nil {
			break
		}

		log.Printf("Migrate: pgdb is trying to connect, attempts left: %d", attempts)
		time.Sleep(defaultTimeout)
		attempts--
	}
	err = m.Up()
	defer func() { _, _ = m.Close() }()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Migrate: up error: %s", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migrate: no change")
		return
	}

	log.Printf("Migrate: up success")
}
