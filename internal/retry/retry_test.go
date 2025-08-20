package retry

import (
	"errors"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestWithRetries_SuccessOnFirstAttempt(t *testing.T) {
	originalDelays := retryDelays
	defer func() { retryDelays = originalDelays }()
	retryDelays = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond}

	errRetryable := errors.New("retryable error")
	checker := func(err error) bool {
		return errors.Is(err, errRetryable)
	}

	op := func() error {
		return nil
	}

	err := WithRetries(op, checker)

	assert.NoError(t, err)
}

func TestWithRetries_SuccessAfterRetries(t *testing.T) {
	originalDelays := retryDelays
	defer func() { retryDelays = originalDelays }()
	retryDelays = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond}

	errRetryable := errors.New("retryable error")
	checker := func(err error) bool {
		return errors.Is(err, errRetryable)
	}

	attempts := 0
	op := func() error {
		attempts++
		if attempts < 3 {
			return errRetryable
		}
		return nil
	}

	err := WithRetries(op, checker)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestWithRetries_FailureAfterAllRetries(t *testing.T) {
	originalDelays := retryDelays
	defer func() { retryDelays = originalDelays }()
	retryDelays = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond}

	errRetryable := errors.New("retryable error")
	checker := func(err error) bool {
		return errors.Is(err, errRetryable)
	}

	attempts := 0
	op := func() error {
		attempts++
		return errRetryable
	}

	err := WithRetries(op, checker)

	assert.Error(t, err)
	assert.Equal(t, errRetryable, err)
	assert.Equal(t, 4, attempts)
}

func TestWithRetries_FailureWithNonRetryableError(t *testing.T) {
	originalDelays := retryDelays
	defer func() { retryDelays = originalDelays }()
	retryDelays = []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond}

	errRetryable := errors.New("retryable error")
	errNonRetryable := errors.New("non-retryable error")
	checker := func(err error) bool {
		return errors.Is(err, errRetryable)
	}

	attempts := 0
	op := func() error {
		attempts++
		return errNonRetryable
	}

	err := WithRetries(op, checker)

	assert.Error(t, err)
	assert.Equal(t, errNonRetryable, err)
	assert.Equal(t, 1, attempts)
}

func TestWithRetries_EmptyDelaysSlice(t *testing.T) {
	originalDelays := retryDelays
	defer func() { retryDelays = originalDelays }()
	retryDelays = []time.Duration{}

	errRetryable := errors.New("retryable error")
	checker := func(err error) bool {
		return errors.Is(err, errRetryable)
	}

	attempts := 0
	op := func() error {
		attempts++
		return errRetryable
	}

	err := WithRetries(op, checker)

	assert.Error(t, err)
	assert.Equal(t, errRetryable, err)
	assert.Equal(t, 1, attempts)
}

func TestWithRetries_NilChecker(t *testing.T) {
	originalDelays := retryDelays
	defer func() { retryDelays = originalDelays }()
	retryDelays = []time.Duration{1 * time.Millisecond}

	testErr := errors.New("test error")
	attempts := 0
	op := func() error {
		attempts++
		return testErr
	}

	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, 1, attempts)
		}
	}()

	err := WithRetries(op, nil)

	t.Fatal("Expected panic due to nil checker, but got result:", err)
}

func TestPgErrorChecker(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "retryable pg error",
			err:      &pgconn.PgError{Code: pgerrcode.ConnectionException},
			expected: true,
		},
		{
			name:     "non-retryable pg error",
			err:      &pgconn.PgError{Code: pgerrcode.InvalidSQLStatementName},
			expected: false,
		},
		{
			name:     "non-pg error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, PgErrorChecker(tt.err))
		})
	}
}

func TestRequestErrorChecker(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "net timeout error",
			err:      &net.DNSError{IsTimeout: true},
			expected: true,
		},
		{
			name:     "connection refused error",
			err:      syscall.ECONNREFUSED,
			expected: true,
		},
		{
			name:     "other net error",
			err:      &net.DNSError{IsTimeout: false},
			expected: false,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, RequestErrorChecker(tt.err))
		})
	}
}
