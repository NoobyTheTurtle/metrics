package postgres

func (c *PostgresClient) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}
