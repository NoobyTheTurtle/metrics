package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

var (
	ErrKeyEmpty = errors.New("key is empty")
)

func CalculateSHA256(data []byte, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("sha256.CalculateSHA256: %w", ErrKeyEmpty)
	}
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(data)
	if err != nil {
		return "", fmt.Errorf("sha256.CalculateSHA256: failed to write data to hmac: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
