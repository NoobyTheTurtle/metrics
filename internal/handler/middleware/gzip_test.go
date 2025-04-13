package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldCompress(t *testing.T) {
	testCases := []struct {
		name           string
		acceptEncoding string
		contentType    string
		expected       bool
	}{
		{
			name:           "No gzip in Accept-Encoding",
			acceptEncoding: "deflate, br",
			contentType:    "application/json",
			expected:       false,
		},
		{
			name:           "Has gzip but non-compressible content type",
			acceptEncoding: "gzip, deflate, br",
			contentType:    "image/jpeg",
			expected:       false,
		},
		{
			name:           "JSON content type with gzip",
			acceptEncoding: "gzip, deflate, br",
			contentType:    "application/json",
			expected:       true,
		},
		{
			name:           "Text content type with gzip",
			acceptEncoding: "gzip, deflate, br",
			contentType:    "text/plain",
			expected:       true,
		},
		{
			name:           "JavaScript content type with gzip",
			acceptEncoding: "gzip, deflate, br",
			contentType:    "application/javascript",
			expected:       true,
		},
		{
			name:           "Has gzip but empty content type",
			acceptEncoding: "gzip, deflate",
			contentType:    "",
			expected:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", tc.acceptEncoding)

			result := shouldCompress(req, tc.contentType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGzipMiddleware(t *testing.T) {
	testCases := []struct {
		name            string
		method          string
		path            string
		acceptEncoding  string
		contentType     string
		requestBody     string
		compressRequest bool
		responseBody    string
		compressed      bool
	}{
		{
			name:           "No compression for client without gzip support",
			method:         http.MethodGet,
			path:           "/test",
			acceptEncoding: "deflate",
			contentType:    "application/json",
			responseBody:   `{"status":"success"}`,
			compressed:     false,
		},
		{
			name:           "No compression for non-compressible content",
			method:         http.MethodGet,
			path:           "/image",
			acceptEncoding: "gzip",
			contentType:    "image/png",
			responseBody:   "binary-data",
			compressed:     false,
		},
		{
			name:           "Compress JSON response",
			method:         http.MethodGet,
			path:           "/api/data",
			acceptEncoding: "gzip",
			contentType:    "application/json",
			responseBody:   `{"status":"success","data":{"id":1,"name":"test"}}`,
			compressed:     true,
		},
		{
			name:           "Compress text response",
			method:         http.MethodGet,
			path:           "/text",
			acceptEncoding: "gzip, deflate",
			contentType:    "text/plain",
			responseBody:   "Text that should be compressed",
			compressed:     true,
		},
		{
			name:           "Compress JavaScript response",
			method:         http.MethodGet,
			path:           "/js",
			acceptEncoding: "gzip",
			contentType:    "application/javascript",
			responseBody:   "function test() { return 'Hello, world!'; }",
			compressed:     true,
		},
		{
			name:            "Process gzipped request",
			method:          http.MethodPost,
			path:            "/api/update",
			acceptEncoding:  "",
			contentType:     "application/json",
			requestBody:     `{"key":"value"}`,
			compressRequest: true,
			responseBody:    "received",
			compressed:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.compressRequest {
					body, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Equal(t, tc.requestBody, string(body))
				}

				w.Write([]byte(tc.responseBody))
			})

			gzipHandler := GzipMiddleware(handler)

			var requestBody io.Reader = nil
			if tc.requestBody != "" {
				if tc.compressRequest {
					var buf bytes.Buffer
					gzWriter := gzip.NewWriter(&buf)
					_, err := gzWriter.Write([]byte(tc.requestBody))
					require.NoError(t, err)
					require.NoError(t, gzWriter.Close())
					requestBody = &buf
				} else {
					requestBody = strings.NewReader(tc.requestBody)
				}
			}

			req := httptest.NewRequest(tc.method, tc.path, requestBody)
			if tc.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tc.acceptEncoding)
			}
			if tc.compressRequest {
				req.Header.Set("Content-Encoding", "gzip")
			}

			recorder := httptest.NewRecorder()
			recorder.Header().Set("Content-Type", tc.contentType)

			gzipHandler.ServeHTTP(recorder, req)

			if tc.compressed {
				assert.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))

				reader, err := gzip.NewReader(recorder.Body)
				require.NoError(t, err)
				defer reader.Close()

				decompressed, err := io.ReadAll(reader)
				require.NoError(t, err)

				assert.Equal(t, tc.responseBody, string(decompressed))
			} else {
				assert.Equal(t, "", recorder.Header().Get("Content-Encoding"))
				assert.Equal(t, tc.responseBody, recorder.Body.String())
			}
		})
	}
}

func TestGzipMiddleware_InvalidGzip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with invalid gzip data")
	})

	gzipHandler := GzipMiddleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("invalid gzip data")))
	req.Header.Set("Content-Encoding", "gzip")

	recorder := httptest.NewRecorder()
	gzipHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "Invalid gzip body", recorder.Body.String())
}
