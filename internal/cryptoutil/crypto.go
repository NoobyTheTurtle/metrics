package cryptoutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// PublicKeyProvider реализует шифрование с помощью публичного ключа
type PublicKeyProvider struct {
	publicKey *rsa.PublicKey
}

// PrivateKeyProvider реализует дешифрование с помощью приватного ключа
type PrivateKeyProvider struct {
	privateKey *rsa.PrivateKey
}

// NewPublicKeyProvider создает новый провайдер публичного ключа из PEM файла
func NewPublicKeyProvider(keyPath string) (*PublicKeyProvider, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("cryptoutil.NewPublicKeyProvider: failed to read key file '%s': %w", keyPath, err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("cryptoutil.NewPublicKeyProvider: failed to decode PEM block from file '%s'", keyPath)
	}

	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("cryptoutil.NewPublicKeyProvider: expected PUBLIC KEY, got %s", block.Type)
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cryptoutil.NewPublicKeyProvider: failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("cryptoutil.NewPublicKeyProvider: key is not an RSA public key")
	}

	return &PublicKeyProvider{publicKey: rsaPublicKey}, nil
}

// NewPrivateKeyProvider создает новый провайдер приватного ключа из PEM файла
func NewPrivateKeyProvider(keyPath string) (*PrivateKeyProvider, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("cryptoutil.NewPrivateKeyProvider: failed to read key file '%s': %w", keyPath, err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("cryptoutil.NewPrivateKeyProvider: failed to decode PEM block from file '%s'", keyPath)
	}

	if block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("cryptoutil.NewPrivateKeyProvider: expected PRIVATE KEY, got %s", block.Type)
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cryptoutil.NewPrivateKeyProvider: failed to parse private key: %w", err)
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("cryptoutil.NewPrivateKeyProvider: key is not an RSA private key")
	}

	return &PrivateKeyProvider{privateKey: rsaPrivateKey}, nil
}

// Encrypt шифрует данные с помощью публичного ключа
func (p *PublicKeyProvider) Encrypt(data []byte) ([]byte, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, p.publicKey, data)
	if err != nil {
		return nil, fmt.Errorf("cryptoutil.PublicKeyProvider.Encrypt: failed to encrypt data: %w", err)
	}
	return encrypted, nil
}

// Decrypt дешифрует данные с помощью приватного ключа
func (p *PrivateKeyProvider) Decrypt(data []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, p.privateKey, data)
	if err != nil {
		return nil, fmt.Errorf("cryptoutil.PrivateKeyProvider.Decrypt: failed to decrypt data: %w", err)
	}
	return decrypted, nil
}

// GenerateKeyPair генерирует новую пару RSA ключей и сохраняет их в файлы
func GenerateKeyPair(privateKeyPath, publicKeyPath string, bits int) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("cryptoutil.GenerateKeyPair: failed to generate private key: %w", err)
	}

	// Сохраняем приватный ключ
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("cryptoutil.GenerateKeyPair: failed to marshal private key: %w", err)
	}

	privateKeyPEM := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	if err := os.WriteFile(privateKeyPath, pem.EncodeToMemory(privateKeyPEM), 0o600); err != nil {
		return fmt.Errorf("cryptoutil.GenerateKeyPair: failed to write private key file: %w", err)
	}

	// Сохраняем публичный ключ
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("cryptoutil.GenerateKeyPair: failed to marshal public key: %w", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	if err := os.WriteFile(publicKeyPath, pem.EncodeToMemory(publicKeyPEM), 0o644); err != nil {
		return fmt.Errorf("cryptoutil.GenerateKeyPair: failed to write public key file: %w", err)
	}

	return nil
}
