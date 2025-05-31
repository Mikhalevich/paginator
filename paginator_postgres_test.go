package paginator_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Mikhalevich/paginator"
)

const (
	dataLen  = 101
	pageSize = 10
)

type PaginatorPostgres struct {
	*suite.Suite

	pag *paginator.Paginator[int]
}

func TestPaginatorPostgresSuit(t *testing.T) {
	t.Parallel()
	suite.Run(t, &PaginatorPostgres{
		Suite: new(suite.Suite),
	})
}

func (s *PaginatorPostgres) SetupSuite() {
	data := make([]int, 0, dataLen)
	for i := range dataLen {
		data = append(data, i+1)
	}

	s.pag = paginator.New(paginator.NewSliceQueryProvider(data), pageSize)
}

func (s *PaginatorPostgres) TearDownSuite() {
}

func (s *PaginatorPostgres) TearDownTest() {
}

func (s *PaginatorPostgres) TearDownSubTest() {
}
