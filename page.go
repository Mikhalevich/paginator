package paginator

type Page[T any] struct {
	Data           []T
	BottomIndex    int
	TopIndex       int
	PageSize       int
	PageNumber     int
	PageTotalCount int
}

func (p *Page[T]) HasNext() bool {
	return p.PageNumber < p.PageTotalCount
}

func (p *Page[T]) Next() int {
	if p.HasNext() {
		return p.PageNumber + 1
	}

	return p.PageTotalCount
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
