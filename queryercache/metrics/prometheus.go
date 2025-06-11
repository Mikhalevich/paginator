package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus struct for prometheus metrics implementation.
// specify four metrics:
// paginator_cache_count_total - total number of count calls.
// paginator_cache_count_hit_total - count cache hit number.
// paginator_cache_query_total - total number of query calls.
// paginator_cache_query_hit_total - query cache hit number.
type Prometheus struct {
	countCounterTotal prometheus.Counter
	countCounterHit   prometheus.Counter

	queryCouterTotal prometheus.Counter
	queryCounterHit  prometheus.Counter
}

// NewPrometheus constructs new Prometheus.
func NewPrometheus() *Prometheus {
	var (
		countTotal = promauto.NewCounter(prometheus.CounterOpts{
			Name: "paginator_cache_count_total",
			Help: "The total number of count processed events",
		})

		countHit = promauto.NewCounter(prometheus.CounterOpts{
			Name: "paginator_cache_count_hit_total",
			Help: "The cache hit number of count events",
		})

		queryTotal = promauto.NewCounter(prometheus.CounterOpts{
			Name: "paginator_cache_query_total",
			Help: "The total number of query processed events",
		})

		queryHit = promauto.NewCounter(prometheus.CounterOpts{
			Name: "paginator_cache_query_hit_total",
			Help: "The cache hit number of query events",
		})
	)

	return &Prometheus{
		countCounterTotal: countTotal,
		countCounterHit:   countHit,
		queryCouterTotal:  queryTotal,
		queryCounterHit:   queryHit,
	}
}

// CountIncrement increment total and cache hit count calls.
func (p *Prometheus) CountIncrement(cached bool) {
	p.countCounterTotal.Inc()

	if cached {
		p.countCounterHit.Inc()
	}
}

// QueryIncrement increment total and cache hit query calls.
func (p *Prometheus) QueryIncrement(cached bool) {
	p.queryCouterTotal.Inc()

	if cached {
		p.queryCounterHit.Inc()
	}
}
