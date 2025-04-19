package adapter

import (
	"github.com/NoobyTheTurtle/metrics/internal/storage/file"
	"github.com/NoobyTheTurtle/metrics/internal/storage/memory"
)

type Getter interface {
	Get(key string) (any, bool)
}

type Setter interface {
	Set(key string, value any) (any, error)
}

type GetAll interface {
	GetAll() map[string]any
}

type Saver interface {
	SaveToFile() error
}

type Loader interface {
	LoadFromFile() error
}

type Storage interface {
	Getter
	Setter
	GetAll
}

type FileSaver interface {
	Saver
	Loader
}

var _ Storage = (*memory.MemoryStorage)(nil)
var _ Storage = (*MockStorage)(nil)

var _ FileSaver = (*file.FileStorage)(nil)
var _ FileSaver = (*MockFileSaver)(nil)
