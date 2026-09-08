package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
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

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3300"
	}

	mux := newMux("public")
	addr := "127.0.0.1:" + port

	log.Printf("Pinqloq sample ready: http://%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
