package paginator

type Page[T any] struct {
	Data       []T
	Count      int
	Offset     int
	PageSize   int
	PageNumber int
	PageCount  int
}

func (p *Page[T]) HasNext() bool {
	return p.PageNumber < p.PageCount
}

func (p *Page[T]) Next() int {
	if p.HasNext() {
		return p.PageNumber + 1
	}

	return p.PageCount
}

func (p *Page[T]) HasPrevious() bool {
	return p.PageNumber > 1
}

func (p *Page[T]) Previous() int {
	if p.HasPrevious() {
		return p.PageNumber - 1
	}

	return 1
}
