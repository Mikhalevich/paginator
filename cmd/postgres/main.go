package main

import (
	"log"
	"os"

	"github.com/Mikhalevich/paginator"
)

const (
	dataLen  = 1024
	pageSize = 10
)

func main() {
	p := paginator.New(NewData(), pageSize)

	page, err := p.Page(1)
	if err != nil {
		log.Printf("first page error: %v", err)
		os.Exit(1)
	}

	log.Println(page.Data)
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
	return t.Data[offset : offset+limit], nil
}

func (t *TestData) Count() (int, error) {
	return len(t.Data), nil
}
