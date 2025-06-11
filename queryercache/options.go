package queryercache

import (
	"time"
)

type options struct {
	CountTTL time.Duration
	QueryTTL time.Duration
	Metrics  CacheMetrics
}

// Option specify option for QueryerCache.
type Option func(opt *options)

// QueryerCache count cache ttl (default 30 seconds).
func WithCountTTL(ttl time.Duration) Option {
	return func(opt *options) {
		opt.CountTTL = ttl
	}
}

// WithQueryTTL query cache ttl (default 30 seconds).
func WithQueryTTL(ttl time.Duration) Option {
	return func(opt *options) {
		opt.QueryTTL = ttl
	}
}

// WithMetrics metrics provider for QueryerCache (default noop impl).
func WithMetrics(m CacheMetrics) Option {
	return func(opt *options) {
		opt.Metrics = m
	}
}
