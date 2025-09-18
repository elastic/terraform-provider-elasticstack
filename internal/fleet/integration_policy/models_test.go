package integration_policy

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_SortInputs(t *testing.T) {
	t.Run("WithExisting", func(t *testing.T) {
		existing := []integrationPolicyInputModel{
			{InputID: types.StringValue("A"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("B"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("C"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("D"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("E"), Enabled: types.BoolValue(true)},
		}

		incoming := []integrationPolicyInputModel{
			{InputID: types.StringValue("G"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("F"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("B"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("E"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("C"), Enabled: types.BoolValue(true)},
		}

		want := []integrationPolicyInputModel{
			{InputID: types.StringValue("B"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("C"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("E"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("G"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("F"), Enabled: types.BoolValue(true)},
		}

		sortInputs(incoming, existing)

		require.Equal(t, want, incoming)
	})

	t.Run("WithEmpty", func(t *testing.T) {
		var existing []integrationPolicyInputModel

		incoming := []integrationPolicyInputModel{
			{InputID: types.StringValue("G"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("F"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("B"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("E"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("C"), Enabled: types.BoolValue(true)},
		}

		want := []integrationPolicyInputModel{
			{InputID: types.StringValue("B"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("C"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("E"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("F"), Enabled: types.BoolValue(true)},
			{InputID: types.StringValue("G"), Enabled: types.BoolValue(true)},
		}

		sortInputs(incoming, existing)

		require.Equal(t, want, incoming)
	})
}

func TestNormalizeVarsJson(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		existing jsontypes.Normalized
		api      jsontypes.Normalized
		want     jsontypes.Normalized
	}{
		{
			name:     "plan defines empty object, but api returns null -> preserve empty object",
			existing: jsontypes.NewNormalizedValue("{}"),
			api:      jsontypes.NewNormalizedNull(),
			want:     jsontypes.NewNormalizedValue("{}"),
		},
		{
			name:     "plan does not define value, api returns null -> preserve null",
			existing: jsontypes.NewNormalizedUnknown(),
			api:      jsontypes.NewNormalizedNull(),
			want:     jsontypes.NewNormalizedNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeVarsJson(tt.existing, tt.api)

			// Compare the states and values
			require.Equal(t, tt.want.IsNull(), result.IsNull(), "IsNull() should match")
			require.Equal(t, tt.want.IsUnknown(), result.IsUnknown(), "IsUnknown() should match")

			if !tt.want.IsNull() && !tt.want.IsUnknown() {
				require.Equal(t, tt.want.ValueString(), result.ValueString(), "ValueString() should match")
			}
		})
	}
}
