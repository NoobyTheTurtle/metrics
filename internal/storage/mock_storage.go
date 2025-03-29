package storage

import (
	"errors"
	"maps"
)

type MockStorage struct {
	gauges            map[string]float64
	counters          map[string]int64
	shouldFailGauge   bool
	shouldFailCounter bool
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MockStorage) UpdateGauge(name string, value float64) error {
	if m.shouldFailGauge {
		return errors.New("gauge update error")
	}
	m.gauges[name] = value
	return nil
}

func (m *MockStorage) UpdateCounter(name string, value int64) error {
	if m.shouldFailCounter {
		return errors.New("counter update error")
	}
	m.counters[name] += value
	return nil
}

func (m *MockStorage) GetGauge(name string) (float64, bool) {
	value, ok := m.gauges[name]
	return value, ok
}

func (m *MockStorage) GetCounter(name string) (int64, bool) {
	value, ok := m.counters[name]
	return value, ok
}

func (m *MockStorage) GetAllGauges() map[string]float64 {
	result := make(map[string]float64, len(m.gauges))
	maps.Copy(result, m.gauges)
	return result
}

func (m *MockStorage) GetAllCounters() map[string]int64 {
	result := make(map[string]int64, len(m.counters))
	maps.Copy(result, m.counters)
	return result
}

func (m *MockStorage) SetShouldFailGauge(shouldFail bool) {
	m.shouldFailGauge = shouldFail
}

func (m *MockStorage) SetShouldFailCounter(shouldFail bool) {
	m.shouldFailCounter = shouldFail
}
