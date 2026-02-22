package tfsdkutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlattenMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  map[string]any
		out map[string]any
	}{
		{
			map[string]any{"key1": map[string]any{"key2": 1}},
			map[string]any{"key1.key2": 1},
		},
		{
			map[string]any{"key1": map[string]any{"key2": map[string]any{"key3": 1}}},
			map[string]any{"key1.key2.key3": 1},
		},
		{
			map[string]any{"key1": 1},
			map[string]any{"key1": 1},
		},
		{
			map[string]any{"key1": "test"},
			map[string]any{"key1": "test"},
		},
		{
			map[string]any{"key1": map[string]any{"key2": 1, "key3": "test", "key4": []int{1, 2, 3}}},
			map[string]any{"key1.key2": 1, "key1.key3": "test", "key1.key4": []int{1, 2, 3}},
		},
	}

	for _, tc := range tests {
		res := flattenMap(tc.in)
		require.Equal(t, tc.out, res)
	}
}
