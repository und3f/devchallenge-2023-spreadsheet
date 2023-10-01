package service

import (
	"encoding/json"
	"log"
	"net/http"

	"devchallenge.it/spreadsheet/internal/formula"
	"github.com/gorilla/mux"
)

type CellResponse struct {
	Value  string `json:"value"`
	Result string `json:"result"`

	// Removed for API format compliance
	Error *string `json:"-"`
}

func (s *Service) getCell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sheetId := vars["sheet_id"]
	cellId := vars["cell_id"]

	if !IsVariable(cellId) {
		log.Printf("Cell ID %q is not valid variable", cellId)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	solver := formula.NewSolver(s.dao, sheetId)
	result, value, formulaError, err := solver.Solve(cellId)
	if err != nil {
		log.Printf("Failed to get cell: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if formulaError == formula.NO_SUCH_CELL {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var errorMsg *string
	if formulaError != nil {
		errorMsg = new(string)
		*errorMsg = formulaError.Error()
	}

	resp := CellResponse{
		Value:  value,
		Result: result,
		Error:  errorMsg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&resp)
}
