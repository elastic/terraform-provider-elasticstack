package customtypes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/stretchr/testify/require"
)

func TestMemorySizeType_String(t *testing.T) {
	require.Equal(t, "customtypes.MemorySizeType", MemorySizeType{}.String())
}

func TestMemorySizeType_ValueType(t *testing.T) {
	require.Equal(t, MemorySize{}, MemorySizeType{}.ValueType(context.Background()))
}

func TestMemorySizeType_Equal(t *testing.T) {
	tests := []struct {
		name     string
		typ      MemorySizeType
		other    attr.Type
		expected bool
	}{
		{
			name:     "equal to another MemorySizeType",
			typ:      MemorySizeType{},
			other:    MemorySizeType{},
			expected: true,
		},
		{
			name:     "not equal to different type",
			typ:      MemorySizeType{},
			other:    DurationType{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.typ.Equal(tt.other))
		})
	}
}
