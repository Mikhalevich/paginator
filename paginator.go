package paginator

import (
	"context"
	"fmt"
)

type Queryer[T any] interface {
	Query(ctx context.Context, offset int, limit int) ([]T, error)
	Count(ctx context.Context) (int, error)
}

type Paginator[T any] struct {
	queryer  Queryer[T]
	pageSize int
}

func New[T any](queryer Queryer[T], pageSize int) *Paginator[T] {
	return &Paginator[T]{
		queryer:  queryer,
		pageSize: pageSize,
	}
}

func (p *Paginator[T]) Page(ctx context.Context, page int) (*Page[T], error) {
	if page <= 0 {
		return nil, fmt.Errorf("invaid page number: %d", page)
	}

	count, err := p.queryer.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("query count: %w", err)
	}

	if count == 0 {
		return &Page[T]{}, nil
	}

	var (
		offset         = p.pageSize * (page - 1)
		limit          = p.pageSize
		pageTotalCount = p.calculatePageCount(count)
	)

	if page > pageTotalCount {
		return nil, fmt.Errorf("invalid page: %d total pages: %d", page, pageTotalCount)
	}

	if page == pageTotalCount {
		limit = count % p.pageSize
	}

	data, err := p.queryer.Query(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("query data: %w", err)
	}

	var (
		bottomIndex = offset + 1
		topIndex    = bottomIndex + len(data) - 1
	)

	return &Page[T]{
		Data:           data,
		BottomIndex:    bottomIndex,
		TopIndex:       topIndex,
		PageSize:       limit,
		PageNumber:     page,
		PageTotalCount: pageTotalCount,
	}, nil
}

func (p *Paginator[T]) calculatePageCount(count int) int {
	var (
		fullPageCount = count / p.pageSize
		partPage      = count % p.pageSize
	)

	if partPage > 0 {
		return fullPageCount + 1
	}

	return fullPageCount
}
