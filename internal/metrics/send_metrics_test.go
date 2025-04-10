package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMetrics_SendMetrics(t *testing.T) {
	tests := []struct {
		name                string
		gauges              map[GaugeMetric]float64
		counters            map[CounterMetric]int64
		serverHandler       http.HandlerFunc
		expectedGaugeURLs   map[string]bool
		expectedCounterURLs map[string]bool
		statusCode          int
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
			expectedGaugeURLs: map[string]bool{
				"/update/gauge/Alloc/1.1":       true,
				"/update/gauge/HeapObjects/2.2": true,
			},
			expectedCounterURLs: map[string]bool{
				"/update/counter/PollCount/5": true,
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
			expectedGaugeURLs: map[string]bool{
				"/update/gauge/Alloc/1.1": true,
			},
			expectedCounterURLs: map[string]bool{
				"/update/counter/PollCount/5": true,
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
			expectedGaugeURLs:   map[string]bool{},
			expectedCounterURLs: map[string]bool{},
			statusCode:          http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "text/plain; charset=utf-8", r.Header.Get("Content-Type"))
				assert.Equal(t, http.MethodPost, r.Method)

				path := r.URL.Path
				if _, ok := tt.expectedGaugeURLs[path]; ok {
					delete(tt.expectedGaugeURLs, path)
				} else if _, ok := tt.expectedCounterURLs[path]; ok {
					delete(tt.expectedCounterURLs, path)
				} else {
					t.Errorf("Unexpected request URL: %s", path)
				}

				tt.serverHandler(w, r)
			}))
			defer server.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockLogger := NewMockmetricsLogger(ctrl)

			if tt.statusCode != http.StatusOK {
				mockLogger.EXPECT().Warn("Server returned status code: %d", tt.statusCode).Times(len(tt.gauges) + len(tt.counters))
			}

			metrics := &Metrics{
				Gauges:    tt.gauges,
				Counters:  tt.counters,
				serverURL: server.URL,
				logger:    mockLogger,
				client:    &http.Client{},
			}

			metrics.SendMetrics()

			assert.Empty(t, tt.expectedGaugeURLs)
			assert.Empty(t, tt.expectedCounterURLs)
		})
	}
}

func TestSendMetric(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		statusCode    int
	}{
		{
			name: "success send metric",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			statusCode: http.StatusOK,
		},
		{
			name: "server error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "text/plain; charset=utf-8", r.Header.Get("Content-Type"))
				assert.Equal(t, http.MethodPost, r.Method)
				tt.serverHandler(w, r)
			}))
			defer server.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockLogger := NewMockmetricsLogger(ctrl)

			if tt.statusCode != http.StatusOK {
				mockLogger.EXPECT().Warn("Server returned status code: %d", tt.statusCode).Times(1)
			}

			metrics := &Metrics{
				Gauges:    make(map[GaugeMetric]float64),
				Counters:  make(map[CounterMetric]int64),
				serverURL: server.URL,
				logger:    mockLogger,
				client:    &http.Client{},
			}

			metrics.sendMetric(server.URL + "/update/gauge/testMetric/1.1")
		})
	}
}
