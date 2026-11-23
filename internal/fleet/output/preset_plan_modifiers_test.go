// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package output

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_presetPlanValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		planType            string
		stateType           types.String
		stateValue          types.String
		configYamlUnchanged bool
		want                types.String
	}{
		{
			name:                "logstash plans null",
			planType:            outputTypeLogstash,
			stateType:           types.StringValue(outputTypeLogstash),
			stateValue:          types.StringNull(),
			configYamlUnchanged: true,
			want:                types.StringNull(),
		},
		{
			name:                "kafka plans null after type change from elasticsearch",
			planType:            outputTypeKafka,
			stateType:           types.StringValue(outputTypeElasticsearch),
			stateValue:          types.StringValue("scale"),
			configYamlUnchanged: true,
			want:                types.StringNull(),
		},
		{
			name:                "elasticsearch create is unknown",
			planType:            outputTypeElasticsearch,
			stateType:           types.StringNull(),
			stateValue:          types.StringNull(),
			configYamlUnchanged: false,
			want:                types.StringUnknown(),
		},
		{
			name:                "remote_elasticsearch create is unknown",
			planType:            outputTypeRemoteElasticsearch,
			stateType:           types.StringNull(),
			stateValue:          types.StringNull(),
			configYamlUnchanged: false,
			want:                types.StringUnknown(),
		},
		{
			name:                "elasticsearch keeps stored preset when config_yaml unchanged",
			planType:            outputTypeElasticsearch,
			stateType:           types.StringValue(outputTypeElasticsearch),
			stateValue:          types.StringValue("throughput"),
			configYamlUnchanged: true,
			want:                types.StringValue("throughput"),
		},
		{
			name:                "remote_elasticsearch keeps stored preset when config_yaml unchanged",
			planType:            outputTypeRemoteElasticsearch,
			stateType:           types.StringValue(outputTypeRemoteElasticsearch),
			stateValue:          types.StringValue("scale"),
			configYamlUnchanged: true,
			want:                types.StringValue("scale"),
		},
		{
			name:                "elasticsearch is unknown when config_yaml changes",
			planType:            outputTypeElasticsearch,
			stateType:           types.StringValue(outputTypeElasticsearch),
			stateValue:          types.StringValue("balanced"),
			configYamlUnchanged: false,
			want:                types.StringUnknown(),
		},
		{
			name:                "remote_elasticsearch is unknown when config_yaml changes",
			planType:            outputTypeRemoteElasticsearch,
			stateType:           types.StringValue(outputTypeRemoteElasticsearch),
			stateValue:          types.StringValue("scale"),
			configYamlUnchanged: false,
			want:                types.StringUnknown(),
		},
		{
			name:                "elasticsearch is unknown after type change from remote_elasticsearch",
			planType:            outputTypeElasticsearch,
			stateType:           types.StringValue(outputTypeRemoteElasticsearch),
			stateValue:          types.StringValue("scale"),
			configYamlUnchanged: true,
			want:                types.StringUnknown(),
		},
		{
			name:                "elasticsearch is unknown when no preset is stored",
			planType:            outputTypeElasticsearch,
			stateType:           types.StringValue(outputTypeElasticsearch),
			stateValue:          types.StringNull(),
			configYamlUnchanged: true,
			want:                types.StringUnknown(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := presetPlanValue(tt.planType, tt.stateType, tt.stateValue, tt.configYamlUnchanged)
			assert.True(t, got.Equal(tt.want), "got %s, want %s", got, tt.want)
		})
	}
}

func Test_presetPlanModifier_description(t *testing.T) {
	t.Parallel()

	modifier := presetPlanModifier()
	assert.Equal(t, modifier.Description(context.Background()), modifier.MarkdownDescription(context.Background()))
}
