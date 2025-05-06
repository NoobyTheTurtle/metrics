package metric

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
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

			if tt.statusCode != http.StatusOK {
				mockLogger.EXPECT().Warn("Failed to send metrics batch: %v", gomock.Any()).Times(1)
			}

			metrics := &Metrics{
				Gauges:    tt.gauges,
				Counters:  tt.counters,
				serverURL: server.URL,
				logger:    mockLogger,
				client:    &http.Client{},
			}

			metrics.SendMetrics()
		})
	}
}

func TestSendMetricsBatch(t *testing.T) {
	tests := []struct {
		name       string
		metric     model.Metric
		statusCode int
	}{
		{
			name: "success send gauge metric",
			metric: func() model.Metric {
				value := 1.1
				return model.Metric{
					ID:    "Alloc",
					MType: Gauge,
					Value: &value,
				}
			}(),
			statusCode: http.StatusOK,
		},
		{
			name: "success send counter metric",
			metric: func() model.Metric {
				delta := int64(5)
				return model.Metric{
					ID:    "PollCount",
					MType: Counter,
					Delta: &delta,
				}
			}(),
			statusCode: http.StatusOK,
		},
		{
			name: "server error",
			metric: func() model.Metric {
				value := 1.1
				return model.Metric{
					ID:    "Alloc",
					MType: Gauge,
					Value: &value,
				}
			}(),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/updates/", r.URL.Path)

				reader, err := gzip.NewReader(r.Body)
				assert.NoError(t, err)

				body, err := io.ReadAll(reader)
				assert.NoError(t, err)
				defer r.Body.Close()

				var receivedMetrics model.Metrics
				err = receivedMetrics.UnmarshalJSON(body)
				assert.NoError(t, err)
				assert.Len(t, receivedMetrics, 1, "Expected exactly one metric")

				receivedMetric := receivedMetrics[0]
				assert.Equal(t, tt.metric.ID, receivedMetric.ID)
				assert.Equal(t, tt.metric.MType, receivedMetric.MType)

				if tt.metric.Delta != nil {
					assert.NotNil(t, receivedMetric.Delta)
					assert.Equal(t, *tt.metric.Delta, *receivedMetric.Delta)
				}

				if tt.metric.Value != nil {
					assert.NotNil(t, receivedMetric.Value)
					assert.Equal(t, *tt.metric.Value, *receivedMetric.Value)
				}

				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockLogger := NewMockMetricsLogger(ctrl)

			metrics := &Metrics{
				Gauges:    make(map[GaugeMetric]float64),
				Counters:  make(map[CounterMetric]int64),
				serverURL: server.URL,
				logger:    mockLogger,
				client:    &http.Client{},
			}

			err := metrics.SendMetricsBatch(model.Metrics{tt.metric})

			if tt.statusCode == http.StatusOK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
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
