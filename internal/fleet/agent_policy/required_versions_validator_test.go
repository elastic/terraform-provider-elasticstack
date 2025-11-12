package agent_policy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestUniqueVersionsValidator(t *testing.T) {
	ctx := context.Background()
	elemType := getRequiredVersionsElementType()

	t.Run("valid - unique versions", func(t *testing.T) {
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

		setValue := types.SetValueMust(elemType, elements)

		req := validator.SetRequest{
			Path:        path.Root("required_versions"),
			ConfigValue: setValue,
		}
		resp := &validator.SetResponse{
			Diagnostics: diag.Diagnostics{},
		}

		UniqueVersions().ValidateSet(ctx, req, resp)

		assert.False(t, resp.Diagnostics.HasError(), "Expected no errors for unique versions")
	})

	t.Run("invalid - duplicate versions", func(t *testing.T) {
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
					"version":    types.StringValue("8.15.0"),
					"percentage": types.Int32Value(100),
				},
			),
		}

		setValue := types.SetValueMust(elemType, elements)

		req := validator.SetRequest{
			Path:        path.Root("required_versions"),
			ConfigValue: setValue,
		}
		resp := &validator.SetResponse{
			Diagnostics: diag.Diagnostics{},
		}

		UniqueVersions().ValidateSet(ctx, req, resp)

		assert.True(t, resp.Diagnostics.HasError(), "Expected error for duplicate versions")
		assert.Contains(t, resp.Diagnostics.Errors()[0].Detail(), "8.15.0")
	})

	t.Run("valid - null value", func(t *testing.T) {
		setValue := types.SetNull(elemType)

		req := validator.SetRequest{
			Path:        path.Root("required_versions"),
			ConfigValue: setValue,
		}
		resp := &validator.SetResponse{
			Diagnostics: diag.Diagnostics{},
		}

		UniqueVersions().ValidateSet(ctx, req, resp)

		assert.False(t, resp.Diagnostics.HasError(), "Expected no errors for null value")
	})

	t.Run("valid - empty set", func(t *testing.T) {
		setValue := types.SetValueMust(elemType, []attr.Value{})

		req := validator.SetRequest{
			Path:        path.Root("required_versions"),
			ConfigValue: setValue,
		}
		resp := &validator.SetResponse{
			Diagnostics: diag.Diagnostics{},
		}

		UniqueVersions().ValidateSet(ctx, req, resp)

		assert.False(t, resp.Diagnostics.HasError(), "Expected no errors for empty set")
	})
}
