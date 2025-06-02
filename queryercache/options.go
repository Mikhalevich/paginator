package queryercache

import (
	"time"
)

type options struct {
	CountTTL time.Duration
}

type Option func(opt *options)

func WithCountTTL(ttl time.Duration) Option {
	return func(opt *options) {
		opt.CountTTL = ttl
	}
}
