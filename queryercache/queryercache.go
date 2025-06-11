package queryercache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Mikhalevich/paginator"
	"github.com/Mikhalevich/paginator/queryercache/metrics"
)

const (
	defaultCountCacheTTL = time.Second * 30
	defaultQueryCacheTTL = time.Second * 30
)

// CacheMetrics specify interface for metrics methods.
type CacheMetrics interface {
	CountIncrement(cached bool)
	QueryIncrement(cached bool)
}

// QueryerCache implementing cache for paginator.Queryer interface.
type QueryerCache[T any] struct {
	queryer paginator.Queryer[T]

	count    value[int]
	countMtx sync.RWMutex

	query    keyValue[[]T]
	queryMtx sync.RWMutex

	metrics CacheMetrics
}

// New conscturcts new QueryerCache.
func New[T any](queryer paginator.Queryer[T], opts ...Option) *QueryerCache[T] {
	defaultOptions := options{
		CountTTL: defaultCountCacheTTL,
		QueryTTL: defaultQueryCacheTTL,
		Metrics:  metrics.NewNoop(),
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
		metrics: defaultOptions.Metrics,
	}
}

func (q *QueryerCache[T]) countValue(withLock bool) (int, bool) {
	if withLock {
		q.countMtx.RLock()
		defer q.countMtx.RUnlock()
	}

	return q.count.Value()
}

// Count returns count value from cache if available and not expired.
// otherwise returns value from queryer.Count and update cache value.
func (q *QueryerCache[T]) Count(ctx context.Context) (int, error) {
	val, cached, err := q.countValueAndUpdateCache(ctx)
	if err != nil {
		return 0, fmt.Errorf("count value and update cache: %w", err)
	}

	q.metrics.CountIncrement(cached)

	return val, nil
}

// countValueAndUpdateCache returns count value and flag specified is it from cache or not.
// call queryer.Count and update cache value if cache is expired.
func (q *QueryerCache[T]) countValueAndUpdateCache(ctx context.Context) (int, bool, error) {
	//nolint:varnamelen
	val, ok := q.countValue(true)
	if ok {
		return val, true, nil
	}

	q.countMtx.Lock()
	defer q.countMtx.Unlock()

	val, ok = q.countValue(false)
	if ok {
		return val, true, nil
	}

	count, err := q.queryer.Count(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("count: %w", err)
	}

	q.count.SetValue(count)

	return count, false, nil
}

// Query returns cached data value if available and not expired.
// otherwise returns value from queryer.Query and update cache value.
func (q *QueryerCache[T]) Query(ctx context.Context, offset int, limit int) ([]T, error) {
	val, cached, err := q.queryValueAndUpdateCache(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("query value and update cache: %w", err)
	}

	q.metrics.QueryIncrement(cached)

	return val, nil
}

// queryValueAndUpdateCache returns query value and flag specified is it from cache or not.
// call queryer.Query and update cache value if cache is expired.
func (q *QueryerCache[T]) queryValueAndUpdateCache(
	ctx context.Context,
	offset int,
	limit int,
) ([]T, bool, error) {
	key := makeQueryKey(offset, limit)

	vals, ok := q.queryValue(key)
	if ok {
		return vals, true, nil
	}

	vals, err := q.queryer.Query(ctx, offset, limit)
	if err != nil {
		return nil, false, fmt.Errorf("query: %w", err)
	}

	q.setQueryValue(key, vals)

	return vals, false, nil
}

// queryValue returns query cache value and expiration flag.
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
