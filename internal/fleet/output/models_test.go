package output

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_outputModel_populateFromAPI(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		unionFn       func(*kbapi.OutputUnion) error
		wantType      string
		wantName      string
		wantHosts     []string
		wantID        string
		wantErr       bool
		errorContains string
	}{
		{
			name: "elasticsearch",
			unionFn: func(u *kbapi.OutputUnion) error {
				id := "output-elasticsearch"
				return u.FromOutputElasticsearch(kbapi.OutputElasticsearch{
					Name:  "Test Elasticsearch",
					Hosts: []string{"https://example:9200"},
					Id:    &id,
				})
			},
			wantType:  "elasticsearch",
			wantName:  "Test Elasticsearch",
			wantHosts: []string{"https://example:9200"},
			wantID:    "output-elasticsearch",
		},
		{
			name: "logstash",
			unionFn: func(u *kbapi.OutputUnion) error {
				id := "output-logstash"
				return u.FromOutputLogstash(kbapi.OutputLogstash{
					Name:  "Test Logstash",
					Hosts: []string{"logstash:5044"},
					Id:    &id,
				})
			},
			wantType:  "logstash",
			wantName:  "Test Logstash",
			wantHosts: []string{"logstash:5044"},
			wantID:    "output-logstash",
		},
		{
			name: "kafka",
			unionFn: func(u *kbapi.OutputUnion) error {
				id := "output-kafka"
				return u.FromOutputKafka(kbapi.OutputKafka{
					Name:     "Test Kafka",
					Hosts:    []string{"kafka:9092"},
					Id:       &id,
					AuthType: kbapi.OutputKafkaAuthTypeNone,
				})
			},
			wantType:  "kafka",
			wantName:  "Test Kafka",
			wantHosts: []string{"kafka:9092"},
			wantID:    "output-kafka",
		},
		{
			name: "unsupported type",
			unionFn: func(u *kbapi.OutputUnion) error {
				id := "output-remote"
				return u.FromOutputRemoteElasticsearch(kbapi.OutputRemoteElasticsearch{
					Name:  "Remote Elasticsearch",
					Hosts: []string{"https://remote:9200"},
					Id:    &id,
				})
			},
			wantErr:       true,
			errorContains: "unhandled output type",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var union kbapi.OutputUnion
			require.NoError(t, tt.unionFn(&union))

			var model outputModel
			diags := model.populateFromAPI(ctx, &union)
			if tt.wantErr {
				require.True(t, diags.HasError(), "expected error but got none")
				errorSummary := ""
				for _, diag := range diags.Errors() {
					errorSummary += diag.Summary() + " " + diag.Detail()
				}
				assert.Contains(t, errorSummary, tt.errorContains)
				return
			}

			require.False(t, diags.HasError(), "unexpected error: %v", diags.Errors())
			assert.Equal(t, tt.wantType, model.Type.ValueString())
			assert.Equal(t, tt.wantName, model.Name.ValueString())
			assert.Equal(t, tt.wantID, model.OutputID.ValueString())

			var hosts []string
			hostDiags := model.Hosts.ElementsAs(ctx, &hosts, false)
			require.False(t, hostDiags.HasError())
			assert.Equal(t, tt.wantHosts, hosts)
		})
	}
}
