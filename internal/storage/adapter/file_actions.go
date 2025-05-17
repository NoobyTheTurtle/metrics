package adapter

import "context"

func (ms *MetricStorage) SaveToFile(ctx context.Context) error {
	if ms.fileStorage == nil {
		return nil
	}
	return ms.fileStorage.SaveToFile(ctx)
}

func (ms *MetricStorage) LoadFromFile(ctx context.Context) error {
	if ms.fileStorage == nil {
		return nil
	}
	return ms.fileStorage.LoadFromFile(ctx)
}
