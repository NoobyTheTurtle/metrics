package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateSHA256(t *testing.T) {
	type args struct {
		data []byte
		key  string
	}
	tests := []struct {
		name string
		args args
		want string
		err  error
	}{
		{
			name: "Simple test with key",
			args: args{
				data: []byte("hello world"),
				key:  "secret_key",
			},
			want: "cf1a418afaafc798df48fd804a2abf6970283afd8c40b41f818ad9b6ca4f8ca8",
			err:  nil,
		},
		{
			name: "Empty data with key",
			args: args{
				data: []byte(""),
				key:  "secret_key",
			},
			want: "f304c11274cabc93abe79f3abb848b94026d3e1c9a49071ea96fbece9b9f2bb0",
			err:  nil,
		},
		{
			name: "Test with empty key",
			args: args{
				data: []byte("hello world"),
				key:  "",
			},
			want: "",
			err:  ErrKeyEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateSHA256(tt.args.data, tt.args.key)

			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
