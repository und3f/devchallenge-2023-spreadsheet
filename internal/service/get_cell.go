package service

import (
	"encoding/json"
	"log"
	"net/http"

	"devchallenge.it/spreadsheet/internal/formula"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

type CellResponse struct {
	Value  string `json:"value"`
	Result string `json:"result"`
}

func (s *Service) getCell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sheetId := vars["sheet_id"]
	cellId := vars["cell_id"]

	if !formula.IsVariable(cellId) {
		log.Printf("Cell ID %q is not valid variable", cellId)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	solver := formula.NewSolver(s.dao, sheetId)
	result, value, _, err := solver.Solve(cellId)
	if err != nil {
		if err == redis.Nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		log.Printf("Failed to get cell: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := CellResponse{
		Value:  value,
		Result: result,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&resp)
}
