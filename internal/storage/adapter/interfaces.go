package adapter

import (
	"context"

	"github.com/NoobyTheTurtle/metrics/internal/storage/file"
	"github.com/NoobyTheTurtle/metrics/internal/storage/memory"
)

type Getter interface {
	Get(ctx context.Context, key string) (any, bool)
}

type Setter interface {
	Set(ctx context.Context, key string, value any) (any, error)
}

type GetAll interface {
	GetAll(ctx context.Context) (map[string]any, error)
}

type Saver interface {
	SaveToFile(ctx context.Context) error
}

type Loader interface {
	LoadFromFile(ctx context.Context) error
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
