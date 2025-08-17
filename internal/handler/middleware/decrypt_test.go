package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NoobyTheTurtle/metrics/internal/cryptoutil"
)

func TestDecryptMiddleware(t *testing.T) {
	privateKeyPath := "test_private_key.pem"
	publicKeyPath := "test_public_key.pem"

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}()

	err := cryptoutil.GenerateKeyPair(privateKeyPath, publicKeyPath, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	publicProvider, err := cryptoutil.NewPublicKeyProvider(publicKeyPath)
	if err != nil {
		t.Fatalf("Failed to create public key provider: %v", err)
	}

	privateProvider, err := cryptoutil.NewPrivateKeyProvider(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to create private key provider: %v", err)
	}

	tests := []struct {
		name           string
		decrypter      Decrypter
		originalData   []byte
		encryptData    bool
		expectedStatus int
	}{
		{
			name:           "No decrypter",
			decrypter:      nil,
			originalData:   []byte(`{"test": "data"}`),
			encryptData:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Encrypted data",
			decrypter:      privateProvider,
			originalData:   []byte(`{"test": "encrypted"}`),
			encryptData:    true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unencrypted data with decrypter",
			decrypter:      privateProvider,
			originalData:   []byte(`{"test": "unencrypted"}`),
			encryptData:    false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testData := tt.originalData

			if tt.encryptData && tt.decrypter != nil {
				encrypted, err := publicProvider.Encrypt(testData)
				if err != nil {
					t.Fatalf("Failed to encrypt test data: %v", err)
				}
				testData = encrypted
			}

			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(testData))

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Failed to read body", http.StatusInternalServerError)
					return
				}

				if tt.encryptData && tt.decrypter != nil {
					if string(body) != string(tt.originalData) {
						t.Errorf("Decrypted data does not match original. Expected: %s, Got: %s", string(tt.originalData), string(body))
					}
				} else {
					if string(body) != string(tt.originalData) {
						t.Errorf("Data does not match original. Expected: %s, Got: %s", string(tt.originalData), string(body))
					}
				}

				w.WriteHeader(http.StatusOK)
			})

			middleware := DecryptMiddleware(tt.decrypter)
			middleware(handler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestDecryptMiddlewareErrors(t *testing.T) {
	privateKeyPath := "test_private_key.pem"
	publicKeyPath := "test_public_key.pem"

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}()

	err := cryptoutil.GenerateKeyPair(privateKeyPath, publicKeyPath, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	privateProvider, err := cryptoutil.NewPrivateKeyProvider(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to create private key provider: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/test", &errorReader{})
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := DecryptMiddleware(privateProvider)
	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
