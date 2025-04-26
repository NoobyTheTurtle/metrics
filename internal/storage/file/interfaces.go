package file

import (
	"context"

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

type MemStorage interface {
	Getter
	Setter
	GetAll
}

var _ MemStorage = (*memory.MemoryStorage)(nil)
var _ MemStorage = (*MockMemStorage)(nil)
