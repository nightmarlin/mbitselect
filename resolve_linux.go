package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func getMBitPathOnCurrentPlatform() (string, error) {
	var buf bytes.Buffer
	// todo: make this portable. I don't use linux so... good luck :D

	// list mounted drives,
	// sort them in mount order,
	// get the first microbit drive,
	// and retrieve it from the output (3rd space-delimited column)
	cmd := exec.Command("sh", "-c", `mount | sort | grep -m 1 "MICROBIT" | awk '{print $3}'`)
	cmd.Stdout = &buf
	cmd.Stderr = os.DevNull

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("getting mounted devices: %w", err)
	}

	line, err := buf.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("reading mounted devices: %w", err)
	} else if line == "" {
		return "", errNotDetected
	}
	return line, nil
}
