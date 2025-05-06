package retry

import (
	"time"
)

type (
	Operation func() error
	Checker   func(err error) bool
)

var retryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func WithRetries(op Operation, checker Checker) error {
	var err error
	var attempt int

	err = op()
	if err == nil {
		return nil
	}

	if !checker(err) {
		return err
	}

	for _, delay := range retryDelays {
		attempt++
		time.Sleep(delay)

		err = op()
		if err == nil {
			return nil
		}

		if !checker(err) {
			return err
		}
	}

	return err
}
