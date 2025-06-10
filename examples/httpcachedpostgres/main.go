package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Mikhalevich/paginator"
	"github.com/Mikhalevich/paginator/queryercache"
	"github.com/Mikhalevich/paginator/queryercache/metrics"
)

const (
	dataLen  = 101
	pageSize = 10
)

func main() {
	dbConn, cleanup, err := runDockerPostgres()
	if err != nil {
		log.Printf("run docker postgres error: %v\n", err)

		return
	}

	defer func() {
		if err := cleanup(); err != nil {
			log.Printf("cleanup error: %v\n", err)
		}
	}()

	postgres := NewPostgresQueryProvider(dbConn)

	if err := postgres.CreateSchema(); err != nil {
		log.Printf("create schema error: %v\n", err)

		return
	}

	if err := postgres.PopulateTestData(dataLen); err != nil {
		log.Printf("populate test data error: %v\n", err)

		return
	}

	handler := NewHandler(paginator.New(
		queryercache.New(
			postgres,
			queryercache.WithCountTTL(time.Minute),
			queryercache.WithQueryTTL(time.Minute),
			queryercache.WithMetrics(metrics.NewPrometheus()),
		),
		pageSize,
	))

	http.HandleFunc("GET /page/{id}/", handler.TestTablePage)
	http.Handle("GET /metrics/", promhttp.Handler())

	//nolint:gosec
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("listen and server error: %v", err)

		return
	}
}

func runDockerPostgres() (*sql.DB, func() error, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("construct pool: %w", err)
	}

	if err := pool.Client.Ping(); err != nil {
		return nil, nil, fmt.Errorf("connect to docker: %w", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "17.5-alpine3.22",
		Env: []string{
			"POSTGRES_DB=test",
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=test",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	if err != nil {
		return nil, nil, fmt.Errorf("run docker: %w", err)
	}

	var sqlDB *sql.DB

	if err := pool.Retry(func() error {
		sqlDB, err = sql.Open("pgx",
			fmt.Sprintf("host=localhost port=%s user=test password=test dbname=test sslmode=disable",
				resource.GetPort("5432/tcp")))
		if err != nil {
			return fmt.Errorf("sql open: %w", err)
		}

		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("ping: %w", err)
		}

		return nil
	}); err != nil {
		return nil, nil, fmt.Errorf("connect to database: %w", err)
	}

	return sqlDB, func() error {
		if err := pool.Purge(resource); err != nil {
			return fmt.Errorf("purge resource: %w", err)
		}

		return nil
	}, nil
}
