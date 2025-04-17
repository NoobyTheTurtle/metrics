package storage

import (
	"os"
	"path/filepath"
	"testing"

	"maps"

	"github.com/NoobyTheTurtle/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_SaveToFile(t *testing.T) {
	tests := []struct {
		name            string
		fileStoragePath string
		gauges          map[string]float64
		counters        map[string]int64
		wantErr         bool
	}{
		{
			name:            "empty file path",
			fileStoragePath: "",
			gauges: map[string]float64{
				"gauge1": 1.1,
			},
			counters: map[string]int64{
				"counter1": 1,
			},
		},
		{
			name:            "valid file path",
			fileStoragePath: filepath.Join(t.TempDir(), "metrics.json"),
			gauges: map[string]float64{
				"gauge1": 1.1,
				"gauge2": 2.2,
			},
			counters: map[string]int64{
				"counter1": 1,
				"counter2": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges:          make(map[string]float64),
				counters:        make(map[string]int64),
				fileStoragePath: tt.fileStoragePath,
			}

			maps.Copy(ms.gauges, tt.gauges)
			maps.Copy(ms.counters, tt.counters)

			err := ms.SaveToFile()

			assert.NoError(t, err)

			if tt.fileStoragePath == "" {
				return
			}

			data, err := os.ReadFile(tt.fileStoragePath)
			require.NoError(t, err)

			var snapshot model.MemSnapshot
			err = snapshot.UnmarshalJSON(data)
			require.NoError(t, err)

			assert.Equal(t, tt.gauges, snapshot.Gauges)
			assert.Equal(t, tt.counters, snapshot.Counters)
		})
	}
}

func TestMemStorage_LoadFromFile(t *testing.T) {
	tests := []struct {
		name             string
		fileStoragePath  string
		initialGauges    map[string]float64
		initialCounters  map[string]int64
		expectedGauges   map[string]float64
		expectedCounters map[string]int64
		fileData         *model.MemSnapshot
	}{
		{
			name:             "empty file path",
			fileStoragePath:  "",
			initialGauges:    map[string]float64{},
			initialCounters:  map[string]int64{},
			expectedGauges:   map[string]float64{},
			expectedCounters: map[string]int64{},
		},
		{
			name:            "file does not exist",
			fileStoragePath: filepath.Join(t.TempDir(), "non_existent.json"),
			initialGauges: map[string]float64{
				"init_gauge": 1.0,
			},
			initialCounters: map[string]int64{
				"init_counter": 1,
			},
			expectedGauges: map[string]float64{
				"init_gauge": 1.0,
			},
			expectedCounters: map[string]int64{
				"init_counter": 1,
			},
		},
		{
			name:            "valid file with data",
			fileStoragePath: filepath.Join(t.TempDir(), "valid_data.json"),
			initialGauges: map[string]float64{
				"init_gauge": 1.0,
			},
			initialCounters: map[string]int64{
				"init_counter": 1,
			},
			expectedGauges: map[string]float64{
				"gauge1": 1.1,
				"gauge2": 2.2,
			},
			expectedCounters: map[string]int64{
				"counter1": 1,
				"counter2": 2,
			},
			fileData: &model.MemSnapshot{
				Gauges: map[string]float64{
					"gauge1": 1.1,
					"gauge2": 2.2,
				},
				Counters: map[string]int64{
					"counter1": 1,
					"counter2": 2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges:          make(map[string]float64),
				counters:        make(map[string]int64),
				fileStoragePath: tt.fileStoragePath,
			}

			if tt.fileData != nil {
				data, err := tt.fileData.MarshalJSON()
				require.NoError(t, err)

				err = os.MkdirAll(filepath.Dir(tt.fileStoragePath), 0755)
				require.NoError(t, err)

				err = os.WriteFile(tt.fileStoragePath, data, 0644)
				require.NoError(t, err)
			} else {
				maps.Copy(ms.gauges, tt.initialGauges)
				maps.Copy(ms.counters, tt.initialCounters)
			}

			err := ms.LoadFromFile()

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedGauges, ms.gauges)
			assert.Equal(t, tt.expectedCounters, ms.counters)
		})
	}
}
