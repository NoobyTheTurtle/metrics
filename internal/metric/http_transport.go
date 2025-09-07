package metric

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/NoobyTheTurtle/metrics/internal/hash"
	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/NoobyTheTurtle/metrics/internal/retry"
)

type HTTPTransport struct {
	serverURL string
	client    *http.Client
	key       string
	encrypter Encrypter
	logger    MetricsLogger
}

func NewHTTPTransport(serverAddress string, useTLS bool, key string, encrypter Encrypter, logger MetricsLogger) *HTTPTransport {
	protocol := "http"
	if useTLS {
		protocol = "https"
	}

	serverURL := fmt.Sprintf("%s://%s", protocol, serverAddress)

	return &HTTPTransport{
		serverURL: serverURL,
		client:    &http.Client{},
		key:       key,
		encrypter: encrypter,
		logger:    logger,
	}
}

func (ht *HTTPTransport) SendMetrics(ctx context.Context, metrics model.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	op := func() error {
		return ht.sendMetricsBatch(ctx, metrics)
	}

	err := retry.WithRetries(op, retry.RequestErrorChecker)
	if err != nil {
		ht.logger.Warn("Failed to send metrics batch: %v", err)
		return err
	}

	return nil
}

func (ht *HTTPTransport) sendMetricsBatch(ctx context.Context, metrics model.Metrics) error {
	jsonData, err := metrics.MarshalJSON()
	if err != nil {
		return fmt.Errorf("HTTPTransport.sendMetricsBatch: error marshaling metrics batch: %w", err)
	}

	var hashHeaderValue string
	if ht.key != "" {
		hash, hashErr := hash.CalculateSHA256(jsonData, ht.key)
		if hashErr != nil {
			ht.logger.Warn("Failed to calculate SHA256 hash for request: %v", hashErr)
		} else {
			hashHeaderValue = hash
		}
	}

	compressedData, err := compressJSON(jsonData)
	if err != nil {
		return fmt.Errorf("HTTPTransport.sendMetricsBatch: error compressing data: %w", err)
	}

	var encryptedData []byte
	if ht.encrypter != nil {
		encryptedData, err = ht.encrypter.Encrypt(compressedData)
		if err != nil {
			ht.logger.Warn("Failed to encrypt data: %v", err)
			encryptedData = compressedData
		}
	} else {
		encryptedData = compressedData
	}

	url := fmt.Sprintf("%s/updates/", ht.serverURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(encryptedData))
	if err != nil {
		return fmt.Errorf("HTTPTransport.sendMetricsBatch: error creating request: %w", err)
	}

	realIP, err := getRealIPAddress()
	if err != nil {
		return fmt.Errorf("HTTPTransport.sendMetricsBatch: error getting real IP address: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Real-IP", realIP.String())

	if hashHeaderValue != "" {
		req.Header.Set("HashSHA256", hashHeaderValue)
	}

	resp, err := ht.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTPTransport.sendMetricsBatch: error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, readErr := readResponseBody(resp)
		if readErr != nil {
			return fmt.Errorf("HTTPTransport.sendMetricsBatch: server returned status code %d, could not read body: %v", resp.StatusCode, readErr)
		}
		return fmt.Errorf("HTTPTransport.sendMetricsBatch: server returned status code %d, body: %s", resp.StatusCode, bodyText)
	}

	return nil
}

func (ht *HTTPTransport) Close() error {
	return nil
}
