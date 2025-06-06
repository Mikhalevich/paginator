package main

import (
	"context"
	"log"

	"github.com/Mikhalevich/paginator"
	"github.com/Mikhalevich/paginator/queryerslice"
)

const (
	dataLen  = 101
	pageSize = 10
)

func main() {
	var (
		pagin = paginator.New(NewSliceProvider(), pageSize)
		page  *paginator.Page[int]
		ctx   = context.Background()
	)

	page, _ = pagin.Page(ctx, 1)
	printPage(page)

	for page.HasNext() {
		page, _ = pagin.Page(ctx, page.Next())

		printPage(page)
	}
}

func printPage(page *paginator.Page[int]) {
	log.Printf("bottom index: %d top index: %d, page number: %d total pages: %d page data: %v",
		page.BottomIndex, page.TopIndex, page.PageNumber, page.PageTotalCount, page.Data)
}

func NewSliceProvider() *queryerslice.QueryerSlice[int] {
	data := make([]int, 0, dataLen)

	for i := range dataLen {
		data = append(data, i+1)
	}

	return queryerslice.New(data)
}
