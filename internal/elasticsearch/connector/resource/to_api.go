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

package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	getconnector "github.com/elastic/go-elasticsearch/v9/typedapi/connector/get"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/connectorfieldtype"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var configurationValuesPath = path.Root("configuration_values")

const (
	configurationSchemaNotRegisteredTitle = "Connector configuration schema not yet registered"
	configurationSchemaNotRegisteredURL   = "https://www.elastic.co/docs/reference/search-connectors/api-tutorial"
)

func configurationSchemaNotRegisteredDetail(serviceType string) string {
	return fmt.Sprintf(
		"Connector configuration schema has not been registered yet. "+
			"The connector service must boot and write a schema for service_type %q before configuration_values can be applied. "+
			"See %s for setup steps.",
		serviceType,
		configurationSchemaNotRegisteredURL,
	)
}

func (data ContentConnectorData) toCreateConnectorBody() elasticsearch.CreateConnectorBody {
	return elasticsearch.CreateConnectorBody{
		Name:        typeutils.OptionalString(data.Name),
		Description: typeutils.OptionalString(data.Description),
		IndexName:   typeutils.OptionalString(data.IndexName),
		IsNative:    typeutils.OptionalBool(data.IsNative),
		Language:    typeutils.OptionalString(data.Language),
		ServiceType: data.ServiceType.ValueString(),
	}
}

func (data ContentConnectorData) toPipelineAPI(ctx context.Context, diags *diag.Diagnostics) estypes.IngestPipelineParams {
	var model connector.PipelineModel
	diags.Append(data.Pipeline.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return estypes.IngestPipelineParams{}
	}
	return estypes.IngestPipelineParams{
		Name:                 model.Name.ValueString(),
		ExtractBinaryContent: model.ExtractBinaryContent.ValueBool(),
		ReduceWhitespace:     model.ReduceWhitespace.ValueBool(),
		RunMlInference:       model.RunMlInference.ValueBool(),
	}
}

func (data ContentConnectorData) toSchedulingAPI(ctx context.Context, diags *diag.Diagnostics) estypes.SchedulingConfiguration {
	var model connector.SchedulingModel
	diags.Append(data.Scheduling.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return estypes.SchedulingConfiguration{}
	}
	return estypes.SchedulingConfiguration{
		Full:          scheduleEntryToAPI(ctx, model.Full, diags),
		Incremental:   scheduleEntryToAPI(ctx, model.Incremental, diags),
		AccessControl: scheduleEntryToAPI(ctx, model.AccessControl, diags),
	}
}

// scheduleEntryToAPI maps an optional ScheduleEntryModel object to *ConnectorScheduling.
// Returns nil for null/unknown objects so the API treats the entry as "leave as-is".
func scheduleEntryToAPI(ctx context.Context, obj fwtypes.Object, diags *diag.Diagnostics) *estypes.ConnectorScheduling {
	if obj.IsNull() || !typeutils.IsKnown(obj) {
		return nil
	}
	var entry connector.ScheduleEntryModel
	diags.Append(obj.As(ctx, &entry, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return &estypes.ConnectorScheduling{
		Enabled:  entry.Enabled.ValueBool(),
		Interval: entry.Interval.ValueString(),
	}
}

func (data ContentConnectorData) toFeaturesAPI(ctx context.Context, diags *diag.Diagnostics) estypes.ConnectorFeatures {
	var model connector.FeaturesModel
	diags.Append(data.Features.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return estypes.ConnectorFeatures{}
	}
	out := estypes.ConnectorFeatures{
		DocumentLevelSecurity:  featureFlagToAPI(ctx, model.DocumentLevelSecurity, diags),
		IncrementalSync:        featureFlagToAPI(ctx, model.IncrementalSync, diags),
		NativeConnectorApiKeys: featureFlagToAPI(ctx, model.NativeConnectorAPIKeys, diags),
	}
	if !model.SyncRules.IsNull() && typeutils.IsKnown(model.SyncRules) {
		var rules connector.SyncRulesModel
		diags.Append(model.SyncRules.As(ctx, &rules, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return out
		}
		out.SyncRules = &estypes.SyncRulesFeature{
			Basic:    featureFlagToAPI(ctx, rules.Basic, diags),
			Advanced: featureFlagToAPI(ctx, rules.Advanced, diags),
		}
	}
	return out
}

// featureFlagToAPI maps an optional FeatureFlagModel object to *FeatureEnabled.
// Returns nil for null/unknown objects so the API treats the flag as "leave as-is".
func featureFlagToAPI(ctx context.Context, obj fwtypes.Object, diags *diag.Diagnostics) *estypes.FeatureEnabled {
	if obj.IsNull() || !typeutils.IsKnown(obj) {
		return nil
	}
	var flag connector.FeatureFlagModel
	diags.Append(obj.As(ctx, &flag, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}
	return &estypes.FeatureEnabled{Enabled: flag.Enabled.ValueBool()}
}

func encodeConfigurationValuesWire(
	planMap, configMap map[string]connector.ConfigurationValueModel,
	diags *diag.Diagnostics,
) map[string]json.RawMessage {
	if len(planMap) == 0 {
		return nil
	}
	out := make(map[string]json.RawMessage, len(planMap))
	for key, planElem := range planMap {
		configElem := planElem
		if configMap != nil {
			if ce, ok := configMap[key]; ok {
				configElem = ce
			}
		}
		raw, err := configurationValueToWireJSON(configElem)
		if err != nil {
			diags.AddError(
				"Invalid configuration value",
				fmt.Sprintf("configuration_values[%q]: %s", key, err.Error()),
			)
			return nil
		}
		out[key] = raw
	}
	return out
}

func configurationValueToWireJSON(elem connector.ConfigurationValueModel) (json.RawMessage, error) {
	switch {
	case typeutils.IsKnown(elem.SecretValue):
		return json.Marshal(elem.SecretValue.ValueString())
	case typeutils.IsKnown(elem.String):
		return json.Marshal(elem.String.ValueString())
	case typeutils.IsKnown(elem.Number):
		f, acc := elem.Number.ValueBigFloat().Float64()
		if acc != big.Exact && acc != big.Below {
			return json.Marshal(elem.Number.ValueBigFloat().Text('f', -1))
		}
		return json.Marshal(f)
	case typeutils.IsKnown(elem.Bool):
		return json.Marshal(elem.Bool.ValueBool())
	case typeutils.IsKnown(elem.JSON):
		return json.RawMessage(elem.JSON.ValueString()), nil
	default:
		return nil, fmt.Errorf("no value branch is set")
	}
}

func activeConfigurationBranch(elem connector.ConfigurationValueModel) string {
	switch {
	case typeutils.IsKnown(elem.SecretValue):
		return connector.SecretValueBranchAttr
	case !elem.SecretValue.IsNull() && elem.SecretValue.IsUnknown():
		return connector.SecretValueBranchAttr
	case typeutils.IsKnown(elem.String):
		return connector.StringBranchAttr
	case typeutils.IsKnown(elem.Number):
		return connector.NumberBranchAttr
	case typeutils.IsKnown(elem.Bool):
		return connector.BoolBranchAttr
	case typeutils.IsKnown(elem.JSON):
		return connector.JSONBranchAttr
	default:
		return ""
	}
}

func populateConfigurationValuesFromAPI(
	ctx context.Context,
	resp *getconnector.Response,
	priorMap map[string]connector.ConfigurationValueModel,
	diags *diag.Diagnostics,
) fwtypes.Map {
	apiValues := configurationValuesFromAPIResponse(resp.Configuration)
	result := make(map[string]connector.ConfigurationValueModel)

	if priorMap != nil {
		for key, priorElem := range priorMap {
			apiValue, hasAPI := apiValues[key]
			if !hasAPI {
				continue
			}
			schemaEntry, hasSchema := resp.Configuration[key]
			branch := activeConfigurationBranch(priorElem)
			if hasSchema && schemaEntry.Sensitive && branch != connector.SecretValueBranchAttr {
				diags.AddWarning(
					"Sensitive configuration value should use secret_value",
					fmt.Sprintf(
						`configuration_values["%s"] is marked sensitive in the connector schema; move the value to the secret_value branch to manage it safely.`,
						key,
					),
				)
				result[key] = priorElem
				continue
			}
			if branch == connector.SecretValueBranchAttr {
				result[key] = priorElem
				continue
			}
			decoded, d := decodeConfigurationValueIntoBranch(apiValue, branch)
			diags.Append(d...)
			if diags.HasError() {
				return fwtypes.MapNull(fwtypes.ObjectType{AttrTypes: connector.ConfigurationValueModelAttrTypes()})
			}
			result[key] = decoded
		}
	} else {
		for key, apiValue := range apiValues {
			schemaEntry, hasSchema := resp.Configuration[key]
			if hasSchema && schemaEntry.Sensitive {
				continue
			}
			branch := schemaTypeToBranch(schemaEntry.Type)
			decoded, d := decodeConfigurationValueIntoBranch(apiValue, branch)
			diags.Append(d...)
			if diags.HasError() {
				return fwtypes.MapNull(fwtypes.ObjectType{AttrTypes: connector.ConfigurationValueModelAttrTypes()})
			}
			result[key] = decoded
		}
	}

	if len(result) == 0 && priorMap == nil {
		return fwtypes.MapNull(fwtypes.ObjectType{AttrTypes: connector.ConfigurationValueModelAttrTypes()})
	}

	return typeutils.MapValueFrom(ctx, result, fwtypes.ObjectType{AttrTypes: connector.ConfigurationValueModelAttrTypes()}, configurationValuesPath, diags)
}

func configurationValuePresent(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	if string(raw) == connector.JSONNullLiteral {
		return false
	}
	return true
}

func configurationValuesFromAPIResponse(
	configuration estypes.ConnectorConfiguration,
) map[string]json.RawMessage {
	out := make(map[string]json.RawMessage)
	for key, props := range configuration {
		if !configurationValuePresent(props.Value) {
			continue
		}
		out[key] = props.Value
	}
	return out
}

func schemaTypeToBranch(fieldType *connectorfieldtype.ConnectorFieldType) string {
	if fieldType == nil {
		return connector.StringBranchAttr
	}
	switch fieldType.Name {
	case connectorfieldtype.Int.Name:
		return connector.NumberBranchAttr
	case connectorfieldtype.Bool.Name:
		return connector.BoolBranchAttr
	case connectorfieldtype.List.Name:
		return connector.StringBranchAttr
	default:
		return connector.StringBranchAttr
	}
}

func decodeConfigurationValueIntoBranch(
	raw json.RawMessage,
	branch string,
) (connector.ConfigurationValueModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch branch {
	case connector.StringBranchAttr, connector.SecretValueBranchAttr:
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return connector.ConfigurationValueModel{String: fwtypes.StringValue(s)}, diags
		}
		var n json.Number
		if err := json.Unmarshal(raw, &n); err == nil {
			return connector.ConfigurationValueModel{String: fwtypes.StringValue(n.String())}, diags
		}
		var b bool
		if err := json.Unmarshal(raw, &b); err == nil {
			return connector.ConfigurationValueModel{String: fwtypes.StringValue(fmt.Sprintf("%t", b))}, diags
		}
		diags.AddError("Failed to decode configuration value", "value is not a JSON string")
		return connector.ConfigurationValueModel{}, diags
	case connector.NumberBranchAttr:
		var n json.Number
		if err := json.Unmarshal(raw, &n); err != nil {
			diags.AddError("Failed to decode configuration value", err.Error())
			return connector.ConfigurationValueModel{}, diags
		}
		bf, _, err := big.ParseFloat(n.String(), 10, 64, big.ToNearestEven)
		if err != nil {
			diags.AddError("Failed to decode configuration value", err.Error())
			return connector.ConfigurationValueModel{}, diags
		}
		return connector.ConfigurationValueModel{Number: fwtypes.NumberValue(bf)}, diags
	case connector.BoolBranchAttr:
		var b bool
		if err := json.Unmarshal(raw, &b); err != nil {
			diags.AddError("Failed to decode configuration value", err.Error())
			return connector.ConfigurationValueModel{}, diags
		}
		return connector.ConfigurationValueModel{Bool: fwtypes.BoolValue(b)}, diags
	case connector.JSONBranchAttr:
		return connector.ConfigurationValueModel{JSON: jsontypes.NewNormalizedValue(string(raw))}, diags
	default:
		diags.AddError("Failed to decode configuration value", fmt.Sprintf("unsupported branch %q", branch))
		return connector.ConfigurationValueModel{}, diags
	}
}

func planObjectSet(obj fwtypes.Object) bool {
	return typeutils.IsKnown(obj) && !obj.IsNull()
}

func planMapSet(m fwtypes.Map) bool {
	return typeutils.IsKnown(m) && !m.IsNull()
}
