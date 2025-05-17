package adapter

type MetricStorage struct {
	storage     Storage
	fileStorage FileStorage
	dbStorage   DatabaseStorage
}

func NewStorage(storage Storage) *MetricStorage {
	return &MetricStorage{
		storage:     storage,
		fileStorage: nil,
		dbStorage:   nil,
	}
}

func NewFileStorage(fileStorage FileStorage) *MetricStorage {
	return &MetricStorage{
		storage:     fileStorage,
		fileStorage: fileStorage,
		dbStorage:   nil,
	}
}

func NewDatabaseStorage(dbStorage DatabaseStorage) *MetricStorage {
	return &MetricStorage{
		storage:     dbStorage,
		fileStorage: nil,
		dbStorage:   dbStorage,
	}
}
