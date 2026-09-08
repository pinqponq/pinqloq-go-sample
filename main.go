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

	pinqloq "github.com/pinqponq/pinqloq-backend/sdk/pinqloq-go"
)

var scenarioMessages = map[int]string{
	http.StatusOK:                  "ok",
	http.StatusBadRequest:          "bad request",
	http.StatusUnauthorized:        "unauthorized",
	http.StatusNotFound:            "not found",
	http.StatusInternalServerError: "internal server error",
}

var (
	pinqloqClient        *pinqloq.Client
	httpCollectionName   string
	manualCollectionName string
)

func newMux(publicDir string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /demo/http/{status}", handleHTTPScenario)
	mux.HandleFunc("GET /api/config", handleConfig)
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

func setupPinqloq() {
	secretKey := os.Getenv("PINQLOQ_SECRET_KEY")
	if secretKey == "" {
		return
	}

	httpCollectionName = os.Getenv("PINQLOQ_HTTP_COLLECTION")
	manualCollectionName = os.Getenv("PINQLOQ_MANUAL_COLLECTION")

	client, err := pinqloq.New(pinqloq.Options{
		SecretKey:             secretKey,
		APILogsCollectionName: httpCollectionName,
		DeviceIdentifier:      deviceIdentifier,
	})
	if err != nil {
		log.Printf("Pinqloq: failed to initialize client: %v", err)
		return
	}

	pinqloqClient = client
}

func handlerWithMiddleware(mux *http.ServeMux) http.Handler {
	if pinqloqClient == nil {
		return mux
	}

	middleware := pinqloqClient.Middleware(pinqloq.RequestLoggingOptions{
		ExcludePaths: []string{"/style.css", "/app.js", "/api/config"},
		RedactFields: []string{"taxNumber"},
		RedactPaths:  []string{"/demo/redaction/endpoint"},
	})

	return middleware(mux)
}

func main() {
	loadDotEnvIfPresent(".env")
	setupPinqloq()

	if pinqloqClient != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := pinqloqClient.Shutdown(ctx); err != nil {
				log.Printf("Pinqloq: shutdown error: %v", err)
			}
		}()
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3300"
	}

	mux := newMux("public")
	handler := handlerWithMiddleware(mux)
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
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}
