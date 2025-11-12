package agent_policy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequiredVersionsValue_SetSemanticEquals(t *testing.T) {
	ctx := context.Background()
	elemType := getRequiredVersionsElementType()

	// Helper function to create a RequiredVersionsValue
	createValue := func(versions []struct {
		version    string
		percentage int32
	}) RequiredVersionsValue {
		elements := make([]attr.Value, 0, len(versions))
		for _, v := range versions {
			obj, diags := types.ObjectValue(
				map[string]attr.Type{
					"version":    types.StringType,
					"percentage": types.Int32Type,
				},
				map[string]attr.Value{
					"version":    types.StringValue(v.version),
					"percentage": types.Int32Value(v.percentage),
				},
			)
			require.False(t, diags.HasError())
			elements = append(elements, obj)
		}
		val, diags := NewRequiredVersionsValue(elemType, elements)
		require.False(t, diags.HasError())
		return val
	}

	t.Run("equal when same versions regardless of percentage", func(t *testing.T) {
		val1 := createValue([]struct {
			version    string
			percentage int32
		}{
			{"8.15.0", 50},
			{"8.16.0", 50},
		})

		val2 := createValue([]struct {
			version    string
			percentage int32
		}{
			{"8.15.0", 100},
			{"8.16.0", 0},
		})

		equal, diags := val1.SetSemanticEquals(ctx, val2)
		require.False(t, diags.HasError())
		assert.True(t, equal, "Values should be semantically equal when versions match")
	})

	t.Run("not equal when different versions", func(t *testing.T) {
		val1 := createValue([]struct {
			version    string
			percentage int32
		}{
			{"8.15.0", 50},
		})

		val2 := createValue([]struct {
			version    string
			percentage int32
		}{
			{"8.16.0", 50},
		})

		equal, diags := val1.SetSemanticEquals(ctx, val2)
		require.False(t, diags.HasError())
		assert.False(t, equal, "Values should not be equal when versions differ")
	})

	t.Run("not equal when different number of versions", func(t *testing.T) {
		val1 := createValue([]struct {
			version    string
			percentage int32
		}{
			{"8.15.0", 50},
		})

		val2 := createValue([]struct {
			version    string
			percentage int32
		}{
			{"8.15.0", 50},
			{"8.16.0", 50},
		})

		equal, diags := val1.SetSemanticEquals(ctx, val2)
		require.False(t, diags.HasError())
		assert.False(t, equal, "Values should not be equal when number of versions differs")
	})

	t.Run("null values are equal", func(t *testing.T) {
		val1 := NewRequiredVersionsValueNull(elemType)
		val2 := NewRequiredVersionsValueNull(elemType)

		equal, diags := val1.SetSemanticEquals(ctx, val2)
		require.False(t, diags.HasError())
		assert.True(t, equal, "Null values should be equal")
	})

	t.Run("null and non-null are not equal", func(t *testing.T) {
		val1 := NewRequiredVersionsValueNull(elemType)
		val2 := createValue([]struct {
			version    string
			percentage int32
		}{
			{"8.15.0", 50},
		})

		equal, diags := val1.SetSemanticEquals(ctx, val2)
		require.False(t, diags.HasError())
		assert.False(t, equal, "Null and non-null values should not be equal")
	})

	t.Run("equal when versions in different order", func(t *testing.T) {
		val1 := createValue([]struct {
			version    string
			percentage int32
		}{
			{"8.15.0", 50},
			{"8.16.0", 50},
		})

		val2 := createValue([]struct {
			version    string
			percentage int32
		}{
			{"8.16.0", 100},
			{"8.15.0", 0},
		})

		equal, diags := val1.SetSemanticEquals(ctx, val2)
		require.False(t, diags.HasError())
		assert.True(t, equal, "Values should be semantically equal when versions match in different order")
	})
}

func TestConvertRequiredVersions(t *testing.T) {
	ctx := context.Background()
	elemType := getRequiredVersionsElementType()

	t.Run("converts valid required versions", func(t *testing.T) {
		elements := []attr.Value{
			types.ObjectValueMust(
				map[string]attr.Type{
					"version":    types.StringType,
					"percentage": types.Int32Type,
				},
				map[string]attr.Value{
					"version":    types.StringValue("8.15.0"),
					"percentage": types.Int32Value(50),
				},
			),
			types.ObjectValueMust(
				map[string]attr.Type{
					"version":    types.StringType,
					"percentage": types.Int32Type,
				},
				map[string]attr.Value{
					"version":    types.StringValue("8.16.0"),
					"percentage": types.Int32Value(50),
				},
			),
		}

		reqVersions, diags := NewRequiredVersionsValue(elemType, elements)
		require.False(t, diags.HasError())

		model := &agentPolicyModel{
			RequiredVersions: reqVersions,
		}

		result, diags := model.convertRequiredVersions(ctx)
		require.False(t, diags.HasError())
		require.NotNil(t, result)
		assert.Len(t, *result, 2)

		// Check that both versions are present
		versions := make(map[string]float32)
		for _, rv := range *result {
			versions[rv.Version] = rv.Percentage
		}
		assert.Equal(t, float32(50), versions["8.15.0"])
		assert.Equal(t, float32(50), versions["8.16.0"])
	})

	t.Run("returns nil for null required versions", func(t *testing.T) {
		model := &agentPolicyModel{
			RequiredVersions: NewRequiredVersionsValueNull(elemType),
		}

		result, diags := model.convertRequiredVersions(ctx)
		require.False(t, diags.HasError())
		assert.Nil(t, result)
	})

	t.Run("returns nil for empty required versions", func(t *testing.T) {
		reqVersions, diags := NewRequiredVersionsValue(elemType, []attr.Value{})
		require.False(t, diags.HasError())

		model := &agentPolicyModel{
			RequiredVersions: reqVersions,
		}

		result, diags := model.convertRequiredVersions(ctx)
		require.False(t, diags.HasError())
		assert.Nil(t, result)
	})
}
