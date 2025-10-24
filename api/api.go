// Package api exposes http api for person
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ---------- HTTP middleware for instrumentation ----------
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	buf        *bytes.Buffer
}

func ExposeApi() {
	initMetrics()

	// Create store and seed one sample
	s := newStore()
	s.create(Person{FirstName: "Asraful", LastName: "haque", Email: "asraf@peek8.io"})

	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/person/list", listPersonsHandler(s)).Methods("GET")
	r.HandleFunc("/person", createPersonHandler(s)).Methods("POST")
	r.HandleFunc("/person/{id}", getPersonHandler(s)).Methods("GET")
	r.HandleFunc("/person/{id}", updatePersonHandler(s)).Methods("PUT")
	r.HandleFunc("/person/{id}", deletePersonHandler(s)).Methods("DELETE")

	// metrics route (prometheus)
	r.Handle("/metrics", promhttp.Handler())

	// simple healthcheck
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	serverAddr := ":8080"
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	// start system metrics collection
	stop := make(chan struct{})
	go collectSystemMetricsLoop(stop)

	log.Printf("Server listening on %s", serverAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func instrument(path string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		httpRequestsInProgress.WithLabelValues(method).Inc()
		start := time.Now()

		rr := newResponseRecorder(w)
		handler(rr, r)

		duration := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
		httpRequestsTotal.WithLabelValues(fmt.Sprint(rr.statusCode), method).Inc()
		httpRequestsInProgress.WithLabelValues(method).Dec()
	}
}

// ---------- Handlers ----------
func listPersonsHandler(s *store) http.HandlerFunc {
	return instrument("/person/list", func(w http.ResponseWriter, r *http.Request) {
		startParam := r.URL.Query().Get("start")
		start := 0
		if startParam != "" {
			start, _ = strconv.Atoi(startParam)
		}

		list := s.list(start, 20)
		personStoreCount.Set(float64(len(list)))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	})
}

func getPersonHandler(s *store) http.HandlerFunc {
	return instrument("/person/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		p, ok := s.get(id)
		if !ok {
			http.Error(w, "person not found", http.StatusNotFound)
			personNotFoundTotal.Inc()
			return
		}
		personStoreCount.Set(float64(s.count()))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	})
}

func createPersonHandler(s *store) http.HandlerFunc {
	return instrument("/person", func(w http.ResponseWriter, r *http.Request) {
		var p Person

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}
		personPayloadSize.Observe(float64(len(body))) // record size

		if err := json.Unmarshal(body, &p); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if p.FirstName == "" || p.LastName == "" {
			http.Error(w, "first_name and last_name required", http.StatusBadRequest)
			return
		}

		if s.count() >= maxStoreLimits {
			http.Error(w, "database full: cannot accept more records", http.StatusInsufficientStorage)
			return
		}

		created := s.create(p)

		personCreatedTotal.Inc()
		personStoreCount.Set(float64(s.count()))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	})
}

func updatePersonHandler(s *store) http.HandlerFunc {
	return instrument("/person/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var p Person
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		updated, ok := s.update(id, p)
		if !ok {
			http.Error(w, "person not found", http.StatusNotFound)
			return
		}
		personStoreCount.Set(float64(s.count()))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
	})
}

func deletePersonHandler(s *store) http.HandlerFunc {
	return instrument("/person/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		ok := s.delete(id)
		if !ok {
			http.Error(w, "person not found", http.StatusNotFound)
			return
		}
		personStoreCount.Set(float64(s.count()))
		personDeletedTotal.Inc()

		w.WriteHeader(http.StatusNoContent)
	})
}
