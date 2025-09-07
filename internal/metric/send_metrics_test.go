package metric

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMetrics_SendMetrics(t *testing.T) {
	tests := []struct {
		name             string
		gauges           map[GaugeMetric]float64
		counters         map[CounterMetric]int64
		serverHandler    http.HandlerFunc
		expectedGauges   map[string]float64
		expectedCounters map[string]int64
		statusCode       int
	}{
		{
			name: "success send metrics",
			gauges: map[GaugeMetric]float64{
				"Alloc":       1.1,
				"HeapObjects": 2.2,
			},
			counters: map[CounterMetric]int64{
				"PollCount": 5,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedGauges: map[string]float64{
				"Alloc":       1.1,
				"HeapObjects": 2.2,
			},
			expectedCounters: map[string]int64{
				"PollCount": 5,
			},
			statusCode: http.StatusOK,
		},
		{
			name: "server error",
			gauges: map[GaugeMetric]float64{
				"Alloc": 1.1,
			},
			counters: map[CounterMetric]int64{
				"PollCount": 5,
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedGauges: map[string]float64{
				"Alloc": 1.1,
			},
			expectedCounters: map[string]int64{
				"PollCount": 5,
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name:     "empty metrics",
			gauges:   map[GaugeMetric]float64{},
			counters: map[CounterMetric]int64{},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedGauges:   map[string]float64{},
			expectedCounters: map[string]int64{},
			statusCode:       http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/updates/", r.URL.Path)

				xRealIP := r.Header.Get("X-Real-IP")
				assert.NotEmpty(t, xRealIP, "X-Real-IP header should be present")

				parsedIP := net.ParseIP(xRealIP)
				assert.NotNil(t, parsedIP, "X-Real-IP should be a valid IP address")

				var body []byte
				var err error

				reader, err := gzip.NewReader(r.Body)
				require.NoError(t, err)

				body, err = io.ReadAll(reader)
				require.NoError(t, err)
				defer r.Body.Close()

				var receivedMetrics model.Metrics
				err = json.Unmarshal(body, &receivedMetrics)
				require.NoError(t, err)

				if len(tt.gauges) == 0 && len(tt.counters) == 0 {
					tt.serverHandler(w, r)
					return
				}

				expectedGauges := make(map[string]float64)
				for k, v := range tt.expectedGauges {
					expectedGauges[k] = v
				}
				expectedCounters := make(map[string]int64)
				for k, v := range tt.expectedCounters {
					expectedCounters[k] = v
				}

				for _, metric := range receivedMetrics {
					if metric.MType == Gauge {
						expectedValue, exists := expectedGauges[metric.ID]
						assert.True(t, exists, "Unexpected gauge metric: %s", metric.ID)
						if exists {
							assert.Equal(t, expectedValue, *metric.Value)
							delete(expectedGauges, metric.ID)
						}
					} else if metric.MType == Counter {
						expectedValue, exists := expectedCounters[metric.ID]
						assert.True(t, exists, "Unexpected counter metric: %s", metric.ID)
						if exists {
							assert.Equal(t, expectedValue, *metric.Delta)
							delete(expectedCounters, metric.ID)
						}
					} else {
						t.Errorf("Unexpected metric type: %s", metric.MType)
					}
				}

				assert.Empty(t, expectedGauges, "Not all expected gauge metrics were received")
				assert.Empty(t, expectedCounters, "Not all expected counter metrics were received")

				tt.serverHandler(w, r)
			}))
			defer server.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockLogger := NewMockMetricsLogger(ctrl)

			mockLogger.EXPECT().Warn("Failed to send metrics batch: %v", gomock.Any()).AnyTimes()

			transport := NewHTTPTransport(server.URL[7:], false, "", nil, mockLogger)
			metrics := NewMetricsWithTransport(mockLogger, transport)
			metrics.Gauges = tt.gauges
			metrics.Counters = tt.counters

			metrics.SendMetrics()
		})
	}
}

func TestCompressJSON(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "compress valid JSON",
			input: []byte(`{"id":"test","type":"gauge","value":123.45}`),
		},
		{
			name:  "compress empty JSON",
			input: []byte(`{}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := compressJSON(tt.input)

			assert.NoError(t, err)
			assert.NotNil(t, compressed)

			reader, err := gzip.NewReader(bytes.NewReader(compressed))
			require.NoError(t, err)

			decompressed, err := io.ReadAll(reader)
			require.NoError(t, err)

			assert.Equal(t, tt.input, decompressed)
		})
	}
}

func TestGetRealIPAddress(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
		expectIP    bool
	}{
		{
			name:        "should return valid IP address",
			expectError: false,
			expectIP:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := getRealIPAddress()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, ip)
			} else {
				assert.NoError(t, err)
				if tt.expectIP {
					assert.NotNil(t, ip)
					assert.NotEqual(t, "127.0.0.1", ip.String())
					assert.NotEqual(t, "::1", ip.String())

					assert.True(t, ip.To4() != nil, "Should be valid IPv4 address")
				} else {
					assert.Nil(t, ip)
				}
			}
		})
	}
}
