package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Mikhalevich/paginator"
	"github.com/Mikhalevich/paginator/queryercache"
	"github.com/Mikhalevich/paginator/queryercache/metrics"
	"github.com/Mikhalevich/paginator/queryerslice"
)

const (
	dataLen  = 101
	pageSize = 10
)

var (
	//nolint:gochecknoglobals
	paginatorProvider = paginator.New(
		queryercache.New(
			NewSliceProvider(),
			queryercache.WithCountTTL(time.Minute),
			queryercache.WithQueryTTL(time.Minute),
			queryercache.WithMetrics(metrics.NewPrometheus()),
		),
		pageSize,
	)
)

func main() {
	http.HandleFunc("GET /page/{id}/", pageHandler)
	http.Handle("GET /metrics/", promhttp.Handler())

	//nolint:gosec
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("listen and server error: %v", err)
	}
}

//nolint:varnamelen
func pageHandler(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid page number", http.StatusBadRequest)

		return
	}

	page, err := paginatorProvider.Page(r.Context(), pageID)
	if err != nil {
		http.Error(w, "paginator error", http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(page.Data); err != nil {
		http.Error(w, "encode page data error", http.StatusInternalServerError)

		return
	}
}

func NewSliceProvider() *queryerslice.QueryerSlice[int] {
	data := make([]int, 0, dataLen)

	for i := range dataLen {
		data = append(data, i+1)
	}

	return queryerslice.New(data)
}
