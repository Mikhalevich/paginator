package paginator

import (
	"time"
)

type options struct {
	IsQueryCountCachable bool
	QueryCountCacheTTL   time.Duration
}

type Option func(opt *options)

func WithQueryCountCache(ttl time.Duration) Option {
	return func(opt *options) {
		opt.IsQueryCountCachable = true
		opt.QueryCountCacheTTL = ttl
	}
}
