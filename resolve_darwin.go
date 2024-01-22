package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getMBitPathOnCurrentPlatform() (string, error) {
	mounted, err := os.ReadDir("/Volumes")
	if err != nil {
		return "", fmt.Errorf("failed to list mounted volumes: %w", err)
	}

	var path string

	for _, entry := range mounted {
		n := entry.Name()
		if strings.Contains(n, "MICROBIT") {
			path = filepath.Join("/Volumes", n)
			break
		}
	}

	if path == "" {
		return "", errNotDetected
	}
	return path, nil
}
