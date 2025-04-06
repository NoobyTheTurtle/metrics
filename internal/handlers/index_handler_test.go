package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_mapGauges(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]float64
		expected []metricData
	}{
		{
			name:     "empty gauges",
			input:    map[string]float64{},
			expected: []metricData{},
		},
		{
			name: "multiple gauges sorted",
			input: map[string]float64{
				"Z": 1.1,
				"A": 2.2,
				"M": 3.3,
			},
			expected: []metricData{
				{Name: "A", Value: 2.2},
				{Name: "M", Value: 3.3},
				{Name: "Z", Value: 1.1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapGauges(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_mapCounters(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int64
		expected []metricData
	}{
		{
			name:     "empty counters",
			input:    map[string]int64{},
			expected: []metricData{},
		},
		{
			name: "multiple counters sorted",
			input: map[string]int64{
				"Z": 1,
				"A": 2,
				"M": 3,
			},
			expected: []metricData{
				{Name: "A", Value: int64(2)},
				{Name: "M", Value: int64(3)},
				{Name: "Z", Value: int64(1)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapCounters(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_handler_indexHandler(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		url                string
		setupStorage       func(*gomock.Controller) *mocks.MockServerStorage
		expectedStatusCode int
		expectedContains   []string
	}{
		{
			name:   "successful metrics page retrieval",
			method: http.MethodGet,
			url:    "/",
			setupStorage: func(ctrl *gomock.Controller) *mocks.MockServerStorage {
				mockStorage := mocks.NewMockServerStorage(ctrl)

				gauges := map[string]float64{
					"Alloc":       15.5,
					"BuckHashSys": 30.25,
				}
				counters := map[string]int64{
					"PollCount": 30,
				}

				mockStorage.EXPECT().GetAllGauges().Return(gauges)
				mockStorage.EXPECT().GetAllCounters().Return(counters)

				return mockStorage
			},
			expectedStatusCode: http.StatusOK,
			expectedContains: []string{
				"<h1>Metrics</h1>",
				"<h2>Gauge</h2>",
				"<h2>Counter</h2>",
				"Alloc",
				"15.5",
				"BuckHashSys",
				"30.25",
				"PollCount",
				"30",
			},
		},
		{
			name:   "empty metrics",
			method: http.MethodGet,
			url:    "/",
			setupStorage: func(ctrl *gomock.Controller) *mocks.MockServerStorage {
				mockStorage := mocks.NewMockServerStorage(ctrl)

				mockStorage.EXPECT().GetAllGauges().Return(map[string]float64{})
				mockStorage.EXPECT().GetAllCounters().Return(map[string]int64{})

				return mockStorage
			},
			expectedStatusCode: http.StatusOK,
			expectedContains: []string{
				"<h1>Metrics</h1>",
				"<h2>Gauge</h2>",
				"<h2>Counter</h2>",
			},
		},
		{
			name:   "wrong method",
			method: http.MethodPost,
			url:    "/",
			setupStorage: func(ctrl *gomock.Controller) *mocks.MockServerStorage {
				return mocks.NewMockServerStorage(ctrl)
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedContains:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := tt.setupStorage(ctrl)

			h := &handler{
				storage: storage,
			}

			r := chi.NewRouter()
			r.Get("/", h.indexHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, body := testRequest(t, ts, tt.method, tt.url)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Expected status code %d, got %d", tt.expectedStatusCode, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				assert.Equal(t, "text/html", resp.Header.Get("Content-Type"),
					"Expected Content-Type text/html, got %s", resp.Header.Get("Content-Type"))

				for _, expectedText := range tt.expectedContains {
					assert.True(t, strings.Contains(body, expectedText),
						"Expected response to contain '%s', but it doesn't. Body: %s", expectedText, body)
				}
			}
		})
	}
}
