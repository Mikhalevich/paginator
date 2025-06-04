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
	if v.TTL == 0 {
		return v.Val, false
	}

	if time.Since(v.LastAccesTime) > v.TTL {
		return v.Val, false
	}

	return v.Val, true
}

func (v *value[T]) SetValue(val T) {
	if v.TTL == 0 {
		return
	}

	v.Val = val
	v.LastAccesTime = time.Now()
}
