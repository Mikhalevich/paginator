package queryerslice

import (
	"context"
)

// QueyrerSlice implementation of paginator.Queryer for slice data.
type QueryerSlice[T any] struct {
	Data []T

	opts *options
}

// New construct new QueryerSlice.
func New[T any](data []T, opts ...Option) *QueryerSlice[T] {
	var defaultOptions options

	for _, v := range opts {
		v(&defaultOptions)
	}

	return &QueryerSlice[T]{
		Data: data,
		opts: &defaultOptions,
	}
}

// Query returns subslice from base slice according offset and limit params.
// if WithCopy option is specified returns copy of subslice.
func (s *QueryerSlice[T]) Query(ctx context.Context, offset int, limit int) ([]T, error) {
	endIndex := offset + limit

	if s.opts.CopySlice {
		sliceCopy := make([]T, endIndex-offset)

		copy(sliceCopy, s.Data[offset:endIndex])

		return sliceCopy, nil
	}

	return s.Data[offset:endIndex], nil
}

// Count returns length of internal slice data.
func (s *QueryerSlice[T]) Count(ctx context.Context) (int, error) {
	return len(s.Data), nil
}
