package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func Test_handler_valueHandler(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		url                string
		setupStorage       func(*storage.MockStorage)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:   "successful gauge retrieval",
			method: http.MethodGet,
			url:    "/value/gauge/HeapObjects",
			setupStorage: func(m *storage.MockStorage) {
				m.UpdateGauge("HeapObjects", 1.2)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "1.2",
		},
		{
			name:   "successful counter retrieval",
			method: http.MethodGet,
			url:    "/value/counter/PollCount",
			setupStorage: func(m *storage.MockStorage) {
				m.UpdateCounter("PollCount", 30)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "30",
		},
		{
			name:               "gauge not found",
			method:             http.MethodGet,
			url:                "/value/gauge/NonExistentGauge",
			setupStorage:       func(m *storage.MockStorage) {},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Gauge not found\n",
		},
		{
			name:               "counter not found",
			method:             http.MethodGet,
			url:                "/value/counter/NonExistentCounter",
			setupStorage:       func(m *storage.MockStorage) {},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Counter not found\n",
		},
		{
			name:               "wrong method",
			method:             http.MethodPost,
			url:                "/value/gauge/HeapObjects",
			setupStorage:       func(m *storage.MockStorage) {},
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "unknown metric type",
			method:             http.MethodGet,
			url:                "/value/unknown/Metric",
			setupStorage:       func(m *storage.MockStorage) {},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Unknown metric type\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := storage.NewMockStorage()
			tt.setupStorage(storage)

			h := &handler{
				storage: storage,
			}

			r := chi.NewRouter()
			r.Get("/value/{metricType}/{metricName}", h.valueHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, body := testRequest(t, ts, tt.method, tt.url)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected status code %d, got %d", tt.expectedStatusCode, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"),
					"Expected Content-Type text/plain; charset=utf-8, got %s", resp.Header.Get("Content-Type"))
			}

			assert.Equal(t, tt.expectedBody, body, "Expected response body '%s', got '%s'", tt.expectedBody, body)
		})
	}
}
