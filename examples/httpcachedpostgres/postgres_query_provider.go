package main

import (
	"context"
	"database/sql"
	"fmt"
)

type TestTable struct {
	ID           int
	OrderedField int
}

type PostgresQueryProvider struct {
	db *sql.DB
}

func NewPostgresQueryProvider(db *sql.DB) *PostgresQueryProvider {
	return &PostgresQueryProvider{
		db: db,
	}
}

func (p *PostgresQueryProvider) CreateSchema() error {
	if _, err := p.db.Exec(`
		CREATE TABLE test(
			id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			ordered_field BIGINT NOT NULL
		)`,
	); err != nil {
		return fmt.Errorf("create test table: %w", err)
	}

	if _, err := p.db.Exec(`
		CREATE INDEX test_ordered_field_idx ON test(ordered_field)`,
	); err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	return nil
}

func (p *PostgresQueryProvider) PopulateTestData(rowsCount int) error {
	trx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	//nolint:errcheck
	defer trx.Rollback()

	stmt, err := trx.Prepare(`INSERT INTO test(ordered_field) VALUES($1)`)
	if err != nil {
		return fmt.Errorf("tx prepare: %w", err)
	}

	defer stmt.Close()

	for i := range rowsCount {
		if _, err := stmt.Exec(i + 1); err != nil {
			return fmt.Errorf("exec: %w", err)
		}
	}

	if err := trx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func (p *PostgresQueryProvider) Query(ctx context.Context, offset int, limit int) ([]TestTable, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT
			id,
			ordered_field
		FROM
			test
		ORDER BY
			ordered_field
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
		if err := rows.Scan(&data.ID, &data.OrderedField); err != nil {
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

func (p *PostgresQueryProvider) Count(ctx context.Context) (int, error) {
	var count int

	if err := p.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*)
		FROM
			test
	`).Scan(&count); err != nil {
		return 0, fmt.Errorf("query row context: %w", err)
	}

	return count, nil
}
