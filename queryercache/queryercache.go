package queryercache

import (
	"context"
	"fmt"
	"time"

	"github.com/Mikhalevich/paginator"
)

const (
	defaultCountCacheTTL = time.Second * 15
)

type QueryerCache[T any] struct {
	queryer paginator.Queryer[T]
	count   value[int]
}

func New[T any](queryer paginator.Queryer[T], opts ...Option) *QueryerCache[T] {
	defaultOptions := options{
		CountTTL: defaultCountCacheTTL,
	}

	for _, o := range opts {
		o(&defaultOptions)
	}

	count := newValue[int](defaultOptions.CountTTL)

	return &QueryerCache[T]{
		queryer: queryer,
		count:   count,
	}
}

func (q *QueryerCache[T]) Count(ctx context.Context) (int, error) {
	val, ok := q.count.Value()
	if ok {
		return val, nil
	}

	count, err := q.queryer.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	q.count.SetValue(count)

	return count, nil
}

func (q *QueryerCache[T]) Query(ctx context.Context, offset int, limit int) ([]T, error) {
	vals, err := q.queryer.Query(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return vals, nil
}
