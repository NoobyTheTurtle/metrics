package transaction

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")

	mock.ExpectBegin()
	tx, err := sqlxDB.Beginx()
	require.NoError(t, err)

	pt := NewPostgresTransaction(tx)

	assert.NotNil(t, pt)
	assert.Equal(t, tx, pt.tx)

	mock.ExpectRollback()
	_ = tx.Rollback()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresTransaction_Get(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		mockSetup     func(sqlmock.Sqlmock)
		expectedValue any
		expectedFound bool
	}{
		{
			name: "get existing float metric",
			key:  "gauge:test_metric",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"value_float", "value_int"}).
					AddRow(42.5, nil)
				mock.ExpectQuery(`SELECT value_float, value_int FROM metrics WHERE key = \$1`).
					WithArgs("gauge:test_metric").
					WillReturnRows(rows)
			},
			expectedValue: 42.5,
			expectedFound: true,
		},
		{
			name: "get existing int metric",
			key:  "counter:test_counter",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"value_float", "value_int"}).
					AddRow(nil, 100)
				mock.ExpectQuery(`SELECT value_float, value_int FROM metrics WHERE key = \$1`).
					WithArgs("counter:test_counter").
					WillReturnRows(rows)
			},
			expectedValue: int64(100),
			expectedFound: true,
		},
		{
			name: "get non-existing metric",
			key:  "gauge:nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT value_float, value_int FROM metrics WHERE key = \$1`).
					WithArgs("gauge:nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			expectedValue: nil,
			expectedFound: false,
		},
		{
			name: "database error",
			key:  "gauge:error_metric",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT value_float, value_int FROM metrics WHERE key = \$1`).
					WithArgs("gauge:error_metric").
					WillReturnError(errors.New("database connection error"))
			},
			expectedValue: nil,
			expectedFound: false,
		},
		{
			name: "both values null",
			key:  "gauge:null_metric",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"value_float", "value_int"}).
					AddRow(nil, nil)
				mock.ExpectQuery(`SELECT value_float, value_int FROM metrics WHERE key = \$1`).
					WithArgs("gauge:null_metric").
					WillReturnRows(rows)
			},
			expectedValue: nil,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")

			mock.ExpectBegin()
			tx, err := sqlxDB.Beginx()
			require.NoError(t, err)

			pt := NewPostgresTransaction(tx)

			tt.mockSetup(mock)

			ctx := context.Background()
			value, found := pt.Get(ctx, tt.key)

			assert.Equal(t, tt.expectedFound, found)
			assert.Equal(t, tt.expectedValue, value)

			mock.ExpectRollback()
			_ = tx.Rollback()
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresTransaction_Set(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         any
		mockSetup     func(sqlmock.Sqlmock)
		expectedValue any
		expectedError bool
	}{
		{
			name:  "set float metric",
			key:   "gauge:test_metric",
			value: 42.5,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"key", "value_float", "value_int"}).
					AddRow("gauge:test_metric", 42.5, nil)
				mock.ExpectQuery(`INSERT INTO metrics \(key, value_float, value_int\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(key\) DO UPDATE SET value_float = \$2, value_int = \$3 RETURNING key, value_float, value_int`).
					WithArgs("gauge:test_metric", 42.5, nil).
					WillReturnRows(rows)
			},
			expectedValue: 42.5,
			expectedError: false,
		},
		{
			name:  "set int metric",
			key:   "counter:test_counter",
			value: int64(100),
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"key", "value_float", "value_int"}).
					AddRow("counter:test_counter", nil, 100)
				mock.ExpectQuery(`INSERT INTO metrics \(key, value_float, value_int\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(key\) DO UPDATE SET value_float = \$2, value_int = \$3 RETURNING key, value_float, value_int`).
					WithArgs("counter:test_counter", nil, int64(100)).
					WillReturnRows(rows)
			},
			expectedValue: int64(100),
			expectedError: false,
		},
		{
			name:          "unsupported value type",
			key:           "gauge:test_metric",
			value:         "string_value",
			mockSetup:     func(mock sqlmock.Sqlmock) {},
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:  "database error",
			key:   "gauge:error_metric",
			value: 42.5,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO metrics \(key, value_float, value_int\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(key\) DO UPDATE SET value_float = \$2, value_int = \$3 RETURNING key, value_float, value_int`).
					WithArgs("gauge:error_metric", 42.5, nil).
					WillReturnError(errors.New("database connection error"))
			},
			expectedValue: nil,
			expectedError: true,
		},
		{
			name:  "invalid result from database",
			key:   "gauge:invalid_result",
			value: 42.5,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"key", "value_float", "value_int"}).
					AddRow("gauge:invalid_result", nil, nil)
				mock.ExpectQuery(`INSERT INTO metrics \(key, value_float, value_int\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(key\) DO UPDATE SET value_float = \$2, value_int = \$3 RETURNING key, value_float, value_int`).
					WithArgs("gauge:invalid_result", 42.5, nil).
					WillReturnRows(rows)
			},
			expectedValue: nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")

			mock.ExpectBegin()
			tx, err := sqlxDB.Beginx()
			require.NoError(t, err)

			pt := NewPostgresTransaction(tx)

			tt.mockSetup(mock)

			ctx := context.Background()
			value, err := pt.Set(ctx, tt.key, tt.value)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedValue, value)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}

			mock.ExpectRollback()
			_ = tx.Rollback()
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresTransaction_GetAll(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(sqlmock.Sqlmock)
		expectedResult map[string]any
		expectedError  bool
	}{
		{
			name: "get all metrics successfully",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"key", "value_float", "value_int"}).
					AddRow("gauge:metric1", 42.5, nil).
					AddRow("counter:metric2", nil, 100).
					AddRow("gauge:metric3", 10.0, nil)
				mock.ExpectQuery(`SELECT key, value_float, value_int FROM metrics`).
					WillReturnRows(rows)
			},
			expectedResult: map[string]any{
				"gauge:metric1":   42.5,
				"counter:metric2": int64(100),
				"gauge:metric3":   10.0,
			},
			expectedError: false,
		},
		{
			name: "empty result set",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"key", "value_float", "value_int"})
				mock.ExpectQuery(`SELECT key, value_float, value_int FROM metrics`).
					WillReturnRows(rows)
			},
			expectedResult: map[string]any{},
			expectedError:  false,
		},
		{
			name: "database error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT key, value_float, value_int FROM metrics`).
					WillReturnError(errors.New("database connection error"))
			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name: "metrics with both values null (should be filtered out)",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"key", "value_float", "value_int"}).
					AddRow("gauge:valid_metric", 42.5, nil).
					AddRow("gauge:invalid_metric", nil, nil).
					AddRow("counter:valid_counter", nil, 100)
				mock.ExpectQuery(`SELECT key, value_float, value_int FROM metrics`).
					WillReturnRows(rows)
			},
			expectedResult: map[string]any{
				"gauge:valid_metric":    42.5,
				"counter:valid_counter": int64(100),
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")

			mock.ExpectBegin()
			tx, err := sqlxDB.Beginx()
			require.NoError(t, err)

			pt := NewPostgresTransaction(tx)

			tt.mockSetup(mock)

			ctx := context.Background()
			result, err := pt.GetAll(ctx)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedResult), len(result))
				for k, v := range tt.expectedResult {
					actualValue, exists := result[k]
					assert.True(t, exists, "Key %s should exist in result", k)
					assert.Equal(t, v, actualValue, "Value for key %s should match", k)
				}
			}

			mock.ExpectRollback()
			_ = tx.Rollback()
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresTransaction_Commit(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name: "successful commit",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectCommit()
			},
			expectedError: false,
		},
		{
			name: "commit error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")

			mock.ExpectBegin()
			tx, err := sqlxDB.Beginx()
			require.NoError(t, err)

			pt := NewPostgresTransaction(tx)

			tt.mockSetup(mock)

			err = pt.Commit()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresTransaction_Rollback(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(sqlmock.Sqlmock)
		expectedError bool
	}{
		{
			name: "successful rollback",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectRollback()
			},
			expectedError: false,
		},
		{
			name: "rollback error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectRollback().WillReturnError(errors.New("rollback failed"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")

			mock.ExpectBegin()
			tx, err := sqlxDB.Beginx()
			require.NoError(t, err)

			pt := NewPostgresTransaction(tx)

			tt.mockSetup(mock)

			err = pt.Rollback()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresTransaction_IntegrationScenario(t *testing.T) {
	t.Run("complete transaction workflow", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		sqlxDB := sqlx.NewDb(db, "postgres")

		mock.ExpectBegin()
		tx, err := sqlxDB.Beginx()
		require.NoError(t, err)

		pt := NewPostgresTransaction(tx)

		ctx := context.Background()

		setRows := sqlmock.NewRows([]string{"key", "value_float", "value_int"}).
			AddRow("gauge:test_metric", 42.5, nil)
		mock.ExpectQuery(`INSERT INTO metrics \(key, value_float, value_int\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(key\) DO UPDATE SET value_float = \$2, value_int = \$3 RETURNING key, value_float, value_int`).
			WithArgs("gauge:test_metric", 42.5, nil).
			WillReturnRows(setRows)

		value, err := pt.Set(ctx, "gauge:test_metric", 42.5)
		assert.NoError(t, err)
		assert.Equal(t, 42.5, value)

		getRows := sqlmock.NewRows([]string{"value_float", "value_int"}).
			AddRow(42.5, nil)
		mock.ExpectQuery(`SELECT value_float, value_int FROM metrics WHERE key = \$1`).
			WithArgs("gauge:test_metric").
			WillReturnRows(getRows)

		retrievedValue, found := pt.Get(ctx, "gauge:test_metric")
		assert.True(t, found)
		assert.Equal(t, 42.5, retrievedValue)

		getAllRows := sqlmock.NewRows([]string{"key", "value_float", "value_int"}).
			AddRow("gauge:test_metric", 42.5, nil).
			AddRow("counter:test_counter", nil, 100)
		mock.ExpectQuery(`SELECT key, value_float, value_int FROM metrics`).
			WillReturnRows(getAllRows)

		allMetrics, err := pt.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, allMetrics, 2)
		assert.Equal(t, 42.5, allMetrics["gauge:test_metric"])
		assert.Equal(t, int64(100), allMetrics["counter:test_counter"])

		mock.ExpectCommit()
		err = pt.Commit()
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
