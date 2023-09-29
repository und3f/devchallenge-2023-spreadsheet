package main

import (
	"log"
	"net/http"
	"os"

	"devchallenge.it/spreadsheet/internal/model"
	"devchallenge.it/spreadsheet/internal/service"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

const ListenAddr = ":8080"

func main() {
	var redisAddr = os.Getenv("REDIS_ADDR")

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	dao := model.NewDao(rdb)

	router := mux.NewRouter()
	apiV1Router := router.PathPrefix("/api/v1").Subrouter()

	service.NewService(apiV1Router, dao)

	http.Handle("/", router)

	log.Printf("Starting webserver at %q", ListenAddr)
	if err := http.ListenAndServe(ListenAddr, nil); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}
