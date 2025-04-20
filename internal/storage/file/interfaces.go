package file

import "github.com/NoobyTheTurtle/metrics/internal/storage/memory"

type Getter interface {
	Get(key string) (any, bool)
}

type Setter interface {
	Set(key string, value any) (any, error)
}

type GetAll interface {
	GetAll() map[string]any
}

type MemStorage interface {
	Getter
	Setter
	GetAll
}

var _ MemStorage = (*memory.MemoryStorage)(nil)
var _ MemStorage = (*MockMemStorage)(nil)
