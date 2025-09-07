package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		provideHeader  bool
		expectedStatus int
		expectNextCall bool
		expectLogError bool
		expectLogInfo  bool
		logMessage     string
	}{
		{
			name:           "empty trusted subnet - should pass without header",
			trustedSubnet:  "",
			xRealIP:        "",
			provideHeader:  false,
			expectedStatus: http.StatusOK,
			expectNextCall: true,
			expectLogError: false,
			expectLogInfo:  false,
		},
		{
			name:           "empty trusted subnet - should pass with header",
			trustedSubnet:  "",
			xRealIP:        "192.168.1.100",
			provideHeader:  true,
			expectedStatus: http.StatusOK,
			expectNextCall: true,
			expectLogError: false,
			expectLogInfo:  false,
		},
		{
			name:           "invalid CIDR - should return 500",
			trustedSubnet:  "invalid-cidr",
			xRealIP:        "192.168.1.100",
			provideHeader:  true,
			expectedStatus: http.StatusInternalServerError,
			expectNextCall: false,
			expectLogError: true,
			expectLogInfo:  false,
			logMessage:     "Invalid CIDR format for trusted subnet",
		},
		{
			name:           "missing X-Real-IP header - should return 403",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "",
			provideHeader:  false,
			expectedStatus: http.StatusForbidden,
			expectNextCall: false,
			expectLogError: false,
			expectLogInfo:  true,
			logMessage:     "Request without X-Real-IP header",
		},
		{
			name:           "invalid IP in header - should return 400",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "invalid-ip",
			provideHeader:  true,
			expectedStatus: http.StatusBadRequest,
			expectNextCall: false,
			expectLogError: false,
			expectLogInfo:  true,
			logMessage:     "Invalid IP address in X-Real-IP header",
		},
		{
			name:           "IP in trusted subnet - should pass",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.100",
			provideHeader:  true,
			expectedStatus: http.StatusOK,
			expectNextCall: true,
			expectLogError: false,
			expectLogInfo:  false,
		},
		{
			name:           "IP not in trusted subnet - should return 403",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "10.0.0.1",
			provideHeader:  true,
			expectedStatus: http.StatusForbidden,
			expectNextCall: false,
			expectLogError: false,
			expectLogInfo:  true,
			logMessage:     "Request from untrusted IP",
		},
		{
			name:           "IPv6 in trusted subnet - should pass",
			trustedSubnet:  "2001:db8::/32",
			xRealIP:        "2001:db8::1",
			provideHeader:  true,
			expectedStatus: http.StatusOK,
			expectNextCall: true,
			expectLogError: false,
			expectLogInfo:  false,
		},
		{
			name:           "IPv6 not in trusted subnet - should return 403",
			trustedSubnet:  "2001:db8::/32",
			xRealIP:        "2001:db9::1",
			provideHeader:  true,
			expectedStatus: http.StatusForbidden,
			expectNextCall: false,
			expectLogError: false,
			expectLogInfo:  true,
			logMessage:     "Request from untrusted IP",
		},
		{
			name:           "single IP subnet - should pass",
			trustedSubnet:  "192.168.1.100/32",
			xRealIP:        "192.168.1.100",
			provideHeader:  true,
			expectedStatus: http.StatusOK,
			expectNextCall: true,
			expectLogError: false,
			expectLogInfo:  false,
		},
		{
			name:           "single IP subnet - should fail for different IP",
			trustedSubnet:  "192.168.1.100/32",
			xRealIP:        "192.168.1.101",
			provideHeader:  true,
			expectedStatus: http.StatusForbidden,
			expectNextCall: false,
			expectLogError: false,
			expectLogInfo:  true,
			logMessage:     "Request from untrusted IP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			logger := NewMockMiddlewareLogger(ctrl)

			if tt.expectLogError {
				logger.EXPECT().Error(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Times(1)
			}

			if tt.expectLogInfo {
				switch tt.name {
				case "missing X-Real-IP header - should return 403":
					logger.EXPECT().Info(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).Times(1)
				case "invalid IP in header - should return 400":
					logger.EXPECT().Info(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).Times(1)
				default:
					logger.EXPECT().Info(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).Times(1)
				}
			}

			var nextCalled bool
			handlerToTest := TrustedSubnetMiddleware(tt.trustedSubnet, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				testHandler.ServeHTTP(w, r)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.provideHeader {
				req.Header.Set("X-Real-IP", tt.xRealIP)
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

func TestTrustedSubnetMiddleware_EmptySubnet_PassThrough(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := NewMockMiddlewareLogger(ctrl)

	var nextCalled bool
	handler := TrustedSubnetMiddleware("", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, nextCalled)
}

func TestTrustedSubnetMiddleware_ValidIP_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := NewMockMiddlewareLogger(ctrl)

	var nextCalled bool
	handler := TrustedSubnetMiddleware("10.0.0.0/8", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-Real-IP", "10.1.2.3")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, nextCalled)
	assert.Equal(t, "success", rr.Body.String())
}

func TestTrustedSubnetMiddleware_InvalidIP_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := NewMockMiddlewareLogger(ctrl)

	logger.EXPECT().Info(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Times(1)

	var nextCalled bool
	handler := TrustedSubnetMiddleware("192.168.0.0/16", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.Header.Set("X-Real-IP", "172.16.0.1")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.False(t, nextCalled)
}

func TestTrustedSubnetMiddleware_MissingHeader_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := NewMockMiddlewareLogger(ctrl)

	logger.EXPECT().Info(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Times(1)

	var nextCalled bool
	handler := TrustedSubnetMiddleware("192.168.1.0/24", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.False(t, nextCalled)
}
