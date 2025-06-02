package paginator_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Mikhalevich/paginator"
	"github.com/Mikhalevich/paginator/queryercache"
)

//nolint:unparam
func initSlicePaginator(dataLen, pageSize int) *paginator.Paginator[int] {
	data := make([]int, 0, dataLen)
	for i := range dataLen {
		data = append(data, i+1)
	}

	return paginator.New(paginator.NewSliceQueryProvider(data), pageSize)
}

func initMockPaginator(
	t *testing.T,
) (*paginator.Paginator[int], *paginator.MockQueryer[int]) {
	t.Helper()

	var (
		ctrl        = gomock.NewController(t)
		mockQueryer = paginator.NewMockQueryer[int](ctrl)
	)

	return paginator.New(mockQueryer, pageSize), mockQueryer
}

func initMockCachedPaginator(
	t *testing.T,
	cacheOpts ...queryercache.Option,
) (*paginator.Paginator[int], *paginator.MockQueryer[int]) {
	t.Helper()

	var (
		ctrl        = gomock.NewController(t)
		mockQueryer = paginator.NewMockQueryer[int](ctrl)
	)

	return paginator.New(queryercache.New(mockQueryer, cacheOpts...), pageSize), mockQueryer
}

func TestFirstPage(t *testing.T) {
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

func TestLastPage(t *testing.T) {
	t.Parallel()

	pg := initSlicePaginator(101, 10)

	page, err := pg.Page(t.Context(), 11)

	require.NoError(t, err)

	require.ElementsMatch(t, []int{101}, page.Data)
	require.Equal(t, 101, page.BottomIndex)
	require.Equal(t, 101, page.TopIndex)
	require.Equal(t, 10, page.PageSize)
	require.Equal(t, 11, page.PageNumber)
	require.Equal(t, 11, page.PageTotalCount)
}

func TestMiddlePage(t *testing.T) {
	t.Parallel()

	pg := initSlicePaginator(101, 10)

	page, err := pg.Page(t.Context(), 5)

	require.NoError(t, err)

	require.ElementsMatch(t, []int{41, 42, 43, 44, 45, 46, 47, 48, 49, 50}, page.Data)
	require.Equal(t, 41, page.BottomIndex)
	require.Equal(t, 50, page.TopIndex)
	require.Equal(t, 10, page.PageSize)
	require.Equal(t, 5, page.PageNumber)
	require.Equal(t, 11, page.PageTotalCount)
}

func TestZeroSlice(t *testing.T) {
	t.Parallel()

	pg := initSlicePaginator(0, 10)

	page, err := pg.Page(t.Context(), 1)

	require.NoError(t, err)

	require.ElementsMatch(t, nil, page.Data)
	require.Equal(t, 0, page.BottomIndex)
	require.Equal(t, 0, page.TopIndex)
	require.Equal(t, 0, page.PageSize)
	require.Equal(t, 0, page.PageNumber)
	require.Equal(t, 0, page.PageTotalCount)
}

func TestInvalidPageZeroPage(t *testing.T) {
	t.Parallel()

	pg := initSlicePaginator(101, 10)

	page, err := pg.Page(t.Context(), 0)

	require.EqualError(t, err, "invaid page number: 0")
	require.Nil(t, page)
}

func TestInvalidPageNegativePage(t *testing.T) {
	t.Parallel()

	pg := initSlicePaginator(101, 10)

	page, err := pg.Page(t.Context(), -1)

	require.EqualError(t, err, "invaid page number: -1")
	require.Nil(t, page)
}

func TestInvalidPageBigPage(t *testing.T) {
	t.Parallel()

	pg := initSlicePaginator(101, 10)

	page, err := pg.Page(t.Context(), 12)

	require.EqualError(t, err, "invalid page: 12 total pages: 11")
	require.Nil(t, page)
}

func TestCountError(t *testing.T) {
	t.Parallel()

	var (
		pag, mockQueryer = initMockPaginator(t)
		ctx              = t.Context()
	)

	mockQueryer.EXPECT().Count(ctx).Return(0, errors.New("some count error"))

	page, err := pag.Page(ctx, 1)

	require.EqualError(t, err, "query count: some count error")
	require.Nil(t, page)
}

func TestQueryError(t *testing.T) {
	t.Parallel()

	var (
		pag, mockQueryer = initMockPaginator(t)
		ctx              = t.Context()
	)

	gomock.InOrder(
		mockQueryer.EXPECT().Count(ctx).Return(11, nil),
		mockQueryer.EXPECT().Query(ctx, 0, pageSize).Return(nil, errors.New("some query error")),
	)

	page, err := pag.Page(ctx, 1)

	require.EqualError(t, err, "query data: some query error")
	require.Nil(t, page)
}

func TestQueryCacheCount(t *testing.T) {
	t.Parallel()

	var (
		pag, mockQueryer = initMockCachedPaginator(t, queryercache.WithCountTTL(time.Minute))
		ctx              = t.Context()
	)

	gomock.InOrder(
		mockQueryer.EXPECT().Count(ctx).Return(3, nil),
		mockQueryer.EXPECT().Query(ctx, 0, pageSize).Return([]int{1, 2, 3}, nil),
		mockQueryer.EXPECT().Query(ctx, 0, pageSize).Return([]int{1, 2, 3}, nil),
	)

	testFlow := func() {
		page, err := pag.Page(ctx, 1)

		require.NoError(t, err)

		require.ElementsMatch(t, []int{1, 2, 3}, page.Data)
		require.Equal(t, 1, page.BottomIndex)
		require.Equal(t, 3, page.TopIndex)
		require.Equal(t, 10, page.PageSize)
		require.Equal(t, 1, page.PageNumber)
		require.Equal(t, 1, page.PageTotalCount)
	}

	testFlow()
	testFlow()
}
