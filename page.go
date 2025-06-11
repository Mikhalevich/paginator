package paginator

// Page represents single page information.
type Page[T any] struct {
	Data           []T
	BottomIndex    int
	TopIndex       int
	PageSize       int
	PageNumber     int
	PageTotalCount int
}

// HasNext returns true if next page is available.
func (p *Page[T]) HasNext() bool {
	return p.PageNumber < p.PageTotalCount
}

// Next returns next page number.
func (p *Page[T]) Next() int {
	if p.HasNext() {
		return p.PageNumber + 1
	}

	return p.PageTotalCount
}

// HasPrevious returns true if previous page is available.
func (p *Page[T]) HasPrevious() bool {
	return p.PageNumber > 1
}

// Previous returns previous page number.
func (p *Page[T]) Previous() int {
	if p.HasPrevious() {
		return p.PageNumber - 1
	}

	return 1
}
