package service

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"devchallenge.it/spreadsheet/internal/formula"
	"devchallenge.it/spreadsheet/internal/model"
	"github.com/gorilla/mux"
)

type SubsribeResponse struct {
	WebhookUrl string `json:"webhook_url"`
}

func (s *Service) subscribeCell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sheetId := vars["sheet_id"]
	cellId := vars["cell_id"]

	id, err := s.dao.CreateSubscription(sheetId, cellId)
	if err != nil {
		log.Printf("Failed to get next id: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	path, err := s.subscribeRoute.URL("subscribe_id", id)
	if err != nil {
		log.Printf("Create subscription url failure: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	host := r.Host
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	url := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path.String(),
	}

	resp := SubsribeResponse{
		WebhookUrl: url.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&resp)
}

func (s *Service) subscribeHook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subId := vars["subscribe_id"]

	subscriber, err := s.dao.Subscribe(subId)
	if err != nil {
		log.Printf("Failed to get subscription: %v", err)

		if err == model.ERROR_NO_SUBSCRIPTION {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, _ := s.dao.GetSubscription(subId)
	sheetId := data["spreadsheetId"]
	cellId := data["cellId"]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	f, _ := w.(http.Flusher)

	encoder := json.NewEncoder(w)

	for {
		_, err := subscriber.ReceiveMessage(r.Context())
		if err != nil {
			log.Printf("Unexpected error: %v", err)
			return
		}

		solver := formula.NewSolver(s.dao, sheetId)
		result, value, formulaError, err := solver.Solve(cellId)
		if err != nil {
			log.Printf("Failed to get cell: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
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

		encoder.Encode(resp)
		f.Flush()
	}
}
