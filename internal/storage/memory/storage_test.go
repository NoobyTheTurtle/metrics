package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryStorage_Success(t *testing.T) {
	storage := NewMemoryStorage()

	assert.NotNil(t, storage)
	assert.NotNil(t, storage.data)
	assert.Empty(t, storage.data)
	assert.IsType(t, &MemoryStorage{}, storage)
}
