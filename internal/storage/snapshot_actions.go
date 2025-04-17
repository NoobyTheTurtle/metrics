package storage

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/NoobyTheTurtle/metrics/internal/model"
)

func (ms *MemStorage) SaveToFile() error {
	if ms.fileStoragePath == "" {
		return nil
	}

	dir := filepath.Dir(ms.fileStoragePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for file storage: %w", err)
	}

	ms.mu.RLock()
	snapshot := model.MemSnapshot{
		Gauges:   make(map[string]float64, len(ms.gauges)),
		Counters: make(map[string]int64, len(ms.counters)),
	}

	maps.Copy(snapshot.Gauges, ms.gauges)
	maps.Copy(snapshot.Counters, ms.counters)

	ms.mu.RUnlock()

	data, err := snapshot.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if err := os.WriteFile(ms.fileStoragePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics to file: %w", err)
	}

	return nil
}

func (ms *MemStorage) LoadFromFile() error {
	if ms.fileStoragePath == "" {
		return nil
	}

	data, err := os.ReadFile(ms.fileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read metrics from file: %w", err)
	}

	var snapshot model.MemSnapshot
	if err := snapshot.UnmarshalJSON(data); err != nil {
		return fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	maps.Copy(ms.gauges, snapshot.Gauges)
	maps.Copy(ms.counters, snapshot.Counters)

	return nil
}
