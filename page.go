package paginator

type Page[T any] struct {
	Data     []T
	Count    int
	Offset   int
	PageSize int
	Number   int
}
