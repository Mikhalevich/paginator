package queryercache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Mikhalevich/paginator"
)

const (
	defaultCountCacheTTL = time.Second * 30
	defaultQueryCacheTTL = time.Second * 30
)

type QueryerCache[T any] struct {
	queryer paginator.Queryer[T]

	count    value[int]
	countMtx sync.RWMutex

	query    keyValue[[]T]
	queryMtx sync.RWMutex
}

func New[T any](queryer paginator.Queryer[T], opts ...Option) *QueryerCache[T] {
	defaultOptions := options{
		CountTTL: defaultCountCacheTTL,
		QueryTTL: defaultQueryCacheTTL,
	}

	for _, o := range opts {
		o(&defaultOptions)
	}

	var (
		count = newValue[int](defaultOptions.CountTTL)
		query = newKeyValue[[]T](defaultOptions.QueryTTL)
	)

	return &QueryerCache[T]{
		queryer: queryer,
		count:   count,
		query:   query,
	}
}

func (q *QueryerCache[T]) countValue(withLock bool) (int, bool) {
	if withLock {
		q.countMtx.RLock()
		defer q.countMtx.RUnlock()
	}

	return q.count.Value()
}

func (q *QueryerCache[T]) Count(ctx context.Context) (int, error) {
	//nolint:varnamelen
	val, ok := q.countValue(true)
	if ok {
		return val, nil
	}

	q.countMtx.Lock()
	defer q.countMtx.Unlock()

	val, ok = q.countValue(false)
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
	key := makeQueryKey(offset, limit)

	vals, ok := q.queryValue(key)
	if ok {
		return vals, nil
	}

	vals, err := q.queryer.Query(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	q.setQueryValue(key, vals)

	return vals, nil
}

func (q *QueryerCache[T]) queryValue(key string) ([]T, bool) {
	q.queryMtx.RLock()
	defer q.queryMtx.RUnlock()

	vals, ok := q.query.Value(key)
	if ok {
		return vals, true
	}

	return nil, false
}

func (q *QueryerCache[T]) setQueryValue(key string, vals []T) {
	q.queryMtx.Lock()
	defer q.queryMtx.Unlock()

	q.query.SetValue(key, vals)
}

func makeQueryKey(offset int, limit int) string {
	return fmt.Sprintf("%d_%d", offset, limit)
}
