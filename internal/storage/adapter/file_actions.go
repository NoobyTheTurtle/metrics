package adapter

func (ms *MetricStorage) SaveToFile() error {
	if ms.fileSaver == nil {
		return nil
	}
	return ms.fileSaver.SaveToFile()
}

func (ms *MetricStorage) LoadFromFile() error {
	if ms.fileSaver == nil {
		return nil
	}
	return ms.fileSaver.LoadFromFile()
}
