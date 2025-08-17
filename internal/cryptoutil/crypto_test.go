package cryptoutil

import (
	"os"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	privateKeyPath := "test_private_key.pem"
	publicKeyPath := "test_public_key.pem"

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}()

	err := GenerateKeyPair(privateKeyPath, publicKeyPath, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		t.Errorf("Private key file was not created")
	}

	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		t.Errorf("Public key file was not created")
	}
}

func TestPublicKeyProvider(t *testing.T) {
	privateKeyPath := "test_private_key.pem"
	publicKeyPath := "test_public_key.pem"

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}()

	err := GenerateKeyPair(privateKeyPath, publicKeyPath, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	provider, err := NewPublicKeyProvider(publicKeyPath)
	if err != nil {
		t.Fatalf("Failed to create public key provider: %v", err)
	}

	testData := []byte("Hello, World!")
	encrypted, err := provider.Encrypt(testData)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	if len(encrypted) == 0 {
		t.Errorf("Encrypted data is empty")
	}
}

func TestPrivateKeyProvider(t *testing.T) {
	privateKeyPath := "test_private_key.pem"
	publicKeyPath := "test_public_key.pem"

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}()

	err := GenerateKeyPair(privateKeyPath, publicKeyPath, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	_, err = NewPrivateKeyProvider(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to create private key provider: %v", err)
	}
}

func TestEncryptionDecryption(t *testing.T) {
	privateKeyPath := "test_private_key.pem"
	publicKeyPath := "test_public_key.pem"

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}()

	err := GenerateKeyPair(privateKeyPath, publicKeyPath, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	publicProvider, err := NewPublicKeyProvider(publicKeyPath)
	if err != nil {
		t.Fatalf("Failed to create public key provider: %v", err)
	}

	privateProvider, err := NewPrivateKeyProvider(privateKeyPath)
	if err != nil {
		t.Fatalf("Failed to create private key provider: %v", err)
	}

	testData := []byte("Hello, World! This is a test message for encryption.")
	encrypted, err := publicProvider.Encrypt(testData)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	decrypted, err := privateProvider.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	if string(decrypted) != string(testData) {
		t.Errorf("Decrypted data does not match original. Expected: %s, Got: %s", string(testData), string(decrypted))
	}
}

func TestInvalidKeyFiles(t *testing.T) {
	_, err := NewPublicKeyProvider("non_existent_file.pem")
	if err == nil {
		t.Errorf("Expected error when reading non-existent file")
	}

	_, err = NewPrivateKeyProvider("non_existent_file.pem")
	if err == nil {
		t.Errorf("Expected error when reading non-existent file")
	}

	invalidFile := "invalid.pem"
	err = os.WriteFile(invalidFile, []byte("invalid pem content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create invalid PEM file: %v", err)
	}
	defer os.Remove(invalidFile)

	_, err = NewPublicKeyProvider(invalidFile)
	if err == nil {
		t.Errorf("Expected error when reading invalid PEM file")
	}

	_, err = NewPrivateKeyProvider(invalidFile)
	if err == nil {
		t.Errorf("Expected error when reading invalid PEM file")
	}
}
