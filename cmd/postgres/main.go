package main

import (
	"log"

	"github.com/Mikhalevich/paginator"
)

const (
	dataLen  = 101
	pageSize = 10
)

func main() {
	var (
		pagin = paginator.New(NewData(), pageSize)
		page  = &paginator.Page[int]{
			PageTotalCount: pageSize,
		}
	)

	for page.HasNext() {
		page, _ = pagin.Page(page.Next())

		log.Printf("bottom index: %d top index: %d, page number: %d total pages: %d page data: %v",
			page.BottomIndex, page.TopIndex, page.PageNumber, page.PageTotalCount, page.Data)
	}
}

type TestData struct {
	Data []int
}

func NewData() *TestData {
	data := make([]int, 0, dataLen)

	for i := range dataLen {
		data = append(data, i+1)
	}

	return &TestData{
		Data: data,
	}
}

func (t *TestData) Query(offset int, limit int) ([]int, error) {
	limit = offset + limit

	if limit > len(t.Data) {
		limit = len(t.Data)
	}

	return t.Data[offset:limit], nil
}

func (t *TestData) Count() (int, error) {
	return len(t.Data), nil
}
