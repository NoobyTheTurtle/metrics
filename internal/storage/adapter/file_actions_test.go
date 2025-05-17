package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMetricStorage_SaveToFile(t *testing.T) {
	tests := []struct {
		name            string
		fileSaverExists bool
		mockError       error
		expectedError   error
	}{
		{
			name:            "successful save",
			fileSaverExists: true,
			mockError:       nil,
			expectedError:   nil,
		},
		{
			name:            "save with error",
			fileSaverExists: true,
			mockError:       errors.New("save error"),
			expectedError:   errors.New("save error"),
		},
		{
			name:            "nil file saver",
			fileSaverExists: false,
			mockError:       nil,
			expectedError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			var mockFileStorage *MockFileStorage

			ms := &MetricStorage{
				storage: mockStorage,
			}

			if tt.fileSaverExists {
				mockFileStorage = NewMockFileStorage(ctrl)
				mockFileStorage.EXPECT().
					SaveToFile(gomock.Any()).
					Return(tt.mockError)
				ms.fileStorage = mockFileStorage
			}

			ctx := context.Background()
			err := ms.SaveToFile(ctx)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricStorage_LoadFromFile(t *testing.T) {
	tests := []struct {
		name            string
		fileSaverExists bool
		mockError       error
		expectedError   error
	}{
		{
			name:            "successful load",
			fileSaverExists: true,
			mockError:       nil,
			expectedError:   nil,
		},
		{
			name:            "load with error",
			fileSaverExists: true,
			mockError:       errors.New("load error"),
			expectedError:   errors.New("load error"),
		},
		{
			name:            "nil file saver",
			fileSaverExists: false,
			mockError:       nil,
			expectedError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := NewMockStorage(ctrl)
			var mockFileStorage *MockFileStorage

			ms := &MetricStorage{
				storage: mockStorage,
			}

			if tt.fileSaverExists {
				mockFileStorage = NewMockFileStorage(ctrl)
				mockFileStorage.EXPECT().
					LoadFromFile(gomock.Any()).
					Return(tt.mockError)
				ms.fileStorage = mockFileStorage
			}

			ctx := context.Background()
			err := ms.LoadFromFile(ctx)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
