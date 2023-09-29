package service

import (
	"net/http"

	"devchallenge.it/spreadsheet/internal/model"
	"github.com/gorilla/mux"
)

type Service struct {
	dao *model.Dao
}

func NewService(r *mux.Router, dao *model.Dao) *Service {
	s := &Service{dao: dao}
	s.Mount(r)
	return s
}

func (s *Service) Mount(r *mux.Router) *mux.Router {
	r.HandleFunc("/{sheet_id}/{cell_id}",
		func(w http.ResponseWriter, r *http.Request) {
			s.upsert(w, r)
		}).Methods(http.MethodPost)

	r.HandleFunc("/{sheet_id}/{cell_id}",
		func(w http.ResponseWriter, r *http.Request) {
			s.getCell(w, r)
		}).Methods(http.MethodGet)

	r.HandleFunc("/{sheet_id}",
		func(w http.ResponseWriter, r *http.Request) {
			s.getSpreadsheet(w, r)
		}).Methods(http.MethodGet)

	return r
}
