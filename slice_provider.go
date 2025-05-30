package paginator

type SliceQueryProvider[T any] struct {
	Data []T
}

func NewSliceQueryProvider[T any](s []T) *SliceQueryProvider[T] {
	return &SliceQueryProvider[T]{
		Data: s,
	}
}

func (s *SliceQueryProvider[T]) Query(offset int, limit int) ([]T, error) {
	limit = offset + limit

	if limit > len(s.Data) {
		limit = len(s.Data)
	}

	return s.Data[offset:limit], nil
}

func (s *SliceQueryProvider[T]) Count() (int, error) {
	return len(s.Data), nil
}
