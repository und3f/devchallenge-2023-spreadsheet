package service

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

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
}
