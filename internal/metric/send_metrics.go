package metric

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
		url := fmt.Sprintf("%s/update/%s/%s/%v", m.serverURL, Gauge, name, value)
		m.sendMetric(url)
	}

	for name, value := range m.Counters {
		url := fmt.Sprintf("%s/update/%s/%s/%v", m.serverURL, Counter, name, value)
		m.sendMetric(url)
	}
}

func (m *Metrics) sendMetric(url string) {
	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		m.logger.Error("Error creating request: %v", err)
		return
	}

	req.Header.Add("Content-Type", "text/plain; charset=utf-8")

	resp, err := m.client.Do(req)
	if err != nil {
		m.logger.Error("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Warn("Server returned status code: %d", resp.StatusCode)
	}
}
