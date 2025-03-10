package main

import (
	"log"
	"net/http"

	"github.com/paranoiachains/metrics/internal/handlers"
	"github.com/paranoiachains/metrics/internal/storage"
)

func main() {
	storage.Storage.Clear()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/gauge/", handlers.MetricHandler("gauge"))
	mux.HandleFunc("/update/counter/", handlers.MetricHandler("counter"))
	mux.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("bad request")
	})

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		panic(err)
	}
}
