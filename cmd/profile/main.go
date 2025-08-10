package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/model"
)

const (
	configPath  = "configs/default.yml"
	profilesDir = "profiles"
	serverURL   = "http://localhost:8080"
	updateURL   = serverURL + "/update/"
)

func main() {
	if err := run(); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: go run cmd/profile/main.go [base|result]")
	}

	command := os.Args[1]

	switch command {
	case "base":
		return captureProfile("base.pprof")
	case "result":
		return captureProfile("result.pprof")
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func captureProfile(filename string) error {
	fmt.Printf("Capturing %s...\n", filename)

	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	if !isServerReady() {
		return fmt.Errorf("server not available. Please start server first")
	}

	fmt.Println("Starting load...")
	go generateLoad(10 * time.Second)

	time.Sleep(5 * time.Second)
	fmt.Printf("Capturing %s...\n", filename)

	if err := saveProfile(filename); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	time.Sleep(5 * time.Second)
	fmt.Printf("Profile saved to %s/%s\n", profilesDir, filename)

	return nil
}

func generateLoad(duration time.Duration) {
	client := &http.Client{Timeout: 5 * time.Second}
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	start := time.Now()
	requestCount := 0

	for {
		select {
		case <-ticker.C:
			if time.Since(start) >= duration {
				fmt.Printf("Sent %d requests\n", requestCount)
				return
			}

			if requestCount%2 == 0 {
				sendGaugeMetric(client, requestCount)
			} else {
				sendCounterMetric(client, requestCount)
			}
			requestCount++

		case <-time.After(duration):
			fmt.Printf("Sent %d requests\n", requestCount)
			return
		}
	}
}

func sendGaugeMetric(client *http.Client, id int) {
	value := float64(id * 123)
	metric := model.Metric{
		ID:    fmt.Sprintf("gauge_%d", id%10),
		MType: "gauge",
		Value: &value,
	}

	data, _ := json.Marshal(metric)
	resp, err := client.Post(updateURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return
	}
	resp.Body.Close()
}

func sendCounterMetric(client *http.Client, id int) {
	delta := int64(1)
	metric := model.Metric{
		ID:    fmt.Sprintf("counter_%d", id%5),
		MType: "counter",
		Delta: &delta,
	}

	data, _ := json.Marshal(metric)
	resp, err := client.Post(updateURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return
	}
	resp.Body.Close()
}

func isServerReady() bool {
	for range 5 {
		if resp, err := http.Get(serverURL + "/ping"); err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func saveProfile(filename string) error {
	resp, err := http.Get(serverURL + "/debug/pprof/heap")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(fmt.Sprintf("%s/%s", profilesDir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
