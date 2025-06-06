package queryerslice

import (
	"context"
)

type QueryerSlice[T any] struct {
	Data []T

	opts *options
}

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

func (s *QueryerSlice[T]) Query(ctx context.Context, offset int, limit int) ([]T, error) {
	endIndex := offset + limit

	if endIndex > len(s.Data) {
		endIndex = len(s.Data)
	}

	if s.opts.CopySlice {
		sliceCopy := make([]T, endIndex-offset)

		copy(sliceCopy, s.Data[offset:endIndex])

		return sliceCopy, nil
	}

	return s.Data[offset:endIndex], nil
}

func (s *QueryerSlice[T]) Count(ctx context.Context) (int, error) {
	return len(s.Data), nil
}
