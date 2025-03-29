package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func Test_handler_updateHandler(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		url                string
		shouldFailGauge    bool
		shouldFailCounter  bool
		expectedStatusCode int
	}{
		{
			name:               "successful gauge update",
			method:             http.MethodPost,
			url:                "/update/gauge/HeapObjects/7770",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "successful counter update",
			method:             http.MethodPost,
			url:                "/update/counter/PollCount/30",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "wrong method",
			method:             http.MethodGet,
			url:                "/update/counter/PollCount/30",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "invalid url format",
			method:             http.MethodPost,
			url:                "/update/gauge/wrong-format",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "empty metric name",
			method:             http.MethodPost,
			url:                "/update/gauge//30",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "invalid gauge value",
			method:             http.MethodPost,
			url:                "/update/gauge/HeapObjects/not-a-number",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "invalid counter value",
			method:             http.MethodPost,
			url:                "/update/counter/PollCount/not-a-number",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "unknown metric type",
			method:             http.MethodPost,
			url:                "/update/unknown/PollCount/30",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "gauge update error",
			method:             http.MethodPost,
			url:                "/update/gauge/HeapObjects/7770",
			shouldFailGauge:    true,
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:               "counter update error",
			method:             http.MethodPost,
			url:                "/update/counter/PollCount/30",
			shouldFailCounter:  true,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := storage.NewMockStorage()
			mockStorage.SetShouldFailGauge(tt.shouldFailGauge)
			mockStorage.SetShouldFailCounter(tt.shouldFailCounter)

			h := &handler{
				storage: mockStorage,
			}

			r := chi.NewRouter()
			r.Post("/update/{metricType}/{metricName}/{metricValue}", h.updateHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testRequest(t, ts, tt.method, tt.url)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected status code %d, got %d", tt.expectedStatusCode, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"),
					"Expected Content-Type text/plain; charset=utf-8, got %s", resp.Header.Get("Content-Type"))
			}
		})
	}
}
