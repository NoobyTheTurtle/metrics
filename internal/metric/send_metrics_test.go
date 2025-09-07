package metric

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NoobyTheTurtle/metrics/internal/hash"
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

				xRealIP := r.Header.Get("X-Real-IP")
				assert.NotEmpty(t, xRealIP, "X-Real-IP header should be present")

				parsedIP := net.ParseIP(xRealIP)
				assert.NotNil(t, parsedIP, "X-Real-IP should be a valid IP address")

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

func TestSendMetricsBatch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/updates/", r.URL.Path)

		xRealIP := r.Header.Get("X-Real-IP")
		assert.NotEmpty(t, xRealIP, "X-Real-IP header should be present")

		parsedIP := net.ParseIP(xRealIP)
		assert.NotNil(t, parsedIP, "X-Real-IP should be a valid IP address")

		reader, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer reader.Close()

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		var receivedMetrics model.Metrics
		err = json.Unmarshal(body, &receivedMetrics)
		require.NoError(t, err)
		assert.Len(t, receivedMetrics, 1)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := NewMockMetricsLogger(ctrl)

	metrics := &Metrics{
		serverURL: server.URL,
		logger:    mockLogger,
		client:    &http.Client{},
	}

	testMetrics := model.Metrics{
		{
			ID:    "test",
			MType: Gauge,
			Value: func() *float64 { v := 1.5; return &v }(),
		},
	}

	err := metrics.SendMetricsBatch(testMetrics)

	assert.NoError(t, err)
}

func TestSendMetricsBatch_EmptyBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reader, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer reader.Close()

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		var receivedMetrics model.Metrics
		err = json.Unmarshal(body, &receivedMetrics)
		require.NoError(t, err)
		assert.Len(t, receivedMetrics, 0)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := NewMockMetricsLogger(ctrl)

	metrics := &Metrics{
		serverURL: server.URL,
		logger:    mockLogger,
		client:    &http.Client{},
	}

	err := metrics.SendMetricsBatch(model.Metrics{})

	assert.NoError(t, err)
}

func TestSendMetricsBatch_NetworkError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := NewMockMetricsLogger(ctrl)

	metrics := &Metrics{
		serverURL: "http://invalid-url-that-does-not-exist.local",
		logger:    mockLogger,
		client:    &http.Client{Timeout: 1 * time.Millisecond},
	}

	testMetrics := model.Metrics{
		{
			ID:    "test",
			MType: Gauge,
			Value: func() *float64 { v := 1.5; return &v }(),
		},
	}

	err := metrics.SendMetricsBatch(testMetrics)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error sending request")
}

func TestSendMetricsBatch_InvalidResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		contentType    string
		expectGzipBody bool
	}{
		{
			name:         "bad request",
			statusCode:   http.StatusBadRequest,
			responseBody: "Bad Request",
			contentType:  "text/plain",
		},
		{
			name:         "internal server error",
			statusCode:   http.StatusInternalServerError,
			responseBody: "Internal Server Error",
			contentType:  "text/plain",
		},
		{
			name:           "service unavailable with gzip",
			statusCode:     http.StatusServiceUnavailable,
			responseBody:   "Service Unavailable",
			contentType:    "text/plain",
			expectGzipBody: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)

				if tt.expectGzipBody {
					w.Header().Set("Content-Encoding", "gzip")

					var buf bytes.Buffer
					gzWriter := gzip.NewWriter(&buf)
					gzWriter.Write([]byte(tt.responseBody))
					gzWriter.Close()

					w.WriteHeader(tt.statusCode)
					w.Write(buf.Bytes())
				} else {
					w.WriteHeader(tt.statusCode)
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockLogger := NewMockMetricsLogger(ctrl)

			metrics := &Metrics{
				serverURL: server.URL,
				logger:    mockLogger,
				client:    &http.Client{},
			}

			testMetrics := model.Metrics{
				{
					ID:    "test",
					MType: Gauge,
					Value: func() *float64 { v := 1.5; return &v }(),
				},
			}

			err := metrics.SendMetricsBatch(testMetrics)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), fmt.Sprintf("server returned status code %d", tt.statusCode))
			assert.Contains(t, err.Error(), tt.responseBody)
		})
	}
}

func TestSendMetricsBatch_WithGzip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))

		reader, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer reader.Close()

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		var receivedMetrics model.Metrics
		err = json.Unmarshal(body, &receivedMetrics)
		require.NoError(t, err)
		assert.Len(t, receivedMetrics, 2)

		w.Header().Set("Content-Encoding", "gzip")
		var buf bytes.Buffer
		gzWriter := gzip.NewWriter(&buf)
		gzWriter.Write([]byte("OK"))
		gzWriter.Close()

		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}))
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := NewMockMetricsLogger(ctrl)

	metrics := &Metrics{
		serverURL: server.URL,
		logger:    mockLogger,
		client:    &http.Client{},
	}

	testMetrics := model.Metrics{
		{
			ID:    "gauge_test",
			MType: Gauge,
			Value: func() *float64 { v := 1.5; return &v }(),
		},
		{
			ID:    "counter_test",
			MType: Counter,
			Delta: func() *int64 { v := int64(10); return &v }(),
		},
	}

	err := metrics.SendMetricsBatch(testMetrics)

	assert.NoError(t, err)
}

func TestSendMetricsBatch_WithHash(t *testing.T) {
	const testKey = "test-secret-key"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hashHeader := r.Header.Get("HashSHA256")
		assert.NotEmpty(t, hashHeader)

		reader, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer reader.Close()

		body, err := io.ReadAll(reader)
		require.NoError(t, err)

		expectedHash, err := hash.CalculateSHA256(body, testKey)
		require.NoError(t, err)
		assert.Equal(t, expectedHash, hashHeader)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := NewMockMetricsLogger(ctrl)

	metrics := &Metrics{
		serverURL: server.URL,
		logger:    mockLogger,
		client:    &http.Client{},
		key:       testKey,
	}

	testMetrics := model.Metrics{
		{
			ID:    "test",
			MType: Gauge,
			Value: func() *float64 { v := 1.5; return &v }(),
		},
	}

	err := metrics.SendMetricsBatch(testMetrics)

	assert.NoError(t, err)
}

func TestSendMetricsBatch_WithEncryption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encryptedBody, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		assert.NotEmpty(t, encryptedBody)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockMetricsLogger(ctrl)
	mockEncrypter := NewMockEncrypter(ctrl)

	mockEncrypter.EXPECT().
		Encrypt(gomock.Any()).
		Return([]byte("encrypted_data"), nil).
		Times(1)

	metrics := &Metrics{
		serverURL: server.URL,
		logger:    mockLogger,
		client:    &http.Client{},
		encrypter: mockEncrypter,
	}

	testMetrics := model.Metrics{
		{
			ID:    "test",
			MType: Gauge,
			Value: func() *float64 { v := 1.5; return &v }(),
		},
	}

	err := metrics.SendMetricsBatch(testMetrics)

	assert.NoError(t, err)
}

func TestSendMetricsBatch_EncryptionFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockMetricsLogger(ctrl)
	mockEncrypter := NewMockEncrypter(ctrl)

	encryptError := assert.AnError
	mockEncrypter.EXPECT().
		Encrypt(gomock.Any()).
		Return(nil, encryptError).
		Times(1)

	mockLogger.EXPECT().
		Warn("Failed to encrypt data: %v", encryptError).
		Times(1)

	metrics := &Metrics{
		serverURL: server.URL,
		logger:    mockLogger,
		client:    &http.Client{},
		encrypter: mockEncrypter,
	}

	testMetrics := model.Metrics{
		{
			ID:    "test",
			MType: Gauge,
			Value: func() *float64 { v := 1.5; return &v }(),
		},
	}

	err := metrics.SendMetricsBatch(testMetrics)

	assert.NoError(t, err)
}

func TestSendMetricsBatch_HashCalculationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hashHeader := r.Header.Get("HashSHA256")
		assert.Empty(t, hashHeader)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := NewMockMetricsLogger(ctrl)

	metrics := &Metrics{
		serverURL: server.URL,
		logger:    mockLogger,
		client:    &http.Client{},
		key:       "",
	}

	testMetrics := model.Metrics{
		{
			ID:    "test",
			MType: Gauge,
			Value: func() *float64 { v := 1.5; return &v }(),
		},
	}

	err := metrics.SendMetricsBatch(testMetrics)

	assert.NoError(t, err)
}

func TestSendMetricsBatch_ResponseReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid gzip data"))
	}))
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := NewMockMetricsLogger(ctrl)

	metrics := &Metrics{
		serverURL: server.URL,
		logger:    mockLogger,
		client:    &http.Client{},
	}

	testMetrics := model.Metrics{
		{
			ID:    "test",
			MType: Gauge,
			Value: func() *float64 { v := 1.5; return &v }(),
		},
	}

	err := metrics.SendMetricsBatch(testMetrics)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not read body")
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
