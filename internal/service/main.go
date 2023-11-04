package service

import (
	"go/ast"
	"net/http"

	"devchallenge.it/spreadsheet/internal/formula/parser"
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

	r.HandleFunc("/{sheet_id}/{cell_id}", CorsHandler).Methods(http.MethodOptions)
	r.HandleFunc("/{sheet_id}", CorsHandler).Methods(http.MethodOptions)
	r.Use(mux.CORSMethodMiddleware(r))

	return r
}

func CorsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func IsVariable(value string) bool {
	tr, err := parser.ParseExpr(value, "")
	if err != nil {
		return false
	}

	_, ok := tr.(*ast.Ident)

	return ok
}
