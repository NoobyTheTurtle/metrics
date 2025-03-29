package metrics

import (
	"fmt"
	"log"
	"net/http"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

func (m *Metrics) SendMetrics() {
	for name, value := range m.Gauges {
		url := fmt.Sprintf("%s/update/%s/%s/%v", m.serverAddress, Gauge, name, value)
		sendMetric(url)
	}

	for name, value := range m.Counters {
		url := fmt.Sprintf("%s/update/%s/%s/%v", m.serverAddress, Counter, name, value)
		sendMetric(url)
	}
}

func sendMetric(url string) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	req.Header.Add("Content-Type", "text/plain; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Server returned status code: %d", resp.StatusCode)
	}
}
