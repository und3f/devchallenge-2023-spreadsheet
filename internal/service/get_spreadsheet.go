package service

import (
	"encoding/json"
	"log"
	"net/http"

	"devchallenge.it/spreadsheet/internal/formula"
	"github.com/gorilla/mux"
)

type SpreadsheetResponse map[string]CellResponse

func (s *Service) getSpreadsheet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sheetId := vars["sheet_id"]

	exists, err := s.dao.IsSpreadsheetExists(sheetId)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	keys, err := s.dao.GetSpreadeetKeys(sheetId)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make(SpreadsheetResponse)

	solver := formula.NewSolver(s.dao, sheetId)
	solver.LoadAllKeys()
	for _, cellId := range keys {
		result, value, err := solver.Solve(cellId)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp[cellId] = CellResponse{
			Result: result,
			Value:  value,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&resp)
}
