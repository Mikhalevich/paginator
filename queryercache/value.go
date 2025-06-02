package queryercache

import (
	"time"
)

type value[T any] struct {
	Val           T
	TTL           time.Duration
	LastAccesTime time.Time
}

func newValue[T any](ttl time.Duration) value[T] {
	return value[T]{
		TTL: ttl,
	}
}

func (v *value[T]) Value() (T, bool) {
	if time.Since(v.LastAccesTime) <= v.TTL {
		return v.Val, true
	}

	return v.Val, false
}

func (v *value[T]) SetValue(val T) {
	v.Val = val
	v.LastAccesTime = time.Now()
}
