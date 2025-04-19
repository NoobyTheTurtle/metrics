package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (fs *FileStorage) Get(key string) (any, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.memStorage.Get(key)
}

func (fs *FileStorage) Set(key string, value any) (any, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	result, err := fs.memStorage.Set(key, value)
	if err != nil {
		return nil, err
	}

	if fs.syncMode && fs.filePath != "" {
		if err := fs.saveToFileInternal(); err != nil {
			return nil, fmt.Errorf("failed to save to file: %w", err)
		}
	}

	return result, nil
}

func (fs *FileStorage) GetAll() map[string]any {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.memStorage.GetAll()
}

func (fs *FileStorage) SaveToFile() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	return fs.saveToFileInternal()
}

func (fs *FileStorage) saveToFileInternal() error {
	if fs.filePath == "" {
		return nil
	}

	dir := filepath.Dir(fs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for file storage: %w", err)
	}

	data := fs.memStorage.GetAll()
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(fs.filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	return nil
}

func (fs *FileStorage) LoadFromFile() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read data from file: %w", err)
	}

	var fileData map[string]any
	if err := json.Unmarshal(data, &fileData); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	for key, value := range fileData {
		_, err := fs.memStorage.Set(key, value)
		if err != nil {
			return fmt.Errorf("failed to set value for key %s: %w", key, err)
		}
	}

	return nil
}
