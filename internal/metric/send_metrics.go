package metric

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
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

	ctx := context.Background()
	err := m.transport.SendMetrics(ctx, metrics)
	if err != nil {
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
		return nil, fmt.Errorf("metric.compressJSON: gzip write error: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("metric.compressJSON: gzip close error: %w", err)
	}

	return buf.Bytes(), nil
}

func readResponseBody(resp *http.Response) (string, error) {
	var reader io.ReadCloser
	var err error

	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("metric.readResponseBody: error creating gzip reader: %w", err)
		}
		defer reader.Close()
	} else {
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("metric.readResponseBody: error reading response body: %w", err)
	}

	return string(body), nil
}

func getRealIPAddress() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var l []net.IP
	var p []net.IP

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				l = append(l, ipnet.IP)
				continue
			}

			if ipnet.IP.IsPrivate() && ipnet.IP.To4() != nil {
				p = append(p, ipnet.IP)
				continue
			}

			if ipnet.IP.To4() != nil {
				return ipnet.IP, nil
			}
		}
	}

	if len(p) > 0 {
		return p[0], nil
	}

	if len(l) > 0 {
		return l[0], nil
	}

	return nil, nil
}
