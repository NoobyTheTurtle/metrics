package postgres

import (
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresClient_runMigrations(t *testing.T) {
	t.Run("nil database connection", func(t *testing.T) {
		client := &PostgresClient{DB: nil}

		err := client.runMigrations()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection is nil")
	})

	t.Run("migration path validation", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("test_db"))
		mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_schema"}).AddRow("public"))

		mock.ExpectQuery(`SELECT pg_advisory_lock`).WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_lock"}).AddRow(1))
		mock.ExpectQuery(`SELECT version, dirty FROM "schema_migrations" LIMIT 1`).WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}))

		sqlxDB := sqlx.NewDb(mockDB, "postgres")
		client := &PostgresClient{DB: sqlxDB}

		err = client.runMigrations()

		assert.Error(t, err)
		assert.True(t,
			strings.Contains(err.Error(), "failed to create migration driver") ||
				strings.Contains(err.Error(), "failed to create migration instance"))
	})

	t.Run("migration driver creation failure", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnError(errors.New("database query failed"))

		sqlxDB := sqlx.NewDb(mockDB, "postgres")
		client := &PostgresClient{DB: sqlxDB}

		err = client.runMigrations()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create migration driver")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("migration driver schema query failure", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("test_db"))
		mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnError(errors.New("schema query failed"))

		sqlxDB := sqlx.NewDb(mockDB, "postgres")
		client := &PostgresClient{DB: sqlxDB}

		err = client.runMigrations()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create migration driver")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresClient_runMigrations_Integration(t *testing.T) {
	t.Run("database connection type validation", func(t *testing.T) {
		mockDB, _, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		client := &PostgresClient{DB: sqlx.NewDb(mockDB, "mysql")}

		err = client.runMigrations()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create migration driver")
	})

	t.Run("real migration path behavior", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("test_db"))
		mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_schema"}).AddRow("public"))

		sqlxDB := sqlx.NewDb(mockDB, "postgres")
		client := &PostgresClient{DB: sqlxDB}

		err = client.runMigrations()

		assert.Error(t, err)
		assert.True(t,
			strings.Contains(err.Error(), "failed to create migration driver") ||
				strings.Contains(err.Error(), "failed to create migration instance"))
	})
}

func TestPostgresClient_runMigrations_EdgeCases(t *testing.T) {
	t.Run("error handling validation", func(t *testing.T) {
		tests := []struct {
			name          string
			setupMock     func(sqlmock.Sqlmock)
			expectedError string
		}{
			{
				name: "database query fails",
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnError(errors.New("connection error"))
				},
				expectedError: "failed to create migration driver",
			},
			{
				name: "schema query fails",
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("test_db"))
					mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnError(errors.New("schema error"))
				},
				expectedError: "failed to create migration driver",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer mockDB.Close()

				tt.setupMock(mock)

				sqlxDB := sqlx.NewDb(mockDB, "postgres")
				client := &PostgresClient{DB: sqlxDB}

				err = client.runMigrations()

				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})

	t.Run("driver creation with different database names", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("metrics_db"))
		mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_schema"}).AddRow("metrics_schema"))

		sqlxDB := sqlx.NewDb(mockDB, "postgres")
		client := &PostgresClient{DB: sqlxDB}

		err = client.runMigrations()

		assert.Error(t, err)
		assert.True(t,
			strings.Contains(err.Error(), "failed to create migration driver") ||
				strings.Contains(err.Error(), "failed to create migration instance"))
	})
}

func TestPostgresClient_runMigrations_ErrorTypes(t *testing.T) {
	t.Run("specific error messages", func(t *testing.T) {
		tests := []struct {
			name          string
			setupClient   func() *PostgresClient
			expectedError string
		}{
			{
				name: "nil database connection",
				setupClient: func() *PostgresClient {
					return &PostgresClient{DB: nil}
				},
				expectedError: "database connection is nil",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				client := tt.setupClient()
				err := client.runMigrations()

				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			})
		}
	})
}

func TestPostgresClient_runMigrations_ComprehensiveCoverage(t *testing.T) {
	t.Run("migration path does not exist", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("test_db"))
		mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_schema"}).AddRow("public"))
		mock.ExpectQuery(`SELECT pg_advisory_lock`).WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_lock"}).AddRow(true))
		mock.ExpectQuery(`SELECT version, dirty FROM "schema_migrations" LIMIT 1`).WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}))

		sqlxDB := sqlx.NewDb(mockDB, "postgres")
		client := &PostgresClient{DB: sqlxDB}

		err = client.runMigrations()

		assert.Error(t, err)
		assert.True(t,
			strings.Contains(err.Error(), "failed to create migration instance") ||
				strings.Contains(err.Error(), "failed to create migration driver"))
	})

	t.Run("various database states", func(t *testing.T) {
		tests := []struct {
			name         string
			databaseName string
			schemaName   string
			expectError  bool
		}{
			{
				name:         "standard postgres database",
				databaseName: "postgres",
				schemaName:   "public",
				expectError:  true,
			},
			{
				name:         "custom database name",
				databaseName: "custom_db",
				schemaName:   "custom_schema",
				expectError:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer mockDB.Close()

				mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow(tt.databaseName))
				mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_schema"}).AddRow(tt.schemaName))

				mock.ExpectQuery(`SELECT pg_advisory_lock`).WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_lock"}).AddRow(true))
				mock.ExpectQuery(`SELECT version, dirty FROM "schema_migrations" LIMIT 1`).WillReturnRows(sqlmock.NewRows([]string{"version", "dirty"}))

				sqlxDB := sqlx.NewDb(mockDB, "postgres")
				client := &PostgresClient{DB: sqlxDB}

				err = client.runMigrations()

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("error conditions coverage", func(t *testing.T) {
		tests := []struct {
			name          string
			setupMock     func(sqlmock.Sqlmock)
			expectedError string
		}{
			{
				name: "advisory lock fails",
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("test_db"))
					mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_schema"}).AddRow("public"))
					mock.ExpectQuery(`SELECT pg_advisory_lock`).WillReturnError(errors.New("advisory lock failed"))
				},
				expectedError: "failed to create migration driver",
			},
			{
				name: "version table query fails",
				setupMock: func(mock sqlmock.Sqlmock) {
					mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_database"}).AddRow("test_db"))
					mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnRows(sqlmock.NewRows([]string{"current_schema"}).AddRow("public"))
					mock.ExpectQuery(`SELECT pg_advisory_lock`).WillReturnRows(sqlmock.NewRows([]string{"pg_advisory_lock"}).AddRow(true))
					mock.ExpectQuery(`SELECT version, dirty FROM "schema_migrations" LIMIT 1`).WillReturnError(errors.New("version table error"))
				},
				expectedError: "failed to create migration driver",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				defer mockDB.Close()

				tt.setupMock(mock)

				sqlxDB := sqlx.NewDb(mockDB, "postgres")
				client := &PostgresClient{DB: sqlxDB}

				err = client.runMigrations()

				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			})
		}
	})
}

func TestPostgresClient_runMigrations_ErrorMessageFormat(t *testing.T) {
	t.Run("nil database error message format", func(t *testing.T) {
		client := &PostgresClient{DB: nil}
		err := client.runMigrations()

		assert.Error(t, err)
		assert.Equal(t, "postgres.PostgresClient.runMigrations: database connection is nil", err.Error())
	})

	t.Run("driver creation error message format", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnError(errors.New("test error"))

		sqlxDB := sqlx.NewDb(mockDB, "postgres")
		client := &PostgresClient{DB: sqlxDB}

		err = client.runMigrations()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "postgres.PostgresClient.runMigrations: failed to create migration driver:")
		assert.Contains(t, err.Error(), "test error")
	})
}

func TestPostgresClient_runMigrations_MethodBehavior(t *testing.T) {
	t.Run("testing method contract", func(t *testing.T) {
		client := &PostgresClient{DB: nil}
		err := client.runMigrations()

		assert.Error(t, err)
		assert.NotNil(t, client)
	})

	t.Run("method error handling consistency", func(t *testing.T) {
		tests := []struct {
			name        string
			client      *PostgresClient
			wantErr     bool
			errContains string
		}{
			{
				name:        "nil client db",
				client:      &PostgresClient{DB: nil},
				wantErr:     true,
				errContains: "database connection is nil",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.client.runMigrations()
				if tt.wantErr {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errContains)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("postgres config validation", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		scenarios := []struct {
			name       string
			dbName     string
			expectCall bool
		}{
			{"empty_db_name", "", true},
			{"valid_db_name", "metrics", true},
			{"unicode_db_name", "мetrics_тест", true},
		}

		for _, sc := range scenarios {
			t.Run(sc.name, func(t *testing.T) {
				if sc.expectCall {
					mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(
						sqlmock.NewRows([]string{"current_database"}).AddRow(sc.dbName))
					mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnError(
						errors.New("test error to prevent further execution"))
				}

				sqlxDB := sqlx.NewDb(mockDB, "postgres")
				client := &PostgresClient{DB: sqlxDB}

				err := client.runMigrations()

				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to create migration driver")
			})
		}
	})
}

func TestPostgresClient_runMigrations_Coverage(t *testing.T) {
	t.Run("driver name consistency", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(
			sqlmock.NewRows([]string{"current_database"}).AddRow("test_db"))
		mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnRows(
			sqlmock.NewRows([]string{"current_schema"}).AddRow("public"))

		sqlxDB := sqlx.NewDb(mockDB, "mysql")
		client := &PostgresClient{DB: sqlxDB}

		err = client.runMigrations()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create migration driver")
	})

	t.Run("file path validation test", func(t *testing.T) {
		mockDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer mockDB.Close()

		mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnRows(
			sqlmock.NewRows([]string{"current_database"}).AddRow("test_db"))
		mock.ExpectQuery(`SELECT CURRENT_SCHEMA\(\)`).WillReturnRows(
			sqlmock.NewRows([]string{"current_schema"}).AddRow("public"))

		sqlxDB := sqlx.NewDb(mockDB, "postgres")
		client := &PostgresClient{DB: sqlxDB}

		err = client.runMigrations()

		assert.Error(t, err)
		assert.True(t,
			strings.Contains(err.Error(), "failed to create migration driver") ||
				strings.Contains(err.Error(), "failed to create migration instance"),
			"Error should be about driver or instance creation, got: %v", err)
	})
}

func BenchmarkPostgresClient_runMigrations_NilDB(b *testing.B) {
	client := &PostgresClient{DB: nil}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.runMigrations()
	}
}

func BenchmarkPostgresClient_runMigrations_DriverError(b *testing.B) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(b, err)
	defer mockDB.Close()

	for i := 0; i < b.N; i++ {
		mock.ExpectQuery(`SELECT CURRENT_DATABASE\(\)`).WillReturnError(errors.New("benchmark error"))
	}

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	client := &PostgresClient{DB: sqlxDB}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.runMigrations()
	}
}
