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

	union, diags := model.toAPIUpdateModel(context.Background(), nil, outputModel{})
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	updateModel, err := union.AsUpdateOutputLogstash()
	require.NoError(t, err)
	require.NotNil(t, updateModel.Type)
	assert.Equal(t, "logstash", string(*updateModel.Type))
	require.NotNil(t, updateModel.Name)
	assert.Equal(t, "Updated Logstash Output", *updateModel.Name)
}

func Test_outputModel_toAPIUpdateModel_logstash_sendsEmptySslToClearFleet(t *testing.T) {
	t.Parallel()

	cas := []string{"placeholder"}
	priorSsl, d := sslToObjectValue(context.Background(), ptrString("placeholder"), &cas, ptrString("placeholder"))
	require.False(t, d.HasError())

	model := outputModel{
		Name:                types.StringValue("Logstash Output (No SSL)"),
		Type:                types.StringValue("logstash"),
		Hosts:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("logstash:5044")}),
		DefaultIntegrations: types.BoolValue(false),
		DefaultMonitoring:   types.BoolValue(false),
		Ssl:                 types.ObjectNull(getSslAttrTypes()),
	}

	prior := outputModel{
		Ssl: priorSsl,
	}

	union, diags := model.toAPIUpdateModel(context.Background(), nil, prior)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	updateModel, err := union.AsUpdateOutputLogstash()
	require.NoError(t, err)
	require.NotNil(t, updateModel.Ssl, "empty ssl object must be sent so Fleet clears stored ssl")
}

func Test_logstashConfigYamlForUpdate(t *testing.T) {
	t.Parallel()

	t.Run("plan set returns plan value", func(t *testing.T) {
		v := logstashConfigYamlForUpdate(types.StringValue("a: 1"), types.StringNull())
		require.NotNil(t, v)
		assert.Equal(t, "a: 1", *v)
	})
	t.Run("plan unset and prior value clears with empty string", func(t *testing.T) {
		v := logstashConfigYamlForUpdate(types.StringNull(), types.StringValue(`"ssl.verification_mode": none`))
		require.NotNil(t, v)
		assert.Empty(t, *v)
	})
	t.Run("plan unset and no prior value omits field", func(t *testing.T) {
		assert.Nil(t, logstashConfigYamlForUpdate(types.StringNull(), types.StringNull()))
	})
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

func Test_outputModel_populateFromAPI_logstash_withoutSSL(t *testing.T) {
	t.Parallel()

	var union kbapi.OutputUnion
	err := union.FromOutputLogstash(kbapi.OutputLogstash{
		Id:                  ptrString("logstash-no-ssl-output-id"),
		Name:                "Fleet Logstash Output No SSL",
		Type:                kbapi.KibanaHTTPAPIsOutputLogstashTypeLogstash,
		Hosts:               []string{"logstash:5044"},
		IsDefault:           ptrBool(false),
		IsDefaultMonitoring: ptrBool(false),
		Ssl:                 nil,
	})
	require.NoError(t, err)

	model := outputModel{
		SpaceIDs: types.SetNull(types.StringType),
	}
	diags := model.populateFromAPI(context.Background(), &union)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	assert.True(t, model.Ssl.IsNull(), "expected ssl to be null when API returns nil ssl")
}

//go:fix inline
func ptrString(v string) *string {
	return new(v)
}

//go:fix inline
func ptrBool(v bool) *bool {
	return new(v)
}
