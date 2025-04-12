package metric

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/model"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

func (m *Metrics) SendMetrics() {
	for name, value := range m.Gauges {
		metricJSON := model.Metrics{
			ID:    string(name),
			MType: Gauge,
			Value: &value,
		}
		m.sendJSONMetric(metricJSON)
	}

	for name, value := range m.Counters {
		metricJSON := model.Metrics{
			ID:    string(name),
			MType: Counter,
			Delta: &value,
		}
		m.sendJSONMetric(metricJSON)
	}
}

func (m *Metrics) sendJSONMetric(metric model.Metrics) {
	jsonData, err := metric.MarshalJSON()
	if err != nil {
		m.logger.Error("Error marshaling metric: %v", err)
		return
	}

	url := fmt.Sprintf("%s/update/", m.serverURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		m.logger.Error("Error creating request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

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
