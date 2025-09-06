package cryptoutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGenerateKeyPair_Success(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "test_private.pem")
	publicKeyPath := filepath.Join(tempDir, "test_public.pem")

	err := GenerateKeyPair(privateKeyPath, publicKeyPath, 2048)

	assert.NoError(t, err)
	assert.FileExists(t, privateKeyPath)
	assert.FileExists(t, publicKeyPath)

	privateKeyInfo, err := os.Stat(privateKeyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), privateKeyInfo.Mode().Perm())

	publicKeyInfo, err := os.Stat(publicKeyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), publicKeyInfo.Mode().Perm())
}

func TestGenerateKeyPair_InvalidPaths_Error(t *testing.T) {
	tests := []struct {
		name           string
		privateKeyPath string
		publicKeyPath  string
		description    string
	}{
		{
			name:           "invalid private key directory",
			privateKeyPath: "/nonexistent/directory/private.pem",
			publicKeyPath:  "public.pem",
			description:    "should fail when private key directory doesn't exist",
		},
		{
			name:           "invalid public key directory",
			privateKeyPath: "private.pem",
			publicKeyPath:  "/nonexistent/directory/public.pem",
			description:    "should fail when public key directory doesn't exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				os.Remove(tt.privateKeyPath)
				os.Remove(tt.publicKeyPath)
			}()

			err := GenerateKeyPair(tt.privateKeyPath, tt.publicKeyPath, 2048)

			assert.Error(t, err, tt.description)
		})
	}
}

func TestGenerateKeyPair_WriteError(t *testing.T) {
	t.Run("readonly directory for private key", func(t *testing.T) {
		tempDir := t.TempDir()
		readonlyDir := filepath.Join(tempDir, "readonly")
		err := os.Mkdir(readonlyDir, 0o444) // только чтение
		require.NoError(t, err)
		defer os.Chmod(readonlyDir, 0o755) // восстанавливаем права для очистки

		privateKeyPath := filepath.Join(readonlyDir, "private.pem")
		publicKeyPath := filepath.Join(tempDir, "public.pem")

		err = GenerateKeyPair(privateKeyPath, publicKeyPath, 2048)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write private key file")
	})

	t.Run("readonly directory for public key", func(t *testing.T) {
		tempDir := t.TempDir()
		readonlyDir := filepath.Join(tempDir, "readonly")
		err := os.Mkdir(readonlyDir, 0o444) // только чтение
		require.NoError(t, err)
		defer os.Chmod(readonlyDir, 0o755) // восстанавливаем права для очистки

		privateKeyPath := filepath.Join(tempDir, "private.pem")
		publicKeyPath := filepath.Join(readonlyDir, "public.pem")

		err = GenerateKeyPair(privateKeyPath, publicKeyPath, 2048)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write public key file")
	})
}

func TestGenerateKeyPair_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	privateKeyPath := filepath.Join(tempDir, "test_private_permissions.pem")
	publicKeyPath := filepath.Join(tempDir, "test_public_permissions.pem")

	err := GenerateKeyPair(privateKeyPath, publicKeyPath, 2048)
	require.NoError(t, err)

	privateKeyInfo, err := os.Stat(privateKeyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), privateKeyInfo.Mode().Perm(), "private key должен иметь права 0600")

	publicKeyInfo, err := os.Stat(publicKeyPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), publicKeyInfo.Mode().Perm(), "public key должен иметь права 0644")

	_, err = NewPrivateKeyProvider(privateKeyPath)
	assert.NoError(t, err, "должна быть возможность загрузить сгенерированный private key")

	_, err = NewPublicKeyProvider(publicKeyPath)
	assert.NoError(t, err, "должна быть возможность загрузить сгенерированный public key")
}
