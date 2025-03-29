package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ ServerStorage = (*mockStorage)(nil)

type mockStorage struct {
	gauges            map[string]float64
	counters          map[string]int64
	shouldFailGauge   bool
	shouldFailCounter bool
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *mockStorage) UpdateGauge(name string, value float64) error {
	if m.shouldFailGauge {
		return errors.New("gauge update error")
	}
	m.gauges[name] = value
	return nil
}

func (m *mockStorage) UpdateCounter(name string, value int64) error {
	if m.shouldFailCounter {
		return errors.New("counter update error")
	}
	m.counters[name] += value
	return nil
}

func (m *mockStorage) GetGauge(name string) (float64, bool) {
	value, ok := m.gauges[name]
	return value, ok
}

func (m *mockStorage) GetCounter(name string) (int64, bool) {
	value, ok := m.counters[name]
	return value, ok
}

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
			storage := newMockStorage()
			storage.shouldFailGauge = tt.shouldFailGauge
			storage.shouldFailCounter = tt.shouldFailCounter

			h := &handler{
				storage: storage,
			}

			req, err := http.NewRequest(tt.method, tt.url, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			h.updateHandler()(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code, "Expected status code %d, got %d", tt.expectedStatusCode, rr.Code)
			assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"), "Expected Content-Type text/plain; charset=utf-8, got %s", rr.Header().Get("Content-Type"))
		})
	}
}
