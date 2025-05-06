package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMetricStorage_GetGauge(t *testing.T) {
	tests := []struct {
		name          string
		metricName    string
		mockValue     any
		mockFound     bool
		expectedValue float64
		expectedFound bool
	}{
		{
			name:          "get existing gauge metric",
			metricName:    "test",
			mockValue:     42.5,
			mockFound:     true,
			expectedValue: 42.5,
			expectedFound: true,
		},
		{
			name:          "get non-existing gauge metric",
			metricName:    "not-exist",
			mockValue:     nil,
			mockFound:     false,
			expectedValue: 0,
			expectedFound: false,
		},
		{
			name:          "get with conversion error",
			metricName:    "test",
			mockValue:     "not a number",
			mockFound:     true,
			expectedValue: 0,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			mockStorage.EXPECT().
				Get(gomock.Any(), addPrefix(tt.metricName, GaugePrefix)).
				Return(tt.mockValue, tt.mockFound)

			ms := &MetricStorage{
				storage: mockStorage,
			}

			ctx := context.Background()
			value, found := ms.GetGauge(ctx, tt.metricName)

			assert.Equal(t, tt.expectedFound, found)
			if tt.expectedFound {
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestMetricStorage_UpdateGauge(t *testing.T) {
	tests := []struct {
		name          string
		metricName    string
		value         float64
		mockReturn    any
		mockError     error
		expectedValue float64
		expectedError bool
	}{
		{
			name:          "update gauge successfully",
			metricName:    "test",
			value:         42.5,
			mockReturn:    42.5,
			mockError:     nil,
			expectedValue: 42.5,
			expectedError: false,
		},
		{
			name:          "update gauge with storage error",
			metricName:    "test",
			value:         42.5,
			mockReturn:    nil,
			mockError:     errors.New("storage error"),
			expectedValue: 0,
			expectedError: true,
		},
		{
			name:          "update gauge with conversion error",
			metricName:    "test",
			value:         42.5,
			mockReturn:    "not a number",
			mockError:     nil,
			expectedValue: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			mockStorage.EXPECT().
				Set(gomock.Any(), addPrefix(tt.metricName, GaugePrefix), tt.value).
				Return(tt.mockReturn, tt.mockError)

			ms := &MetricStorage{
				storage: mockStorage,
			}

			ctx := context.Background()
			value, err := ms.UpdateGauge(ctx, tt.metricName, tt.value)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestMetricStorage_GetAllGauges(t *testing.T) {
	tests := []struct {
		name           string
		mockData       map[string]any
		expectedResult map[string]float64
	}{
		{
			name: "get all gauges",
			mockData: map[string]any{
				"gauge:metric1":   42.5,
				"gauge:metric2":   10.1,
				"counter:metric3": 5,
			},
			expectedResult: map[string]float64{
				"metric1": 42.5,
				"metric2": 10.1,
			},
		},
		{
			name:           "empty storage",
			mockData:       map[string]any{},
			expectedResult: map[string]float64{},
		},
		{
			name: "only counters",
			mockData: map[string]any{
				"counter:metric1": 5,
				"counter:metric2": 10,
			},
			expectedResult: map[string]float64{},
		},
		{
			name: "with invalid gauge value",
			mockData: map[string]any{
				"gauge:metric1":     42.5,
				"gauge:invalidType": "not a number",
			},
			expectedResult: map[string]float64{
				"metric1": 42.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			mockStorage.EXPECT().
				GetAll(gomock.Any()).
				Return(tt.mockData, nil)

			ms := &MetricStorage{
				storage: mockStorage,
			}

			ctx := context.Background()
			result, err := ms.GetAllGauges(ctx)

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expectedResult), len(result))
			for k, v := range tt.expectedResult {
				resultValue, exists := result[k]
				assert.True(t, exists)
				assert.Equal(t, v, resultValue)
			}
		})
	}
}

func TestMetricStorage_GetCounter(t *testing.T) {
	tests := []struct {
		name          string
		metricName    string
		mockValue     any
		mockFound     bool
		expectedValue int64
		expectedFound bool
	}{
		{
			name:          "get existing counter metric",
			metricName:    "test",
			mockValue:     int64(42),
			mockFound:     true,
			expectedValue: 42,
			expectedFound: true,
		},
		{
			name:          "get non-existing counter metric",
			metricName:    "not-exist",
			mockValue:     nil,
			mockFound:     false,
			expectedValue: 0,
			expectedFound: false,
		},
		{
			name:          "get with conversion error",
			metricName:    "test",
			mockValue:     "not a number",
			mockFound:     true,
			expectedValue: 0,
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			mockStorage.EXPECT().
				Get(gomock.Any(), addPrefix(tt.metricName, CounterPrefix)).
				Return(tt.mockValue, tt.mockFound)

			ms := &MetricStorage{
				storage: mockStorage,
			}

			ctx := context.Background()
			value, found := ms.GetCounter(ctx, tt.metricName)

			assert.Equal(t, tt.expectedFound, found)
			if tt.expectedFound {
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestMetricStorage_UpdateCounter(t *testing.T) {
	tests := []struct {
		name          string
		metricName    string
		value         int64
		mockGetValue  any
		mockGetFound  bool
		mockSetReturn any
		mockSetError  error
		expectedValue int64
		expectedError bool
	}{
		{
			name:          "update counter with existing value",
			metricName:    "test",
			value:         5,
			mockGetValue:  int64(10),
			mockGetFound:  true,
			mockSetReturn: int64(15),
			mockSetError:  nil,
			expectedValue: 15,
			expectedError: false,
		},
		{
			name:          "update counter with non-existing value",
			metricName:    "test",
			value:         5,
			mockGetValue:  nil,
			mockGetFound:  false,
			mockSetReturn: int64(5),
			mockSetError:  nil,
			expectedValue: 5,
			expectedError: false,
		},
		{
			name:          "update counter with storage error",
			metricName:    "test",
			value:         5,
			mockGetValue:  int64(10),
			mockGetFound:  true,
			mockSetReturn: nil,
			mockSetError:  errors.New("storage error"),
			expectedValue: 0,
			expectedError: true,
		},
		{
			name:          "update counter with conversion error",
			metricName:    "test",
			value:         5,
			mockGetValue:  int64(10),
			mockGetFound:  true,
			mockSetReturn: "not a number",
			mockSetError:  nil,
			expectedValue: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			mockStorage.EXPECT().
				Get(gomock.Any(), addPrefix(tt.metricName, CounterPrefix)).
				Return(tt.mockGetValue, tt.mockGetFound)

			valueToSet := tt.value
			if tt.mockGetFound {
				if current, ok := tt.mockGetValue.(int64); ok {
					valueToSet += current
				}
			}

			mockStorage.EXPECT().
				Set(gomock.Any(), addPrefix(tt.metricName, CounterPrefix), valueToSet).
				Return(tt.mockSetReturn, tt.mockSetError)

			ms := &MetricStorage{
				storage: mockStorage,
			}

			ctx := context.Background()
			value, err := ms.UpdateCounter(ctx, tt.metricName, tt.value)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestMetricStorage_GetAllCounters(t *testing.T) {
	tests := []struct {
		name           string
		mockData       map[string]any
		expectedResult map[string]int64
	}{
		{
			name: "get all counters",
			mockData: map[string]any{
				"counter:metric1": int64(42),
				"counter:metric2": int64(10),
				"gauge:metric3":   42.5,
			},
			expectedResult: map[string]int64{
				"metric1": 42,
				"metric2": 10,
			},
		},
		{
			name:           "empty storage",
			mockData:       map[string]any{},
			expectedResult: map[string]int64{},
		},
		{
			name: "only gauges",
			mockData: map[string]any{
				"gauge:metric1": 42.5,
				"gauge:metric2": 10.1,
			},
			expectedResult: map[string]int64{},
		},
		{
			name: "with invalid counter value",
			mockData: map[string]any{
				"counter:metric1":     int64(42),
				"counter:invalidType": "not a number",
			},
			expectedResult: map[string]int64{
				"metric1": 42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			mockStorage.EXPECT().
				GetAll(gomock.Any()).
				Return(tt.mockData, nil)

			ms := &MetricStorage{
				storage: mockStorage,
			}

			ctx := context.Background()
			result, err := ms.GetAllCounters(ctx)

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expectedResult), len(result))
			for k, v := range tt.expectedResult {
				resultValue, exists := result[k]
				assert.True(t, exists)
				assert.Equal(t, v, resultValue)
			}
		})
	}
}
