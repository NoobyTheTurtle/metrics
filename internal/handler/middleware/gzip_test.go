package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
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
		name           string
		acceptEncoding string
		contentType    string
		responseBody   string
		compressed     bool
	}{
		{
			name:           "No compression for client without gzip support",
			acceptEncoding: "deflate",
			contentType:    "application/json",
			responseBody:   `{"status":"success"}`,
			compressed:     false,
		},
		{
			name:           "No compression for non-compressible content",
			acceptEncoding: "gzip",
			contentType:    "image/png",
			responseBody:   "binary-data",
			compressed:     false,
		},
		{
			name:           "Compress JSON response",
			acceptEncoding: "gzip",
			contentType:    "application/json",
			responseBody:   `{"status":"success","data":{"id":1,"name":"test"}}`,
			compressed:     true,
		},
		{
			name:           "Compress text response",
			acceptEncoding: "gzip, deflate",
			contentType:    "text/plain",
			responseBody:   "This is some example text that should be compressed with gzip",
			compressed:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tc.responseBody))
			})

			gzipHandler := GzipMiddleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", tc.acceptEncoding)

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
