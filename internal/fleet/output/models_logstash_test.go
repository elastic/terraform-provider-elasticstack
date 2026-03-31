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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_outputModel_toAPICreateModel_logstash(t *testing.T) {
	t.Parallel()

	model := outputModel{
		Name:                types.StringValue("Test Logstash Output"),
		OutputID:            types.StringValue("test-logstash-output"),
		Type:                types.StringValue("logstash"),
		Hosts:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("logstash:5044")}),
		DefaultIntegrations: types.BoolValue(false),
		DefaultMonitoring:   types.BoolValue(false),
		Ssl:                 types.ObjectNull(getSslAttrTypes()),
	}

	union, diags := model.toAPICreateModel(context.Background(), nil)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	createModel, err := union.AsNewOutputLogstash()
	require.NoError(t, err)
	assert.Equal(t, "logstash", string(createModel.Type))
	assert.Equal(t, "Test Logstash Output", createModel.Name)
	assert.Equal(t, []string{"logstash:5044"}, createModel.Hosts)
	assert.Equal(t, "test-logstash-output", *createModel.Id)
}

func Test_outputModel_toAPIUpdateModel_logstash(t *testing.T) {
	t.Parallel()

	model := outputModel{
		Name:                types.StringValue("Updated Logstash Output"),
		Type:                types.StringValue("logstash"),
		Hosts:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("logstash:5044")}),
		DefaultIntegrations: types.BoolValue(false),
		DefaultMonitoring:   types.BoolValue(false),
		Ssl:                 types.ObjectNull(getSslAttrTypes()),
	}

	union, diags := model.toAPIUpdateModel(context.Background(), nil)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	updateModel, err := union.AsUpdateOutputLogstash()
	require.NoError(t, err)
	require.NotNil(t, updateModel.Type)
	assert.Equal(t, "logstash", string(*updateModel.Type))
	require.NotNil(t, updateModel.Name)
	assert.Equal(t, "Updated Logstash Output", *updateModel.Name)
}

func Test_outputModel_populateFromAPI_logstash(t *testing.T) {
	t.Parallel()

	var union kbapi.OutputUnion
	err := union.FromOutputLogstash(kbapi.OutputLogstash{
		Id:                  new("logstash-output-id"),
		Name:                "Fleet Logstash Output",
		Type:                kbapi.KibanaHTTPAPIsOutputLogstashTypeLogstash,
		Hosts:               []string{"logstash:5044"},
		IsDefault:           new(false),
		IsDefaultMonitoring: new(false),
	})
	require.NoError(t, err)

	model := outputModel{
		SpaceIDs: types.SetNull(types.StringType),
	}
	diags := model.populateFromAPI(context.Background(), &union)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.Equal(t, "logstash-output-id", model.ID.ValueString())
	assert.Equal(t, "logstash-output-id", model.OutputID.ValueString())
	assert.Equal(t, "Fleet Logstash Output", model.Name.ValueString())
	assert.Equal(t, "logstash", model.Type.ValueString())
	assert.Equal(t, []attr.Value{types.StringValue("logstash:5044")}, model.Hosts.Elements())
}

//go:fix inline
func ptrString(v string) *string {
	return new(v)
}

//go:fix inline
func ptrBool(v bool) *bool {
	return new(v)
}
