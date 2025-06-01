package queryercache

import (
	"context"
	"fmt"
	"time"
)

type Queryer[T any] interface {
	Query(ctx context.Context, offset int, limit int) ([]T, error)
	Count(ctx context.Context) (int, error)
}

type QueryerCache[T any] struct {
	queryer Queryer[T]
	count   value[int]
}

func New[T any](queryer Queryer[T], countTTL time.Duration) *QueryerCache[T] {
	return &QueryerCache[T]{
		queryer: queryer,
		count:   newValue[int](countTTL),
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
