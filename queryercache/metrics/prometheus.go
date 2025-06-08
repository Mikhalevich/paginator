package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Prometheus struct {
	countCounterTotal prometheus.Counter
	countCounterHit   prometheus.Counter

	queryCouterTotal prometheus.Counter
	queryCounterHit  prometheus.Counter
}

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

func (p *Prometheus) CountIncrement(cached bool) {
	p.countCounterTotal.Inc()

	if cached {
		p.countCounterHit.Inc()
	}
}

func (p *Prometheus) QueryIncrement(cached bool) {
	p.queryCouterTotal.Inc()

	if cached {
		p.queryCounterHit.Inc()
	}
}
