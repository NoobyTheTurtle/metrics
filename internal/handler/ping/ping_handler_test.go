package ping

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDBPinger(ctrl)
	mockLogger := NewMockPingLogger(ctrl)

	handler := NewHandler(mockDB, mockLogger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockDB, handler.db)
	assert.Equal(t, mockLogger, handler.logger)
}

func TestHandler_PingHandler(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := NewMockDBPinger(ctrl)
		mockLogger := NewMockPingLogger(ctrl)

		handler := NewHandler(mockDB, mockLogger)

		mockDB.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)

		req := httptest.NewRequest("GET", "/ping", nil)
		w := httptest.NewRecorder()

		pingHandler := handler.PingHandler()
		pingHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ping error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := NewMockDBPinger(ctrl)
		mockLogger := NewMockPingLogger(ctrl)

		handler := NewHandler(mockDB, mockLogger)

		mockDB.EXPECT().Ping(gomock.Any()).Return(assert.AnError).Times(1)
		mockLogger.EXPECT().Error("Database connection failed: %v", assert.AnError).Times(1)

		req := httptest.NewRequest("GET", "/ping", nil)
		w := httptest.NewRecorder()

		pingHandler := handler.PingHandler()
		pingHandler(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_Integration(t *testing.T) {
	t.Run("handler lifecycle", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := NewMockDBPinger(ctrl)
		mockLogger := NewMockPingLogger(ctrl)

		handler := NewHandler(mockDB, mockLogger)

		mockDB.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)

		req := httptest.NewRequest("GET", "/ping", nil)
		w := httptest.NewRecorder()

		pingHandler := handler.PingHandler()
		pingHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("handler with context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := NewMockDBPinger(ctrl)
		mockLogger := NewMockPingLogger(ctrl)

		handler := NewHandler(mockDB, mockLogger)

		mockDB.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)

		ctx := context.Background()
		req := httptest.NewRequest("GET", "/ping", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		pingHandler := handler.PingHandler()
		pingHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
