package paginator_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"

	"github.com/Mikhalevich/paginator"
)

type PaginatorPostgres struct {
	*suite.Suite

	dbCleanup func() error
	pag       *paginator.Paginator[TestTable]
}

func TestPaginatorPostgresSuit(t *testing.T) {
	t.Parallel()

	suite.Run(t, &PaginatorPostgres{
		Suite: new(suite.Suite),
	})
}

func (s *PaginatorPostgres) SetupSuite() {
	sqlDB, cleanup, err := connectToDatabase()
	if err != nil {
		s.FailNow("could not connect to database", err)
	}

	s.dbCleanup = cleanup

	if err := createDB(sqlDB); err != nil {
		s.FailNow("create db", err)
	}

	s.pag = paginator.New(&SqlQueryProvider{
		db: sqlDB,
	}, pageSize)
}

func (s *PaginatorPostgres) TearDownSuite() {
	if err := s.dbCleanup(); err != nil {
		s.FailNow("db cleanup", err)
	}
}

func (s *PaginatorPostgres) TearDownTest() {
}

func (s *PaginatorPostgres) TearDownSubTest() {
}

func connectToDatabase() (*sql.DB, func() error, error) {
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

func createDB(sql *sql.DB) error {
	if _, err := sql.Exec(`
		CREATE TABLE test(
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			int_field INTEGER NOT NULL,
			text_field TEXT NOT NULL
		)`,
	); err != nil {
		return fmt.Errorf("create test table: %w", err)
	}

	if err := populateTestData(sql); err != nil {
		return fmt.Errorf("pupulate test data: %w", err)
	}

	return nil
}

func populateTestData(sql *sql.DB) error {
	trx, err := sql.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	//nolint:errcheck
	defer trx.Rollback()

	stmt, err := trx.Prepare(`INSERT INTO test(int_field, text_field) VALUES($1, $2)`)
	if err != nil {
		return fmt.Errorf("tx prepare: %w", err)
	}

	defer stmt.Close()

	for i := range sqlRows {
		if _, err := stmt.Exec(i+1, fmt.Sprintf("text_%d", i+1)); err != nil {
			return fmt.Errorf("exec: %w", err)
		}
	}

	if err := trx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

type TestTable struct {
	ID         int
	Int_Field  int
	Text_Field string
}

type SqlQueryProvider struct {
	db *sql.DB
}

func (s *SqlQueryProvider) Query(ctx context.Context, offset int, limit int) ([]TestTable, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			int_field,
			text_field
		FROM
			test
		OFFSET $1
		LIMIT $2
	`, offset, limit)

	if err != nil {
		return nil, fmt.Errorf("query context: %w", err)
	}

	defer rows.Close()

	results := make([]TestTable, 0, limit)

	for rows.Next() {
		var data TestTable
		if err := rows.Scan(&data.ID, &data.Int_Field, &data.Text_Field); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		results = append(results, data)
	}

	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("rows close: %w", err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return results, nil
}

func (s *SqlQueryProvider) Count(ctx context.Context) (int, error) {
	var count int

	if err := s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*)
		FROM
			test
	`).Scan(&count); err != nil {
		return 0, fmt.Errorf("query row context: %w", err)
	}

	return count, nil
}

func (s *PaginatorPostgres) TestFirstPage() {
	var (
		ctx = context.Background()
	)

	page, err := s.pag.Page(ctx, 1)

	s.Require().NoError(err)

	s.Require().ElementsMatch([]TestTable{
		{
			ID:         1,
			Int_Field:  1,
			Text_Field: "text_1",
		},
		{
			ID:         2,
			Int_Field:  2,
			Text_Field: "text_2",
		},
		{
			ID:         3,
			Int_Field:  3,
			Text_Field: "text_3",
		},
		{
			ID:         4,
			Int_Field:  4,
			Text_Field: "text_4",
		},
		{
			ID:         5,
			Int_Field:  5,
			Text_Field: "text_5",
		},
		{
			ID:         6,
			Int_Field:  6,
			Text_Field: "text_6",
		},
		{
			ID:         7,
			Int_Field:  7,
			Text_Field: "text_7",
		},
		{
			ID:         8,
			Int_Field:  8,
			Text_Field: "text_8",
		},
		{
			ID:         9,
			Int_Field:  9,
			Text_Field: "text_9",
		},
		{
			ID:         10,
			Int_Field:  10,
			Text_Field: "text_10",
		},
	}, page.Data)
	s.Require().Equal(1, page.BottomIndex)
	s.Require().Equal(10, page.TopIndex)
	s.Require().Equal(10, page.PageSize)
	s.Require().Equal(1, page.PageNumber)
	s.Require().Equal(11, page.PageTotalCount)
}
