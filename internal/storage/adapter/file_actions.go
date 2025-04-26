package adapter

import "context"

func (ms *MetricStorage) SaveToFile(ctx context.Context) error {
	if ms.fileSaver == nil {
		return nil
	}
	return ms.fileSaver.SaveToFile(ctx)
}

func (ms *MetricStorage) LoadFromFile(ctx context.Context) error {
	if ms.fileSaver == nil {
		return nil
	}
	return ms.fileSaver.LoadFromFile(ctx)
}
