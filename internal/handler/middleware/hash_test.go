package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/NoobyTheTurtle/metrics/internal/hash"
)

func TestHashValidator(t *testing.T) {
	secretKey := "testSecret"
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	tests := []struct {
		name                 string
		key                  string
		requestBody          string
		hashHeader           string
		provideHeader        bool
		calculateDynamicHash bool
		expectedStatus       int
		expectNextCall       bool
	}{
		{
			name:                 "empty key - should pass",
			key:                  "",
			requestBody:          "test body",
			provideHeader:        true,
			hashHeader:           "any",
			calculateDynamicHash: false,
			expectedStatus:       http.StatusOK,
			expectNextCall:       true,
		},
		{
			name:                 "no hash header - should fail",
			key:                  secretKey,
			requestBody:          "test body",
			provideHeader:        false,
			calculateDynamicHash: false,
			expectedStatus:       http.StatusBadRequest,
			expectNextCall:       false,
		},
		{
			name:                 "mismatched hash - should fail",
			key:                  secretKey,
			requestBody:          "test body",
			provideHeader:        true,
			hashHeader:           "wronghash",
			calculateDynamicHash: false,
			expectedStatus:       http.StatusBadRequest,
			expectNextCall:       false,
		},
		{
			name:                 "valid hash - should pass",
			key:                  secretKey,
			requestBody:          "test body",
			provideHeader:        true,
			calculateDynamicHash: true,
			expectedStatus:       http.StatusOK,
			expectNextCall:       true,
		},
		{
			name:                 "empty body with valid hash - should pass",
			key:                  secretKey,
			requestBody:          "",
			provideHeader:        true,
			calculateDynamicHash: true,
			expectedStatus:       http.StatusOK,
			expectNextCall:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			logger := NewMockMiddlewareLogger(ctrl)
			var nextCalled bool
			handlerToTest := HashValidator(tt.key, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				testHandler.ServeHTTP(w, r)
			}))

			reqBody := bytes.NewBufferString(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", reqBody)

			if tt.name == "no hash header - should fail" {
				logger.EXPECT().Info(
					"Request without HashSHA256 header from %s for %s",
					gomock.Any(),
					gomock.Any(),
				).Times(1)
			}
			if tt.name == "mismatched hash - should fail" {
				logger.EXPECT().Info(
					"Hash mismatch for request from %s for %s. Incoming: %s, Calculated: %s",
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Times(1)
			}

			if tt.provideHeader {
				headerValue := tt.hashHeader
				if tt.calculateDynamicHash {
					calculatedHash, err := hash.CalculateSHA256([]byte(tt.requestBody), tt.key)
					require.NoError(t, err)
					headerValue = calculatedHash
				}
				req.Header.Set("HashSHA256", headerValue)
			}

			rr := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectNextCall, nextCalled)

			if tt.expectNextCall && tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "OK", rr.Body.String())
			}
		})
	}
}

func TestHashAppender(t *testing.T) {
	defaultTestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	tests := []struct {
		name             string
		key              string
		handler          http.HandlerFunc
		expectedHash     bool
		expectedResponse string
		expectedStatus   int
	}{
		{
			name:             "empty key - no hash added",
			key:              "",
			handler:          defaultTestHandler,
			expectedHash:     false,
			expectedResponse: `{"status":"ok"}`,
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "with key - hash added",
			key:              "testSecretAppender",
			handler:          defaultTestHandler,
			expectedHash:     true,
			expectedResponse: `{"status":"ok"}`,
			expectedStatus:   http.StatusOK,
		},
		{
			name: "handler writes nothing",
			key:  "testSecretAppender",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
			expectedHash:     true,
			expectedResponse: "",
			expectedStatus:   http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			logger := NewMockMiddlewareLogger(ctrl)

			handlerToTest := HashAppender(tt.key, logger)(tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			responseBody, err := io.ReadAll(rr.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResponse, string(responseBody))

			if tt.expectedHash {
				actualHash := rr.Header().Get("HashSHA256")
				require.NotEmpty(t, actualHash, "HashSHA256 header should be present")

				expectedCalculatedHash, err := hash.CalculateSHA256(responseBody, tt.key)
				require.NoError(t, err)
				assert.Equal(t, expectedCalculatedHash, actualHash)
			} else {
				assert.Empty(t, rr.Header().Get("HashSHA256"), "HashSHA256 header should not be present")
			}
		})
	}
}
