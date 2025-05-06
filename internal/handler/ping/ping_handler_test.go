package ping

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_PingHandler(t *testing.T) {
	tests := []struct {
		name               string
		setupMocks         func(*gomock.Controller) (*MockDBPinger, *MockPingLogger)
		expectedStatusCode int
	}{
		{
			name: "successful ping",
			setupMocks: func(ctrl *gomock.Controller) (*MockDBPinger, *MockPingLogger) {
				mockDB := NewMockDBPinger(ctrl)
				mockLogger := NewMockPingLogger(ctrl)

				mockDB.EXPECT().Ping(gomock.Any()).Return(nil)

				return mockDB, mockLogger
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "database connection error",
			setupMocks: func(ctrl *gomock.Controller) (*MockDBPinger, *MockPingLogger) {
				mockDB := NewMockDBPinger(ctrl)
				mockLogger := NewMockPingLogger(ctrl)

				dbErr := errors.New("database connection error")
				mockDB.EXPECT().Ping(gomock.Any()).Return(dbErr)
				mockLogger.EXPECT().Error("Database connection failed: %v", dbErr)

				return mockDB, mockLogger
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB, mockLogger := tt.setupMocks(ctrl)

			h := &Handler{
				db:     mockDB,
				logger: mockLogger,
			}

			r := chi.NewRouter()
			r.Get("/ping", h.PingHandler())

			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, _ := testutil.TestRequest(t, ts, http.MethodGet, "/ping", "")
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
		})
	}
}
