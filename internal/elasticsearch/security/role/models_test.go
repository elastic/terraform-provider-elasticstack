package role

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_slicesEqualIgnoreOrder(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "empty slices are equal",
			a:    []string{},
			b:    []string{},
			want: true,
		},
		{
			name: "identical slices are equal",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "c"},
			want: true,
		},
		{
			name: "different order but same elements",
			a:    []string{"a", "b", "c"},
			b:    []string{"c", "a", "b"},
			want: true,
		},
		{
			name: "different lengths are not equal",
			a:    []string{"a", "b"},
			b:    []string{"a", "b", "c"},
			want: false,
		},
		{
			name: "different elements are not equal",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "d"},
			want: false,
		},
		{
			name: "duplicate elements handled correctly",
			a:    []string{"a", "a", "b"},
			b:    []string{"a", "b", "b"},
			want: false,
		},
		{
			name: "duplicate elements same in both",
			a:    []string{"a", "a", "b"},
			b:    []string{"b", "a", "a"},
			want: true,
		},
		{
			name: "nil vs empty slice",
			a:    nil,
			b:    []string{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slicesEqualIgnoreOrder(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
