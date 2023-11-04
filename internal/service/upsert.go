package service

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"devchallenge.it/spreadsheet/internal/formula"
	"devchallenge.it/spreadsheet/internal/formula/parser"
	"github.com/gorilla/mux"
)

type UpsertPayload struct {
	Value string
}

func (s *Service) upsert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sheetId := vars["sheet_id"]
	cellId := vars["cell_id"]

	if !IsVariable(cellId) {
		log.Printf("Cell ID %q is not valid variable", cellId)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if strings.Compare(contentType, "application/json") != 0 {
		log.Printf("Upsert invalid content type %s", contentType)
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
		formulaError, err = s.checkDependentFormula(sheetId, cellId, solver)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if formulaError != nil {
			result = "ERROR"
		}
	}

	if formulaError == nil {
		if err := s.dao.SetCell(sheetId, cellId, payload.Value); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := s.dao.AddDependatFormula(sheetId, cellId, parser.FindAllIdentifiers(payload.Value)); err != nil {
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

	if errorMsg == nil {
		s.notifyDependents(sheetId, cellId)
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

func (s *Service) checkDependentFormula(spreadsheet, cellId string, solver *formula.Solver) (formulaError error, err error) {
	log.Printf("check, %s", cellId)
	deps, err := s.dao.GetDependants(spreadsheet, cellId)
	if err != nil {
		return
	}

	for _, depCellId := range deps {
		_, _, formulaError, _ = solver.Solve(depCellId)
		if formulaError != nil {
			return
		}

		formulaError, err = s.checkDependentFormula(spreadsheet, depCellId, solver)
		if formulaError != nil || err != nil {
			return
		}
	}

	return
}

func (s *Service) notifyDependents(spreadsheet, cellId string) {
	deps, err := s.dao.GetDependants(spreadsheet, cellId)
	if err != nil {
		return
	}

	s.dao.NotifyCellChange(spreadsheet, cellId)

	for _, depCellId := range deps {
		s.notifyDependents(spreadsheet, depCellId)
	}
}
