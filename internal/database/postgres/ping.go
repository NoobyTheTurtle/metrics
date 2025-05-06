package postgres

import (
	"context"
	"fmt"
)

func (c *PostgresClient) Ping(ctx context.Context) error {
	if c.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return c.DB.PingContext(ctx)
}
