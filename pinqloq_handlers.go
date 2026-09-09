package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"

	pinqloq "github.com/pinqponq/pinqloq-go-sdk"
)

const deviceIdentifier = "go-sample"

var logLevelsByName = map[string]pinqloq.LogLevel{
	"debug":       pinqloq.LogLevelDebug,
	"information": pinqloq.LogLevelInformation,
	"warning":     pinqloq.LogLevelWarning,
	"error":       pinqloq.LogLevelError,
	"fatal":       pinqloq.LogLevelFatal,
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"configured":       pinqloqClient != nil,
		"httpCollection":   httpCollectionName,
		"manualCollection": manualCollectionName,
	})
}

func handleManualEvent(w http.ResponseWriter, r *http.Request) {
	if pinqloqClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "pinqloq not configured"})
		return
	}

	levelName := r.PathValue("level")
	level, ok := logLevelsByName[levelName]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported level"})
		return
	}

	_, err := pinqloqClient.Logger().Enqueue(pinqloq.LogEntry{
		Event:            "go_sample.manual_event",
		DeviceIdentifier: deviceIdentifier,
		LogLevel:         level,
		LogSourceType:    pinqloq.LogSourceTypeBackend,
		CollectionName:   manualCollectionName,
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
	if pinqloqClient == nil {
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
