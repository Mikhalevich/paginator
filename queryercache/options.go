package queryercache

import (
	"time"
)

type options struct {
	CountTTL time.Duration
	QueryTTL time.Duration
	Metrics  CacheMetrics
}

type Option func(opt *options)

func WithCountTTL(ttl time.Duration) Option {
	return func(opt *options) {
		opt.CountTTL = ttl
	}
}

func WithQueryTTL(ttl time.Duration) Option {
	return func(opt *options) {
		opt.QueryTTL = ttl
	}
}

func WithMetrics(m CacheMetrics) Option {
	return func(opt *options) {
		opt.Metrics = m
	}
}
