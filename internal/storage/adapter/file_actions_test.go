package adapter

import (
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
			var mockFileSaver *MockFileSaver

			ms := &MetricStorage{
				storage: mockStorage,
			}

			if tt.fileSaverExists {
				mockFileSaver = NewMockFileSaver(ctrl)
				mockFileSaver.EXPECT().
					SaveToFile().
					Return(tt.mockError)
				ms.fileSaver = mockFileSaver
			}

			err := ms.SaveToFile()

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
			var mockFileSaver *MockFileSaver

			ms := &MetricStorage{
				storage: mockStorage,
			}

			if tt.fileSaverExists {
				mockFileSaver = NewMockFileSaver(ctrl)
				mockFileSaver.EXPECT().
					LoadFromFile().
					Return(tt.mockError)
				ms.fileSaver = mockFileSaver
			}

			err := ms.LoadFromFile()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
