package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TODO: Add mutex for file operations

func (fs *FileStorage) Get(key string) (any, bool) {
	return fs.memStorage.Get(key)
}

func (fs *FileStorage) Set(key string, value any) (any, error) {
	result, err := fs.memStorage.Set(key, value)
	if err != nil {
		return nil, err
	}

	if fs.syncMode && fs.filePath != "" {
		if err := fs.SaveToFile(); err != nil {
			return nil, fmt.Errorf("failed to save to file: %w", err)
		}
	}

	return result, nil
}

func (fs *FileStorage) GetAll() map[string]any {
	return fs.memStorage.GetAll()
}

func (fs *FileStorage) SaveToFile() error {
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
