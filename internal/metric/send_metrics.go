package metric

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/NoobyTheTurtle/metrics/internal/model"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

func (m *Metrics) SendMetrics() {
	metrics := m.prepareMetricsBatch()
	if len(metrics) == 0 {
		return
	}

	if err := m.SendMetricsBatch(metrics); err != nil {
		m.logger.Warn("Failed to send metrics batch: %v", err)
	}
}

func (m *Metrics) prepareMetricsBatch() model.Metrics {
	metrics := make(model.Metrics, 0, len(m.Gauges)+len(m.Counters))

	for name, value := range m.Gauges {
		valueCopy := value
		metrics = append(metrics, model.Metric{
			ID:    string(name),
			MType: Gauge,
			Value: &valueCopy,
		})
	}

	for name, value := range m.Counters {
		valueCopy := value
		metrics = append(metrics, model.Metric{
			ID:    string(name),
			MType: Counter,
			Delta: &valueCopy,
		})
	}

	return metrics
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

func readResponseBody(resp *http.Response) (string, error) {
	var reader io.ReadCloser
	var err error

	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error creating gzip reader: %w", err)
		}
		defer reader.Close()
	} else {
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return string(body), nil
}

func (m *Metrics) SendMetricsBatch(metrics model.Metrics) error {
	jsonData, err := metrics.MarshalJSON()
	if err != nil {
		return fmt.Errorf("error marshaling metrics batch: %w", err)
	}

	compressedData, err := compressJSON(jsonData)
	if err != nil {
		return fmt.Errorf("error compressing data: %w", err)
	}

	url := fmt.Sprintf("%s/updates/", m.serverURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(compressedData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, readErr := readResponseBody(resp)
		if readErr != nil {
			return fmt.Errorf("server returned status code: %d, could not read body: %v", resp.StatusCode, readErr)
		}
		return fmt.Errorf("server returned status code: %d, body: %s", resp.StatusCode, bodyText)
	}

	return nil
}
