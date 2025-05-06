package retry

import (
	"errors"
	"net"
	"syscall"
)

func RequestErrorChecker(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}

	return false
}
