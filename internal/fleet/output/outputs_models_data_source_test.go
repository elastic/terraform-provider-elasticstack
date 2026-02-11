package output

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_outputsDataSourceModel_populateFromAPI_filters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	outputs := []kbapi.OutputUnion{
		outputUnionElasticsearch(t, "es-1", true, false, utils.Pointer("")),
		outputUnionLogstash(t, "ls-1", false, true),
		outputUnionKafka(t, "kafka-1", false, false),
	}

	tests := []struct {
		name                     string
		model                    outputsDataSourceModel
		wantIDs                  []string
		assertTrustedFingerprint bool
	}{
		{
			name: "no filters returns all outputs",
			model: outputsDataSourceModel{
				OutputID:            types.StringNull(),
				Type:                types.StringNull(),
				DefaultIntegrations: types.BoolNull(),
				DefaultMonitoring:   types.BoolNull(),
			},
			wantIDs:                  []string{"es-1", "ls-1", "kafka-1"},
			assertTrustedFingerprint: true,
		},
		{
			name: "type filter returns only matching outputs",
			model: outputsDataSourceModel{
				OutputID:            types.StringNull(),
				Type:                types.StringValue("kafka"),
				DefaultIntegrations: types.BoolNull(),
				DefaultMonitoring:   types.BoolNull(),
			},
			wantIDs: []string{"kafka-1"},
		},
		{
			name: "output_id filter returns only matching output",
			model: outputsDataSourceModel{
				OutputID:            types.StringValue("ls-1"),
				Type:                types.StringNull(),
				DefaultIntegrations: types.BoolNull(),
				DefaultMonitoring:   types.BoolNull(),
			},
			wantIDs: []string{"ls-1"},
		},
		{
			name: "default_integrations filter matches boolean",
			model: outputsDataSourceModel{
				OutputID:            types.StringNull(),
				Type:                types.StringNull(),
				DefaultIntegrations: types.BoolValue(false),
				DefaultMonitoring:   types.BoolNull(),
			},
			wantIDs: []string{"ls-1", "kafka-1"},
		},
		{
			name: "combined filters return only matches",
			model: outputsDataSourceModel{
				OutputID:            types.StringNull(),
				Type:                types.StringValue("logstash"),
				DefaultIntegrations: types.BoolNull(),
				DefaultMonitoring:   types.BoolValue(true),
			},
			wantIDs: []string{"ls-1"},
		},
		{
			name: "output_id filter can return empty list",
			model: outputsDataSourceModel{
				OutputID:            types.StringValue("missing"),
				Type:                types.StringNull(),
				DefaultIntegrations: types.BoolNull(),
				DefaultMonitoring:   types.BoolNull(),
			},
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			model := tt.model
			diags := model.populateFromAPI(ctx, outputs)
			require.False(t, diags.HasError(), "unexpected error: %v", diags.Errors())

			items := outputItemsFromModel(t, ctx, model)
			ids := make([]string, 0, len(items))
			for _, item := range items {
				ids = append(ids, item.ID.ValueString())
			}
			assert.ElementsMatch(t, tt.wantIDs, ids)

			if tt.assertTrustedFingerprint {
				item := outputItemByID(t, items, "es-1")
				assert.True(t, item.CaTrustedFingerprint.IsNull())
			}
		})
	}
}

func Test_outputsDataSourceModel_populateFromAPI_unsupportedType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := outputsDataSourceModel{
		OutputID:            types.StringNull(),
		Type:                types.StringNull(),
		DefaultIntegrations: types.BoolNull(),
		DefaultMonitoring:   types.BoolNull(),
	}

	outputs := []kbapi.OutputUnion{outputUnionRemoteElasticsearch(t, "remote-1")}
	diags := model.populateFromAPI(ctx, outputs)

	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Summary(), "unhandled output type")
}

func outputItemsFromModel(t *testing.T, ctx context.Context, model outputsDataSourceModel) []outputItemModel {
	t.Helper()

	var diags diag.Diagnostics
	items := utils.ListTypeAs[outputItemModel](ctx, model.Items, path.Root("items"), &diags)
	require.False(t, diags.HasError(), "unexpected error: %v", diags.Errors())
	return items
}

func outputItemByID(t *testing.T, items []outputItemModel, id string) outputItemModel {
	t.Helper()

	for _, item := range items {
		if item.ID.ValueString() == id {
			return item
		}
	}

	require.FailNow(t, "missing item", "output %q not found", id)
	return outputItemModel{}
}

func outputUnionElasticsearch(t *testing.T, id string, isDefault bool, isDefaultMonitoring bool, caTrustedFingerprint *string) kbapi.OutputUnion {
	t.Helper()

	union := kbapi.OutputUnion{}
	err := union.FromOutputElasticsearch(kbapi.OutputElasticsearch{
		Id:                   &id,
		Name:                 "Elasticsearch " + id,
		Hosts:                []string{"https://example:9200"},
		IsDefault:            &isDefault,
		IsDefaultMonitoring:  &isDefaultMonitoring,
		CaTrustedFingerprint: caTrustedFingerprint,
		Type:                 kbapi.OutputElasticsearchTypeElasticsearch,
	})
	require.NoError(t, err)
	return union
}

func outputUnionLogstash(t *testing.T, id string, isDefault bool, isDefaultMonitoring bool) kbapi.OutputUnion {
	t.Helper()

	union := kbapi.OutputUnion{}
	err := union.FromOutputLogstash(kbapi.OutputLogstash{
		Id:                  &id,
		Name:                "Logstash " + id,
		Hosts:               []string{"logstash:5044"},
		IsDefault:           &isDefault,
		IsDefaultMonitoring: &isDefaultMonitoring,
		Type:                kbapi.OutputLogstashTypeLogstash,
	})
	require.NoError(t, err)
	return union
}

func outputUnionKafka(t *testing.T, id string, isDefault bool, isDefaultMonitoring bool) kbapi.OutputUnion {
	t.Helper()

	union := kbapi.OutputUnion{}
	err := union.FromOutputKafka(kbapi.OutputKafka{
		Id:                  &id,
		Name:                "Kafka " + id,
		Hosts:               []string{"kafka:9092"},
		IsDefault:           &isDefault,
		IsDefaultMonitoring: &isDefaultMonitoring,
		AuthType:            kbapi.OutputKafkaAuthTypeNone,
		Type:                kbapi.OutputKafkaTypeKafka,
	})
	require.NoError(t, err)
	return union
}

func outputUnionRemoteElasticsearch(t *testing.T, id string) kbapi.OutputUnion {
	t.Helper()

	union := kbapi.OutputUnion{}
	err := union.FromOutputRemoteElasticsearch(kbapi.OutputRemoteElasticsearch{
		Id:    &id,
		Name:  "Remote Elasticsearch",
		Hosts: []string{"https://remote:9200"},
		Type:  kbapi.OutputRemoteElasticsearchTypeRemoteElasticsearch,
	})
	require.NoError(t, err)
	return union
}
