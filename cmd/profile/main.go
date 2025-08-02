package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/profile/main.go [base|result]")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "base":
		captureProfile("base.pprof")
	case "result":
		captureProfile("result.pprof")
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func captureProfile(filename string) {
	fmt.Printf("Capturing %s...\n", filename)

	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		log.Fatalf("Failed to create profiles directory: %v", err)
	}

	if !isServerReady() {
		log.Fatal("Server not available. Please start server first.")
	}

	fmt.Println("Starting load...")
	go generateLoad(10 * time.Second)

	time.Sleep(5 * time.Second)
	fmt.Printf("Capturing %s...\n", filename)

	if err := saveProfile(filename); err != nil {
		log.Fatalf("Failed to save profile: %v", err)
	}

	time.Sleep(5 * time.Second)
	fmt.Printf("Profile saved to %s/%s\n", profilesDir, filename)
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
