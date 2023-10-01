package service

import (
	"encoding/json"
	"log"
	"net/http"

	"devchallenge.it/spreadsheet/internal/formula"
	"github.com/gorilla/mux"
)

type UpsertPayload struct {
	Value string
}

func (s *Service) upsert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sheetId := vars["sheet_id"]
	cellId := vars["cell_id"]

	if !formula.IsVariable(cellId) {
		log.Printf("Cell ID %q is not valid variable", cellId)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload UpsertPayload
	err := NewJsonDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("Body decode error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	solver := formula.NewSolver(s.dao, sheetId)
	solver.SetCell(cellId, payload.Value)
	result, value, formulaError, err := solver.Solve(cellId)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var responseStatus int
	var errorMsg *string

	if formulaError == nil {
		if err := s.dao.SetCell(sheetId, cellId, payload.Value); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		responseStatus = http.StatusCreated
	} else {
		responseStatus = http.StatusUnprocessableEntity
		errorMsg = new(string)
		*errorMsg = formulaError.Error()
	}

	resp := CellResponse{
		Value:  value,
		Result: result,
		Error:  errorMsg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseStatus)
	json.NewEncoder(w).Encode(&resp)
}
