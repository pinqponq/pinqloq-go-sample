package main

import (
	"context"
	"net/http"
	"sync"
	"time"

	pinqloq "github.com/pinqponq/pinqloq-go-sdk"
)

const deviceIdentifier = "go-sample"

type session struct {
	mu               sync.RWMutex
	client           *pinqloq.Client
	httpCollection   string
	manualCollection string
	middleware       func(http.Handler) http.Handler
}

var currentSession = &session{}

func (s *session) configure(secretKey, httpCollection, manualCollection string) error {
	client, err := pinqloq.New(pinqloq.Options{
		SecretKey:             secretKey,
		APILogsCollectionName: httpCollection,
		DeviceIdentifier:      deviceIdentifier,
	})
	if err != nil {
		return err
	}

	mw := client.Middleware(pinqloq.RequestLoggingOptions{
		ExcludePaths: []string{"/style.css", "/app.js", "/api/config", "/api/session"},
		RedactFields: []string{"taxNumber"},
		RedactPaths:  []string{"/demo/redaction/endpoint"},
	})

	s.mu.Lock()
	previous := s.client
	s.client = client
	s.httpCollection = httpCollection
	s.manualCollection = manualCollection
	s.middleware = mw
	s.mu.Unlock()

	if previous != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = previous.Shutdown(ctx)
	}

	return nil
}

func (s *session) read() (client *pinqloq.Client, httpCollection, manualCollection string, mw func(http.Handler) http.Handler) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.client, s.httpCollection, s.manualCollection, s.middleware
}

func (s *session) shutdown(ctx context.Context) {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client != nil {
		_ = client.Shutdown(ctx)
	}
}
