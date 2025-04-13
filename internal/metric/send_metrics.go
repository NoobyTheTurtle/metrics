package metric

import (
	"bytes"
	"compress/gzip"
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

func compressJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	_, err := gzWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("gzip write error: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("gzip close error: %w", err)
	}

	return buf.Bytes(), nil
}

func (m *Metrics) sendJSONMetric(metric model.Metrics) {
	jsonData, err := metric.MarshalJSON()
	if err != nil {
		m.logger.Error("Error marshaling metric: %v", err)
		return
	}

	compressedData, err := compressJSON(jsonData)
	if err != nil {
		m.logger.Error("Error compressing data: %v", err)
		return
	}

	url := fmt.Sprintf("%s/update/", m.serverURL)
	var req *http.Request

	req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(compressedData))
	if err != nil {
		m.logger.Error("Error creating request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

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
