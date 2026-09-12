package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	pinqloq "github.com/pinqponq/pinqloq-go-sdk/v2"
)

var logLevelsByName = map[string]pinqloq.LogLevel{
	"debug":       pinqloq.LogLevelDebug,
	"information": pinqloq.LogLevelInformation,
	"warning":     pinqloq.LogLevelWarning,
	"error":       pinqloq.LogLevelError,
	"fatal":       pinqloq.LogLevelFatal,
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	client, httpCollection, manualCollection, _ := currentSession.read()

	writeJSON(w, http.StatusOK, map[string]any{
		"configured":       client != nil,
		"httpCollection":   httpCollection,
		"manualCollection": manualCollection,
	})
}

func handleSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SecretKey        string `json:"secretKey"`
		HTTPCollection   string `json:"httpCollection"`
		ManualCollection string `json:"manualCollection"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	secretKey := strings.TrimSpace(body.SecretKey)
	httpCollection := strings.TrimSpace(body.HTTPCollection)
	manualCollection := strings.TrimSpace(body.ManualCollection)

	if secretKey == "" || httpCollection == "" || manualCollection == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "secretKey, httpCollection and manualCollection are all required"})
		return
	}

	if httpCollection == manualCollection {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "httpCollection and manualCollection must be different"})
		return
	}

	if err := currentSession.configure(secretKey, httpCollection, manualCollection); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"configured":       true,
		"httpCollection":   httpCollection,
		"manualCollection": manualCollection,
	})
}

func handleManualEvent(w http.ResponseWriter, r *http.Request) {
	client, _, manualCollection, _ := currentSession.read()
	if client == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "pinqloq not configured"})
		return
	}

	levelName := r.PathValue("level")
	level, ok := logLevelsByName[levelName]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported level"})
		return
	}

	_, err := client.Enqueue(pinqloq.LogEntry{
		Event:            "go_sample.manual_event",
		DeviceIdentifier: deviceIdentifier,
		LogLevel:         level,
		LogSourceType:    pinqloq.LogSourceTypeBackend,
		CollectionName:   manualCollection,
		Metadata:         map[string]string{"triggeredFrom": "test-lab"},
		Detail:           map[string]string{"note": "Synthetic manual event from the Go sample test lab."},
	}, nil, nil)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "queued", "level": levelName})
}

func handleRedactionCapture(w http.ResponseWriter, r *http.Request) {
	client, _, _, _ := currentSession.read()
	if client == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "pinqloq not configured"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	fields := make([]string, 0, len(payload))
	for key := range payload {
		fields = append(fields, key)
	}
	sort.Strings(fields)

	writeJSON(w, http.StatusOK, map[string]any{"status": "captured", "fields": fields})
}
