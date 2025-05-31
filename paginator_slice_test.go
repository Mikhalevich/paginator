package paginator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Mikhalevich/paginator"
)

func initSlicePaginator(dataLen, pageSize int) *paginator.Paginator[int] {
	data := make([]int, 0, dataLen)
	for i := range dataLen {
		data = append(data, i+1)
	}

	return paginator.New(paginator.NewSliceQueryProvider(data), pageSize)
}

func TestFirstChunk(t *testing.T) {
	t.Parallel()

	pg := initSlicePaginator(101, 10)

	page, err := pg.Page(t.Context(), 1)

	require.NoError(t, err)

	require.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, page.Data)
	require.Equal(t, 1, page.BottomIndex)
	require.Equal(t, 10, page.TopIndex)
	require.Equal(t, 10, page.PageSize)
	require.Equal(t, 1, page.PageNumber)
	require.Equal(t, 11, page.PageTotalCount)
}
