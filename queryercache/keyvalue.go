package queryercache

import (
	"time"
)

type keyValue[T any] struct {
	Val           map[string]T
	TTL           time.Duration
	LastAccesTime time.Time
}

func newKeyValue[T any](ttl time.Duration) keyValue[T] {
	return keyValue[T]{
		Val: make(map[string]T),
		TTL: ttl,
	}
}

func (v *keyValue[T]) Value(key string) (T, bool) {
	var defaultVal T

	if v.TTL == 0 {
		return defaultVal, false
	}

	if time.Since(v.LastAccesTime) > v.TTL {
		return defaultVal, false
	}

	if val, ok := v.Val[key]; ok {
		return val, true
	}

	return defaultVal, false
}

func (v *keyValue[T]) SetValue(key string, val T) {
	if v.TTL == 0 {
		return
	}

	v.Val[key] = val

	if time.Since(v.LastAccesTime) > v.TTL {
		v.LastAccesTime = time.Now()
	}
}
