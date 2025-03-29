package logger

import (
	"bytes"
	"fmt"
	"sync"
)

type MockLogger struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		buffer: bytes.Buffer{},
	}
}

func (l *MockLogger) Info(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buffer.WriteString(fmt.Sprintf("[INFO] "+format+"\n", args...))
}

func (l *MockLogger) Error(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buffer.WriteString(fmt.Sprintf("[ERROR] "+format+"\n", args...))
}

func (l *MockLogger) Debug(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buffer.WriteString(fmt.Sprintf("[DEBUG] "+format+"\n", args...))
}

func (l *MockLogger) Warn(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buffer.WriteString(fmt.Sprintf("[WARN] "+format+"\n", args...))
}

func (l *MockLogger) GetOutput() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buffer.String()
}
