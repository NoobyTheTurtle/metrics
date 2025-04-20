package adapter

type MetricStorage struct {
	storage   Storage
	fileSaver FileSaver
}

func NewMetricStorage(storage Storage, fileSaver FileSaver) *MetricStorage {
	return &MetricStorage{
		storage:   storage,
		fileSaver: fileSaver,
	}
}

func NewMetricStorageNoFile(storage Storage) *MetricStorage {
	return &MetricStorage{
		storage: storage,
	}
}
