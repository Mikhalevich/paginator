package paginator_test

import (
	"errors"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Mikhalevich/paginator"
	"github.com/Mikhalevich/paginator/queryercache"
	"github.com/Mikhalevich/paginator/queryerslice"
)

const (
	sqlTestRows      = 101
	sqlBenchmartRows = 10001
	dataLen          = 101
	pageSize         = 10
)

func initSlicePaginator(dataLen, pageSize int, opts ...queryerslice.Option) *paginator.Paginator[int] {
	data := make([]int, 0, dataLen)
	for i := range dataLen {
		data = append(data, i+1)
	}

	return paginator.New(queryerslice.New(data, opts...), pageSize)
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

	testCase := func(t *testing.T, pg *paginator.Paginator[int]) {
		t.Helper()

		page, err := pg.Page(t.Context(), 1)

		require.NoError(t, err)

		require.ElementsMatch(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, page.Data)
		require.Equal(t, 1, page.BottomIndex)
		require.Equal(t, 10, page.TopIndex)
		require.Equal(t, 10, page.PageSize)
		require.Equal(t, 1, page.PageNumber)
		require.Equal(t, 11, page.PageTotalCount)
	}

	t.Run("inplace", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10))
	})

	t.Run("copy", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10, queryerslice.WithCopy()))
	})
}

func TestLastPage(t *testing.T) {
	t.Parallel()

	testCase := func(t *testing.T, pg *paginator.Paginator[int]) {
		t.Helper()

		page, err := pg.Page(t.Context(), 11)

		require.NoError(t, err)

		require.ElementsMatch(t, []int{101}, page.Data)
		require.Equal(t, 101, page.BottomIndex)
		require.Equal(t, 101, page.TopIndex)
		require.Equal(t, 1, page.PageSize)
		require.Equal(t, 11, page.PageNumber)
		require.Equal(t, 11, page.PageTotalCount)
	}

	t.Run("inplace", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10))
	})

	t.Run("copy", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10, queryerslice.WithCopy()))
	})
}

func TestMiddlePage(t *testing.T) {
	t.Parallel()

	testCase := func(t *testing.T, pg *paginator.Paginator[int]) {
		t.Helper()

		page, err := pg.Page(t.Context(), 5)

		require.NoError(t, err)

		require.ElementsMatch(t, []int{41, 42, 43, 44, 45, 46, 47, 48, 49, 50}, page.Data)
		require.Equal(t, 41, page.BottomIndex)
		require.Equal(t, 50, page.TopIndex)
		require.Equal(t, 10, page.PageSize)
		require.Equal(t, 5, page.PageNumber)
		require.Equal(t, 11, page.PageTotalCount)
	}

	t.Run("inplace", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10))
	})

	t.Run("copy", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10, queryerslice.WithCopy()))
	})
}

func TestZeroSlice(t *testing.T) {
	t.Parallel()

	testCase := func(t *testing.T, pg *paginator.Paginator[int]) {
		t.Helper()

		page, err := pg.Page(t.Context(), 1)

		require.NoError(t, err)

		require.ElementsMatch(t, nil, page.Data)
		require.Equal(t, 0, page.BottomIndex)
		require.Equal(t, 0, page.TopIndex)
		require.Equal(t, 0, page.PageSize)
		require.Equal(t, 0, page.PageNumber)
		require.Equal(t, 0, page.PageTotalCount)
	}

	t.Run("inplace", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(0, 10))
	})

	t.Run("copy", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(0, 10, queryerslice.WithCopy()))
	})
}

func TestInvalidPageZeroPage(t *testing.T) {
	t.Parallel()

	testCase := func(t *testing.T, pg *paginator.Paginator[int]) {
		t.Helper()

		page, err := pg.Page(t.Context(), 0)

		require.EqualError(t, err, "invaid page number: 0")
		require.Nil(t, page)
	}

	t.Run("inplace", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10))
	})

	t.Run("copy", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10, queryerslice.WithCopy()))
	})
}

func TestInvalidPageNegativePage(t *testing.T) {
	t.Parallel()

	testCase := func(t *testing.T, pg *paginator.Paginator[int]) {
		t.Helper()

		page, err := pg.Page(t.Context(), -1)

		require.EqualError(t, err, "invaid page number: -1")
		require.Nil(t, page)
	}

	t.Run("inplace", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10))
	})

	t.Run("copy", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10, queryerslice.WithCopy()))
	})
}

func TestInvalidPageBigPage(t *testing.T) {
	t.Parallel()

	testCase := func(t *testing.T, pg *paginator.Paginator[int]) {
		t.Helper()

		page, err := pg.Page(t.Context(), 12)

		require.EqualError(t, err, "invalid page: 12 total pages: 11")
		require.Nil(t, page)
	}

	t.Run("inplace", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10))
	})

	t.Run("copy", func(t *testing.T) {
		t.Parallel()
		testCase(t, initSlicePaginator(101, 10, queryerslice.WithCopy()))
	})
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

func TestQueryCountCache(t *testing.T) {
	t.Parallel()

	var (
		pag, mockQueryer = initMockCachedPaginator(
			t,
			queryercache.WithCountTTL(time.Minute),
			queryercache.WithQueryTTL(time.Minute),
		)
		ctx = t.Context()
	)

	gomock.InOrder(
		mockQueryer.EXPECT().Count(ctx).Return(3, nil),
		mockQueryer.EXPECT().Query(ctx, 0, 3).Return([]int{1, 2, 3}, nil),
	)

	testFlow := func() {
		page, err := pag.Page(ctx, 1)

		require.NoError(t, err)

		require.ElementsMatch(t, []int{1, 2, 3}, page.Data)
		require.Equal(t, 1, page.BottomIndex)
		require.Equal(t, 3, page.TopIndex)
		require.Equal(t, 3, page.PageSize)
		require.Equal(t, 1, page.PageNumber)
		require.Equal(t, 1, page.PageTotalCount)
	}

	testFlow()
	testFlow()
}

func TestCountCache(t *testing.T) {
	t.Parallel()

	var (
		pag, mockQueryer = initMockCachedPaginator(
			t,
			queryercache.WithCountTTL(time.Minute),
			queryercache.WithQueryTTL(0),
		)
		ctx = t.Context()
	)

	gomock.InOrder(
		mockQueryer.EXPECT().Count(ctx).Return(3, nil),
		mockQueryer.EXPECT().Query(ctx, 0, 3).Return([]int{1, 2, 3}, nil),
		mockQueryer.EXPECT().Query(ctx, 0, 3).Return([]int{1, 2, 3}, nil),
	)

	testFlow := func() {
		page, err := pag.Page(ctx, 1)

		require.NoError(t, err)

		require.ElementsMatch(t, []int{1, 2, 3}, page.Data)
		require.Equal(t, 1, page.BottomIndex)
		require.Equal(t, 3, page.TopIndex)
		require.Equal(t, 3, page.PageSize)
		require.Equal(t, 1, page.PageNumber)
		require.Equal(t, 1, page.PageTotalCount)
	}

	testFlow()
	testFlow()
}

func TestQueryCache(t *testing.T) {
	t.Parallel()

	var (
		pag, mockQueryer = initMockCachedPaginator(
			t,
			queryercache.WithCountTTL(0),
			queryercache.WithQueryTTL(time.Minute),
		)
		ctx = t.Context()
	)

	gomock.InOrder(
		mockQueryer.EXPECT().Count(ctx).Return(3, nil),
		mockQueryer.EXPECT().Query(ctx, 0, 3).Return([]int{1, 2, 3}, nil),
		mockQueryer.EXPECT().Count(ctx).Return(3, nil),
	)

	testFlow := func() {
		page, err := pag.Page(ctx, 1)

		require.NoError(t, err)

		require.ElementsMatch(t, []int{1, 2, 3}, page.Data)
		require.Equal(t, 1, page.BottomIndex)
		require.Equal(t, 3, page.TopIndex)
		require.Equal(t, 3, page.PageSize)
		require.Equal(t, 1, page.PageNumber)
		require.Equal(t, 1, page.PageTotalCount)
	}

	testFlow()
	testFlow()
}

func BenchmarkInplaceSlice(b *testing.B) {
	pag := initSlicePaginator(10001, 50)

	page, err := pag.Page(b.Context(), 1)
	if err != nil {
		b.Fatal("request first page", err)
	}

	pagesCount := page.PageTotalCount

	for b.Loop() {
		//nolint:gosec
		if _, err := pag.Page(b.Context(), rand.Int()%pagesCount+1); err != nil {
			b.Fatal("request page", err)
		}
	}
}

func BenchmarkCopySlice(b *testing.B) {
	pag := initSlicePaginator(10001, 50, queryerslice.WithCopy())

	page, err := pag.Page(b.Context(), 1)
	if err != nil {
		b.Fatal("request first page", err)
	}

	pagesCount := page.PageTotalCount

	for b.Loop() {
		//nolint:gosec
		if _, err := pag.Page(b.Context(), rand.Int()%pagesCount+1); err != nil {
			b.Fatal("request page", err)
		}
	}
}
