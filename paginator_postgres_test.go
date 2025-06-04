package paginator_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"

	"github.com/Mikhalevich/paginator"
	"github.com/Mikhalevich/paginator/queryercache"
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

	if err := populateTestData(sqlDB, sqlTestRows); err != nil {
		s.FailNow("pupulate test data", err)
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
			id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			int_field BIGINT NOT NULL
		)`,
	); err != nil {
		return fmt.Errorf("create test table: %w", err)
	}

	if _, err := sql.Exec(`
		CREATE INDEX test_int_field_idx ON test(int_field)`,
	); err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	return nil
}

func populateTestData(sql *sql.DB, rowsCount int) error {
	trx, err := sql.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	//nolint:errcheck
	defer trx.Rollback()

	stmt, err := trx.Prepare(`INSERT INTO test(int_field) VALUES($1)`)
	if err != nil {
		return fmt.Errorf("tx prepare: %w", err)
	}

	defer stmt.Close()

	for range rowsCount {
		//nolint:gosec
		if _, err := stmt.Exec(rand.Int()); err != nil {
			return fmt.Errorf("exec: %w", err)
		}
	}

	if err := trx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

type TestTable struct {
	ID        int
	Int_Field int
}

type SqlQueryProvider struct {
	db *sql.DB
}

func (s *SqlQueryProvider) Query(ctx context.Context, offset int, limit int) ([]TestTable, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id,
			int_field
		FROM
			test
		ORDER BY
			int_field
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
		if err := rows.Scan(&data.ID, &data.Int_Field); err != nil {
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

	s.Require().Len(page.Data, 10)
	s.Require().Equal(1, page.BottomIndex)
	s.Require().Equal(10, page.TopIndex)
	s.Require().Equal(10, page.PageSize)
	s.Require().Equal(1, page.PageNumber)
	s.Require().Equal(11, page.PageTotalCount)
}

func BenchmarkPaginatorPostgres(b *testing.B) {
	sqlDB, cleanup, err := connectToDatabase()
	if err != nil {
		b.Fatal("could not connect to database", err)
	}

	//nolint:errcheck
	defer cleanup()

	if err := createDB(sqlDB); err != nil {
		b.Fatal("create db", err)
	}

	if err := populateTestData(sqlDB, sqlBenchmartRows); err != nil {
		b.Fatal("pupulate test data", err)
	}

	pag := paginator.New(&SqlQueryProvider{
		db: sqlDB,
	}, pageSize)

	page, err := pag.Page(b.Context(), 1)
	if err != nil {
		b.Fatal("first page", err)
	}

	pagesCount := page.PageTotalCount

	for b.Loop() {
		//nolint:gosec
		if _, err := pag.Page(b.Context(), rand.Int()%pagesCount+1); err != nil {
			b.Fatal("get page", err)
		}
	}
}

func BenchmarkCachedPaginatorPostgres(b *testing.B) {
	sqlDB, cleanup, err := connectToDatabase()
	if err != nil {
		b.Fatal("could not connect to database", err)
	}

	//nolint:errcheck
	defer cleanup()

	if err := createDB(sqlDB); err != nil {
		b.Fatal("create db", err)
	}

	if err := populateTestData(sqlDB, sqlBenchmartRows); err != nil {
		b.Fatal("pupulate test data", err)
	}

	pag := paginator.New(
		queryercache.New(
			&SqlQueryProvider{
				db: sqlDB,
			}, queryercache.WithCountTTL(time.Minute*5),
		),
		pageSize,
	)

	page, err := pag.Page(b.Context(), 1)
	if err != nil {
		b.Fatal("first page", err)
	}

	pagesCount := page.PageTotalCount

	for b.Loop() {
		//nolint:gosec
		if _, err := pag.Page(b.Context(), rand.Int()%pagesCount+1); err != nil {
			b.Fatal("get page", err)
		}
	}
}
