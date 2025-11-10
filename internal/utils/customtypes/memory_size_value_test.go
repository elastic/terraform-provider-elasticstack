package customtypes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestMemorySize_Type(t *testing.T) {
	require.Equal(t, MemorySizeType{}, MemorySize{}.Type(context.Background()))
}

func TestMemorySize_Equal(t *testing.T) {
	tests := []struct {
		name          string
		expectedEqual bool
		val           MemorySize
		otherVal      attr.Value
	}{
		{
			name:          "not equal if the other value is not a memory size",
			expectedEqual: false,
			val:           NewMemorySizeValue("128mb"),
			otherVal:      basetypes.NewBoolValue(true),
		},
		{
			name:          "not equal if the memory sizes are not equal",
			expectedEqual: false,
			val:           NewMemorySizeValue("128mb"),
			otherVal:      NewMemorySizeValue("256mb"),
		},
		{
			name:          "not equal if the memory sizes are semantically equal but string values are not equal",
			expectedEqual: false,
			val:           NewMemorySizeValue("1gb"),
			otherVal:      NewMemorySizeValue("1024mb"),
		},
		{
			name:          "equal if the memory size string values are equal",
			expectedEqual: true,
			val:           NewMemorySizeValue("128mb"),
			otherVal:      NewMemorySizeValue("128mb"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expectedEqual, tt.val.Equal(tt.otherVal))
		})
	}
}

func TestMemorySize_ValidateAttribute(t *testing.T) {
	tests := []struct {
		name          string
		memorySize    MemorySize
		expectedDiags diag.Diagnostics
	}{
		{
			name:       "unknown is valid",
			memorySize: NewMemorySizeNull(),
		},
		{
			name:       "null is valid",
			memorySize: NewMemorySizeUnknown(),
		},
		{
			name:       "valid memory sizes are valid - bytes",
			memorySize: NewMemorySizeValue("1024"),
		},
		{
			name:       "valid memory sizes are valid - kilobytes",
			memorySize: NewMemorySizeValue("128k"),
		},
		{
			name:       "valid memory sizes are valid - kilobytes with B",
			memorySize: NewMemorySizeValue("128kb"),
		},
		{
			name:       "valid memory sizes are valid - megabytes",
			memorySize: NewMemorySizeValue("128m"),
		},
		{
			name:       "valid memory sizes are valid - megabytes with B",
			memorySize: NewMemorySizeValue("128mb"),
		},
		{
			name:       "valid memory sizes are valid - uppercase megabytes",
			memorySize: NewMemorySizeValue("128MB"),
		},
		{
			name:       "valid memory sizes are valid - gigabytes",
			memorySize: NewMemorySizeValue("2g"),
		},
		{
			name:       "valid memory sizes are valid - gigabytes with B",
			memorySize: NewMemorySizeValue("2gb"),
		},
		{
			name:       "valid memory sizes are valid - terabytes",
			memorySize: NewMemorySizeValue("1t"),
		},
		{
			name:       "non-memory strings are invalid",
			memorySize: NewMemorySizeValue("not a memory size"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("memory_size"),
					"Invalid memory size string value",
					"A string value was provided that is not a valid memory size format\n\nGiven value \"not a memory size\"\nExpected format: number followed by optional unit (k/K, m/M, g/G, t/T) and optional 'b/B' suffix",
				),
			},
		},
		{
			name:       "negative numbers are invalid",
			memorySize: NewMemorySizeValue("-128mb"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("memory_size"),
					"Invalid memory size string value",
					"A string value was provided that is not a valid memory size format\n\nGiven value \"-128mb\"\nExpected format: number followed by optional unit (k/K, m/M, g/G, t/T) and optional 'b/B' suffix",
				),
			},
		},
		{
			name:       "float numbers are invalid",
			memorySize: NewMemorySizeValue("128.5mb"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("memory_size"),
					"Invalid memory size string value",
					"A string value was provided that is not a valid memory size format\n\nGiven value \"128.5mb\"\nExpected format: number followed by optional unit (k/K, m/M, g/G, t/T) and optional 'b/B' suffix",
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := xattr.ValidateAttributeResponse{}

			tt.memorySize.ValidateAttribute(
				context.Background(),
				xattr.ValidateAttributeRequest{
					Path: path.Root("memory_size"),
				},
				&resp,
			)

			if tt.expectedDiags == nil {
				require.Nil(t, resp.Diagnostics)
			} else {
				require.Equal(t, tt.expectedDiags, resp.Diagnostics)
			}
		})
	}
}

func TestMemorySize_StringSemanticEquals(t *testing.T) {
	tests := []struct {
		name               string
		memorySize         MemorySize
		otherVal           basetypes.StringValuable
		expectedEqual      bool
		expectedErrorDiags bool
	}{
		{
			name:               "should error if the other value is not a memory size",
			memorySize:         NewMemorySizeValue("128mb"),
			otherVal:           basetypes.NewStringValue("128mb"),
			expectedEqual:      false,
			expectedErrorDiags: true,
		},
		{
			name:          "two null values are semantically equal",
			memorySize:    NewMemorySizeNull(),
			otherVal:      NewMemorySizeNull(),
			expectedEqual: true,
		},
		{
			name:          "null is not equal to unknown",
			memorySize:    NewMemorySizeNull(),
			otherVal:      NewMemorySizeUnknown(),
			expectedEqual: false,
		},
		{
			name:          "null is not equal to a string value",
			memorySize:    NewMemorySizeNull(),
			otherVal:      NewMemorySizeValue("128mb"),
			expectedEqual: false,
		},
		{
			name:          "two unknown values are semantically equal",
			memorySize:    NewMemorySizeUnknown(),
			otherVal:      NewMemorySizeUnknown(),
			expectedEqual: true,
		},
		{
			name:          "unknown is not equal to a string value",
			memorySize:    NewMemorySizeUnknown(),
			otherVal:      NewMemorySizeValue("128mb"),
			expectedEqual: false,
		},
		{
			name:          "two equal values are semantically equal",
			memorySize:    NewMemorySizeValue("128mb"),
			otherVal:      NewMemorySizeValue("128mb"),
			expectedEqual: true,
		},
		{
			name:          "two semantically equal values - gb to mb",
			memorySize:    NewMemorySizeValue("2g"),
			otherVal:      NewMemorySizeValue("2048m"),
			expectedEqual: true,
		},
		{
			name:          "two semantically equal values - gb with B to mb",
			memorySize:    NewMemorySizeValue("2gb"),
			otherVal:      NewMemorySizeValue("2048mb"),
			expectedEqual: true,
		},
		{
			name:          "two semantically equal values - different case",
			memorySize:    NewMemorySizeValue("128MB"),
			otherVal:      NewMemorySizeValue("128mb"),
			expectedEqual: true,
		},
		{
			name:          "two semantically equal values - kb to bytes (rounded to MB)",
			memorySize:    NewMemorySizeValue("2048k"),
			otherVal:      NewMemorySizeValue("2097152"),
			expectedEqual: true,
		},
		{
			name:          "bytes that don't round to same MB are not equal",
			memorySize:    NewMemorySizeValue("1048576"), // exactly 1MB
			otherVal:      NewMemorySizeValue("1048575"), // 1 byte less, rounds to 0MB
			expectedEqual: false,
		},
		{
			name:          "partial MB values round down to same MB",
			memorySize:    NewMemorySizeValue("1500000"), // ~1.43MB, rounds to 1MB
			otherVal:      NewMemorySizeValue("1048576"), // exactly 1MB
			expectedEqual: true,
		},
		{
			name:               "errors if this value is invalid",
			memorySize:         NewMemorySizeValue("not a memory size"),
			otherVal:           NewMemorySizeValue("128mb"),
			expectedEqual:      false,
			expectedErrorDiags: true,
		},
		{
			name:               "errors if the other value is invalid",
			memorySize:         NewMemorySizeValue("128mb"),
			otherVal:           NewMemorySizeValue("not a memory size"),
			expectedEqual:      false,
			expectedErrorDiags: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEqual, diags := tt.memorySize.StringSemanticEquals(context.Background(), tt.otherVal)

			require.Equal(t, tt.expectedEqual, isEqual)
			require.Equal(t, tt.expectedErrorDiags, diags.HasError())
		})
	}
}

func TestMemorySize_ParseBytes(t *testing.T) {
	tests := []struct {
		name          string
		memorySize    MemorySize
		expectedBytes int64
		expectedError bool
	}{
		{
			name:          "null value should error",
			memorySize:    NewMemorySizeNull(),
			expectedError: true,
		},
		{
			name:          "unknown value should error",
			memorySize:    NewMemorySizeUnknown(),
			expectedError: true,
		},
		{
			name:          "bytes without unit",
			memorySize:    NewMemorySizeValue("1048576"), // exactly 1MB
			expectedBytes: 1048576,
		},
		{
			name:          "bytes without unit - rounds down",
			memorySize:    NewMemorySizeValue("1048575"), // 1 byte less than 1MB
			expectedBytes: 0,                             // rounds down to 0MB
		},
		{
			name:          "bytes without unit - partial MB rounds down",
			memorySize:    NewMemorySizeValue("1500000"), // ~1.43MB
			expectedBytes: 1048576,                       // rounds down to 1MB
		},
		{
			name:          "kilobytes",
			memorySize:    NewMemorySizeValue("1024k"), // exactly 1MB
			expectedBytes: 1024 * 1024,
		},
		{
			name:          "kilobytes - partial MB rounds down",
			memorySize:    NewMemorySizeValue("1000k"), // ~976KB, rounds down to 0MB
			expectedBytes: 0,
		},
		{
			name:          "kilobytes with B suffix",
			memorySize:    NewMemorySizeValue("1024kb"), // exactly 1MB
			expectedBytes: 1024 * 1024,
		},
		{
			name:          "megabytes",
			memorySize:    NewMemorySizeValue("128m"),
			expectedBytes: 128 * 1024 * 1024,
		},
		{
			name:          "megabytes with B suffix",
			memorySize:    NewMemorySizeValue("128mb"),
			expectedBytes: 128 * 1024 * 1024,
		},
		{
			name:          "uppercase megabytes",
			memorySize:    NewMemorySizeValue("128MB"),
			expectedBytes: 128 * 1024 * 1024,
		},
		{
			name:          "gigabytes",
			memorySize:    NewMemorySizeValue("2g"),
			expectedBytes: 2 * 1024 * 1024 * 1024,
		},
		{
			name:          "gigabytes with B suffix",
			memorySize:    NewMemorySizeValue("2gb"),
			expectedBytes: 2 * 1024 * 1024 * 1024,
		},
		{
			name:          "terabytes",
			memorySize:    NewMemorySizeValue("1t"),
			expectedBytes: 1024 * 1024 * 1024 * 1024,
		},
		{
			name:          "terabytes with B suffix",
			memorySize:    NewMemorySizeValue("1tb"),
			expectedBytes: 1024 * 1024 * 1024 * 1024,
		},
		{
			name:          "invalid format",
			memorySize:    NewMemorySizeValue("not a memory size"),
			expectedError: true,
		},
		{
			name:          "invalid number",
			memorySize:    NewMemorySizeValue("abcmb"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, diags := tt.memorySize.ConvertToMB()

			if tt.expectedError {
				require.True(t, diags.HasError())
			} else {
				require.False(t, diags.HasError())
				require.Equal(t, tt.expectedBytes, bytes)
			}
		})
	}
}
