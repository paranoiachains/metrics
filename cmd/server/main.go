package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/paranoiachains/metrics/internal/handlers"
	"github.com/paranoiachains/metrics/internal/storage"
)

var S = storage.NewMemStorage()

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
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	fmt.Println(m)
	fmt.Println("asdasd")
}
