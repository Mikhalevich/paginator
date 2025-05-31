package paginator

import "context"

type SliceQueryProvider[T any] struct {
	Data []T
}

func NewSliceQueryProvider[T any](s []T) *SliceQueryProvider[T] {
	return &SliceQueryProvider[T]{
		Data: s,
	}
}

func (s *SliceQueryProvider[T]) Query(ctx context.Context, offset int, limit int) ([]T, error) {
	limit = offset + limit

	if limit > len(s.Data) {
		limit = len(s.Data)
	}

	return s.Data[offset:limit], nil
}

func (s *SliceQueryProvider[T]) Count(ctx context.Context) (int, error) {
	return len(s.Data), nil
}
