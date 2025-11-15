package integration_policy

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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

func TestOutputIdHandling(t *testing.T) {
	t.Run("populateFromAPI", func(t *testing.T) {
		model := &integrationPolicyModel{}
		outputId := "test-output-id"
		data := &kbapi.PackagePolicy{
			Id:      "test-id",
			Name:    "test-policy",
			Enabled: true,
			Package: &struct {
				ExperimentalDataStreamFeatures *[]struct {
					DataStream string `json:"data_stream"`
					Features   struct {
						DocValueOnlyNumeric *bool `json:"doc_value_only_numeric,omitempty"`
						DocValueOnlyOther   *bool `json:"doc_value_only_other,omitempty"`
						SyntheticSource     *bool `json:"synthetic_source,omitempty"`
						Tsdb                *bool `json:"tsdb,omitempty"`
					} `json:"features"`
				} `json:"experimental_data_stream_features,omitempty"`
				FipsCompatible *bool   `json:"fips_compatible,omitempty"`
				Name           string  `json:"name"`
				RequiresRoot   *bool   `json:"requires_root,omitempty"`
				Title          *string `json:"title,omitempty"`
				Version        string  `json:"version"`
			}{
				Name:    "test-integration",
				Version: "1.0.0",
			},
			OutputId: &outputId,
		}

		diags := model.populateFromAPI(context.Background(), data)
		require.Empty(t, diags)
		require.Equal(t, "test-output-id", model.OutputID.ValueString())
	})

	t.Run("toAPIModel", func(t *testing.T) {
		model := integrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			OutputID:           types.StringValue("test-output-id"),
		}

		feat := features{
			SupportsPolicyIds: true,
			SupportsOutputId:  true,
		}

		result, diags := model.toAPIModel(context.Background(), false, feat)
		require.Empty(t, diags)
		require.NotNil(t, result.OutputId)
		require.Equal(t, "test-output-id", *result.OutputId)
	})

	t.Run("toAPIModel_unsupported_version", func(t *testing.T) {
		model := integrationPolicyModel{
			Name:               types.StringValue("test-policy"),
			IntegrationName:    types.StringValue("test-integration"),
			IntegrationVersion: types.StringValue("1.0.0"),
			OutputID:           types.StringValue("test-output-id"),
		}

		feat := features{
			SupportsPolicyIds: true,
			SupportsOutputId:  false, // Simulate unsupported version
		}

		_, diags := model.toAPIModel(context.Background(), false, feat)
		require.Len(t, diags, 1)
		require.Equal(t, "Unsupported Elasticsearch version", diags[0].Summary())
		require.Contains(t, diags[0].Detail(), "Output ID is only supported in Elastic Stack")
	})
}
