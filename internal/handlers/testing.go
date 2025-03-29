package handlers

import (
	"errors"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var _ ServerStorage = (*mockStorage)(nil)

type mockStorage struct {
	gauges            map[string]float64
	counters          map[string]int64
	shouldFailGauge   bool
	shouldFailCounter bool
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *mockStorage) UpdateGauge(name string, value float64) error {
	if m.shouldFailGauge {
		return errors.New("gauge update error")
	}
	m.gauges[name] = value
	return nil
}

func (m *mockStorage) UpdateCounter(name string, value int64) error {
	if m.shouldFailCounter {
		return errors.New("counter update error")
	}
	m.counters[name] += value
	return nil
}

func (m *mockStorage) GetGauge(name string) (float64, bool) {
	value, ok := m.gauges[name]
	return value, ok
}

func (m *mockStorage) GetCounter(name string) (int64, bool) {
	value, ok := m.counters[name]
	return value, ok
}

func (m *mockStorage) GetAllGauges() map[string]float64 {
	result := make(map[string]float64, len(m.gauges))
	maps.Copy(result, m.gauges)
	return result
}

func (m *mockStorage) GetAllCounters() map[string]int64 {
	result := make(map[string]int64, len(m.gauges))
	maps.Copy(result, m.counters)
	return result
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
