package paginator

import (
	"fmt"
)

type Queryer[T any] interface {
	Query(offset int, limit int) ([]T, error)
	Count() (int, error)
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

func (p *Paginator[T]) Page(page int) (*Page[T], error) {
	count, err := p.queryer.Count()
	if err != nil {
		return nil, fmt.Errorf("query count: %w", err)
	}

	if count == 0 {
		return &Page[T]{}, nil
	}

	var (
		offset    = p.pageSize * (page - 1)
		pageCount = p.calculatePageCount(count)
	)

	if offset > count {
		return nil, fmt.Errorf("invalid page: %d total pages: %d", page, pageCount)
	}

	data, err := p.queryer.Query(offset, p.pageSize)
	if err != nil {
		return nil, fmt.Errorf("query data: %w", err)
	}

	return &Page[T]{
		Data:       data,
		Count:      count,
		Offset:     offset,
		PageSize:   p.pageSize,
		PageNumber: page,
		PageCount:  pageCount,
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
