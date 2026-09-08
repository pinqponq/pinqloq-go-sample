package main

import (
	"os"
	"strings"
)

func loadDotEnvIfPresent(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		os.Setenv(key, strings.TrimSpace(value))
	}
}
