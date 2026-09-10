package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var scenarioMessages = map[int]string{
	http.StatusOK:                  "ok",
	http.StatusBadRequest:          "bad request",
	http.StatusUnauthorized:        "unauthorized",
	http.StatusNotFound:            "not found",
	http.StatusInternalServerError: "internal server error",
}

func newMux(publicDir string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /demo/http/{status}", handleHTTPScenario)
	mux.HandleFunc("GET /api/config", handleConfig)
	mux.HandleFunc("POST /api/session", handleSession)
	mux.HandleFunc("POST /demo/manual/{level}", handleManualEvent)
	mux.HandleFunc("POST /demo/redaction/fields", handleRedactionCapture)
	mux.HandleFunc("POST /demo/redaction/endpoint", handleRedactionCapture)
	mux.Handle("/", http.FileServer(http.Dir(publicDir)))
	return mux
}

func handleHTTPScenario(w http.ResponseWriter, r *http.Request) {
	statusCode, err := strconv.Atoi(r.PathValue("status"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
		return
	}

	message, ok := scenarioMessages[statusCode]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unsupported status"})
		return
	}

	writeJSON(w, statusCode, map[string]any{"scenario": statusCode, "message": message})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func dynamicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, _, mw := currentSession.read()
		if mw == nil {
			next.ServeHTTP(w, r)
			return
		}
		mw(next).ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3300"
	}

	handler := dynamicMiddleware(newMux("public"))
	addr := "127.0.0.1:" + port

	server := &http.Server{Addr: addr, Handler: handler}

	go func() {
		log.Printf("Pinqloq sample ready: http://%s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	currentSession.shutdown(ctx)
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}
