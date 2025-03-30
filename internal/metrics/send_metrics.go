package metrics

import (
	"fmt"
	"net/http"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

func (m *Metrics) SendMetrics() {
	for name, value := range m.Gauges {
		url := fmt.Sprintf("%s/update/%s/%s/%v", m.serverUrl, Gauge, name, value)
		sendMetric(url, m.logger)
	}

	for name, value := range m.Counters {
		url := fmt.Sprintf("%s/update/%s/%s/%v", m.serverUrl, Counter, name, value)
		sendMetric(url, m.logger)
	}
}

func sendMetric(url string, logger Logger) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		logger.Error("Error creating request: %v", err)
		return
	}

	req.Header.Add("Content-Type", "text/plain; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("Server returned status code: %d", resp.StatusCode)
	}
}
