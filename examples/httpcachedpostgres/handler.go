package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Mikhalevich/paginator"
)

type Handler struct {
	paginatorProvider *paginator.Paginator[TestTable]
}

func NewHandler(p *paginator.Paginator[TestTable]) *Handler {
	return &Handler{
		paginatorProvider: p,
	}
}

//nolint:varnamelen
func (h *Handler) TestTablePage(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid page number", http.StatusBadRequest)

		return
	}

	page, err := h.paginatorProvider.Page(r.Context(), pageID)
	if err != nil {
		http.Error(w, fmt.Sprintf("paginator error: %s", err.Error()), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t")

	//nolint:musttag
	if err := encoder.Encode(page.Data); err != nil {
		http.Error(w, "encode page data error", http.StatusInternalServerError)

		return
	}
}
