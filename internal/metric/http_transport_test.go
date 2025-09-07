package metric

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHTTPTransport_SendMetrics(t *testing.T) {
	tests := []struct {
		name       string
		metrics    model.Metrics
		statusCode int
		expectErr  bool
	}{
		{
			name: "successful send",
			metrics: model.Metrics{
				{
					ID:    "test_gauge",
					MType: "gauge",
					Value: func() *float64 { v := 123.45; return &v }(),
				},
				{
					ID:    "test_counter",
					MType: "counter",
					Delta: func() *int64 { v := int64(10); return &v }(),
				},
			},
			statusCode: http.StatusOK,
			expectErr:  false,
		},
		{
			name: "server error",
			metrics: model.Metrics{
				{
					ID:    "test_gauge",
					MType: "gauge",
					Value: func() *float64 { v := 123.45; return &v }(),
				},
			},
			statusCode: http.StatusInternalServerError,
			expectErr:  true,
		},
		{
			name:       "empty metrics",
			metrics:    model.Metrics{},
			statusCode: http.StatusOK,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
				assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
				assert.Equal(t, "/updates/", r.URL.Path)
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockLogger := NewMockMetricsLogger(ctrl)

			if tt.expectErr {
				mockLogger.EXPECT().Warn("Failed to send metrics batch: %v", gomock.Any()).Times(1)
			}

			transport := NewHTTPTransport(server.URL[7:], false, "", nil, mockLogger)
			ctx := context.Background()

			err := transport.SendMetrics(ctx, tt.metrics)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHTTPTransport_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := NewMockMetricsLogger(ctrl)

	transport := NewHTTPTransport("localhost:8080", false, "", nil, mockLogger)
	err := transport.Close()
	assert.NoError(t, err)
}
