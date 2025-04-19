package file

import "sync"

type FileStorage struct {
	mu         sync.RWMutex
	memStorage MemStorage
	filePath   string
	syncMode   bool
}

func NewFileStorage(memStorage MemStorage, filePath string, syncMode bool) *FileStorage {
	return &FileStorage{
		memStorage: memStorage,
		filePath:   filePath,
		syncMode:   syncMode,
	}
}
