package integration_policy

import (
	"testing"

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
