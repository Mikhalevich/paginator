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
		pagin = paginator.New(NewSliceProvider(), pageSize)
		page  *paginator.Page[int]
	)

	page, _ = pagin.Page(1)
	printPage(page)

	for page.HasNext() {
		page, _ = pagin.Page(page.Next())

		printPage(page)
	}
}

func printPage(page *paginator.Page[int]) {
	log.Printf("bottom index: %d top index: %d, page number: %d total pages: %d page data: %v",
		page.BottomIndex, page.TopIndex, page.PageNumber, page.PageTotalCount, page.Data)
}

func NewSliceProvider() *paginator.SliceQueryProvider[int] {
	data := make([]int, 0, dataLen)

	for i := range dataLen {
		data = append(data, i+1)
	}

	return paginator.NewSliceQueryProvider(data)
}
