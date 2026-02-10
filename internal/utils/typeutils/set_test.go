package typeutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonEmptySetOrDefault(t *testing.T) {
	ctx := context.Background()

	t.Run("empty slice with null original returns null", func(t *testing.T) {
		original := types.SetNull(types.StringType)
		result, diags := NonEmptySetOrDefault(ctx, original, types.StringType, []string{})
		require.False(t, diags.HasError())
		assert.True(t, result.IsNull())
	})

	t.Run("empty slice with unknown original returns unknown", func(t *testing.T) {
		original := types.SetUnknown(types.StringType)
		result, diags := NonEmptySetOrDefault(ctx, original, types.StringType, []string{})
		require.False(t, diags.HasError())
		assert.True(t, result.IsUnknown())
	})

	t.Run("empty slice with known empty set original returns empty set", func(t *testing.T) {
		original, _ := types.SetValueFrom(ctx, types.StringType, []string{})
		result, diags := NonEmptySetOrDefault(ctx, original, types.StringType, []string{})
		require.False(t, diags.HasError())
		assert.False(t, result.IsNull())
		assert.False(t, result.IsUnknown())
		assert.Equal(t, 0, len(result.Elements()))
	})

	t.Run("empty slice with known non-empty set original returns empty set", func(t *testing.T) {
		original, _ := types.SetValueFrom(ctx, types.StringType, []string{"value1"})
		result, diags := NonEmptySetOrDefault(ctx, original, types.StringType, []string{})
		require.False(t, diags.HasError())
		assert.False(t, result.IsNull())
		assert.False(t, result.IsUnknown())
		assert.Equal(t, 0, len(result.Elements()))
	})

	t.Run("non-empty slice returns set with values", func(t *testing.T) {
		original := types.SetNull(types.StringType)
		result, diags := NonEmptySetOrDefault(ctx, original, types.StringType, []string{"value1", "value2"})
		require.False(t, diags.HasError())
		assert.False(t, result.IsNull())
		assert.Equal(t, 2, len(result.Elements()))
	})
}

func TestSetValueFromOptionalComputed(t *testing.T) {
	ctx := context.Background()

	t.Run("empty slice with null original returns null", func(t *testing.T) {
		original := types.SetNull(types.StringType)
		result, diags := SetValueFromOptionalComputed(ctx, original, types.StringType, []string{})
		require.False(t, diags.HasError())
		assert.True(t, result.IsNull())
	})

	t.Run("empty slice with unknown original returns unknown", func(t *testing.T) {
		original := types.SetUnknown(types.StringType)
		result, diags := SetValueFromOptionalComputed(ctx, original, types.StringType, []string{})
		require.False(t, diags.HasError())
		assert.True(t, result.IsUnknown())
	})

	t.Run("empty slice with known empty set original returns empty set", func(t *testing.T) {
		original, _ := types.SetValueFrom(ctx, types.StringType, []string{})
		result, diags := SetValueFromOptionalComputed(ctx, original, types.StringType, []string{})
		require.False(t, diags.HasError())
		assert.False(t, result.IsNull())
		assert.False(t, result.IsUnknown())
		assert.Equal(t, 0, len(result.Elements()))
	})

	t.Run("empty slice with known non-empty set original returns empty set", func(t *testing.T) {
		original, _ := types.SetValueFrom(ctx, types.StringType, []string{"value1"})
		result, diags := SetValueFromOptionalComputed(ctx, original, types.StringType, []string{})
		require.False(t, diags.HasError())
		assert.False(t, result.IsNull())
		assert.False(t, result.IsUnknown())
		assert.Equal(t, 0, len(result.Elements()))
	})

	t.Run("non-empty slice returns set with values regardless of original", func(t *testing.T) {
		original := types.SetNull(types.StringType)
		result, diags := SetValueFromOptionalComputed(ctx, original, types.StringType, []string{"value1", "value2"})
		require.False(t, diags.HasError())
		assert.False(t, result.IsNull())
		assert.Equal(t, 2, len(result.Elements()))
	})

	t.Run("nil slice with known set returns empty set", func(t *testing.T) {
		original, _ := types.SetValueFrom(ctx, types.StringType, []string{"old"})
		result, diags := SetValueFromOptionalComputed(ctx, original, types.StringType, []string(nil))
		require.False(t, diags.HasError())
		assert.False(t, result.IsNull())
		assert.Equal(t, 0, len(result.Elements()))
	})
}
