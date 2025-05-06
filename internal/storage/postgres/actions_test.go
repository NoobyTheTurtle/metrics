package postgres

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

func TestPostgresStorage_Get(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		setupMock     func(sqlmock.Sqlmock)
		expectedValue any
		expectedFound bool
	}{
		{
			name: "get existing float value",
			key:  "floatKey",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"value_float", "value_int"}).
					AddRow(42.5, nil)
				mock.ExpectQuery("SELECT value_float, value_int FROM metrics WHERE key = \\$1").
					WithArgs("floatKey").
					WillReturnRows(rows)
			},
			expectedValue: 42.5,
			expectedFound: true,
		},
		{
			name: "get existing int value",
			key:  "intKey",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"value_float", "value_int"}).
					AddRow(nil, 42)
				mock.ExpectQuery("SELECT value_float, value_int FROM metrics WHERE key = \\$1").
					WithArgs("intKey").
					WillReturnRows(rows)
			},
			expectedValue: int64(42),
			expectedFound: true,
		},
		{
			name: "get non-existing value",
			key:  "notFound",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT value_float, value_int FROM metrics WHERE key = \\$1").
					WithArgs("notFound").
					WillReturnError(sql.ErrNoRows)
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

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			tt.setupMock(mock)

			ps := &PostgresStorage{db: sqlxDB}
			ctx := context.Background()

			value, found := ps.Get(ctx, tt.key)

			assert.Equal(t, tt.expectedFound, found)
			if tt.expectedFound {
				assert.Equal(t, tt.expectedValue, value)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresStorage_Set(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         any
		setupMock     func(sqlmock.Sqlmock)
		expectedValue any
		expectedError error
	}{
		{
			name:  "set float value",
			key:   "floatKey",
			value: 42.5,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"value_float", "value_int"}).
					AddRow(42.5, nil)
				mock.ExpectQuery("INSERT INTO metrics").
					WithArgs("floatKey", sql.NullFloat64{Float64: 42.5, Valid: true}, sql.NullInt64{}).
					WillReturnRows(rows)
			},
			expectedValue: 42.5,
			expectedError: nil,
		},
		{
			name:  "set int value",
			key:   "intKey",
			value: int64(42),
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"value_float", "value_int"}).
					AddRow(nil, 42)
				mock.ExpectQuery("INSERT INTO metrics").
					WithArgs("intKey", sql.NullFloat64{}, sql.NullInt64{Int64: 42, Valid: true}).
					WillReturnRows(rows)
			},
			expectedValue: int64(42),
			expectedError: nil,
		},
		{
			name:  "database error",
			key:   "errorKey",
			value: 42.5,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO metrics").
					WithArgs("errorKey", sql.NullFloat64{Float64: 42.5, Valid: true}, sql.NullInt64{}).
					WillReturnError(errors.New("database error"))
			},
			expectedValue: nil,
			expectedError: errors.New("failed to set metric: database error"),
		},
		{
			name:  "unsupported type",
			key:   "stringKey",
			value: "string value",
			setupMock: func(mock sqlmock.Sqlmock) {
			},
			expectedValue: nil,
			expectedError: errors.New("unsupported value type: string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			tt.setupMock(mock)

			ps := &PostgresStorage{db: sqlxDB}
			ctx := context.Background()

			value, err := ps.Set(ctx, tt.key, tt.value)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresStorage_GetAll(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(sqlmock.Sqlmock)
		expectedData  map[string]any
		expectedError error
	}{
		{
			name: "get all metrics",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"key", "value_float", "value_int"}).
					AddRow("floatKey", 42.5, nil).
					AddRow("intKey", nil, 42)
				mock.ExpectQuery("SELECT key, value_float, value_int FROM metrics").
					WillReturnRows(rows)
			},
			expectedData: map[string]any{
				"floatKey": 42.5,
				"intKey":   int64(42),
			},
			expectedError: nil,
		},
		{
			name: "empty result",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"key", "value_float", "value_int"})
				mock.ExpectQuery("SELECT key, value_float, value_int FROM metrics").
					WillReturnRows(rows)
			},
			expectedData:  map[string]any{},
			expectedError: nil,
		},
		{
			name: "database error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT key, value_float, value_int FROM metrics").
					WillReturnError(errors.New("database error"))
			},
			expectedData:  nil,
			expectedError: errors.New("failed to get all metrics: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			tt.setupMock(mock)

			ps := &PostgresStorage{db: sqlxDB}
			ctx := context.Background()

			data, err := ps.GetAll(ctx)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedData), len(data))
				for k, v := range tt.expectedData {
					assert.Equal(t, v, data[k])
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
