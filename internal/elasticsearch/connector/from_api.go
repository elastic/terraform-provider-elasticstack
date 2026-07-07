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

package connector

import (
	"context"

	getconnector "github.com/elastic/go-elasticsearch/v9/typedapi/connector/get"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// CoreConnectorFields holds the scalar connector fields shared between the
// resource and data source models.
type CoreConnectorFields struct {
	ServiceType    fwtypes.String `tfsdk:"service_type"`
	Name           fwtypes.String `tfsdk:"name"`
	Description    fwtypes.String `tfsdk:"description"`
	IndexName      fwtypes.String `tfsdk:"index_name"`
	IsNative       fwtypes.Bool   `tfsdk:"is_native"`
	Language       fwtypes.String `tfsdk:"language"`
	APIKeyID       fwtypes.String `tfsdk:"api_key_id"`
	APIKeySecretID fwtypes.String `tfsdk:"api_key_secret_id"`
	Pipeline       fwtypes.Object `tfsdk:"pipeline"`
	Scheduling     fwtypes.Object `tfsdk:"scheduling"`
	Features       fwtypes.Object `tfsdk:"features"`
}

// PopulateCoreConnectorFieldsFromAPI maps scalar connector fields from the API
// response into a CoreConnectorFields value, setting null for absent optional
// fields. Shared by the connector resource Read and the data source Read.
func PopulateCoreConnectorFieldsFromAPI(
	ctx context.Context,
	resp *getconnector.Response,
	diags *diag.Diagnostics,
) CoreConnectorFields {
	f := CoreConnectorFields{IsNative: fwtypes.BoolValue(resp.IsNative)}

	f.ServiceType = typeutils.StringishPointerValue(resp.ServiceType)
	f.Name = typeutils.StringishPointerValue(resp.Name)
	f.Description = typeutils.StringishPointerValue(resp.Description)
	f.IndexName = typeutils.StringishPointerValue(resp.IndexName)
	f.Language = typeutils.StringishPointerValue(resp.Language)
	f.APIKeyID = typeutils.StringishPointerValue(resp.ApiKeyId)
	f.APIKeySecretID = typeutils.StringishPointerValue(resp.ApiKeySecretId)

	f.Pipeline = PopulatePipelineFromAPI(ctx, resp.Pipeline, diags)
	f.Scheduling = PopulateSchedulingFromAPI(ctx, resp.Scheduling, diags)
	f.Features = PopulateFeaturesFromAPI(ctx, resp.Features, diags)

	return f
}

// PopulatePipelineFromAPI converts an Elasticsearch IngestPipelineParams into
// a typed Terraform object, returning a null object when the API value is nil.
// Shared by the connector resource Read and the data source Read.
func PopulatePipelineFromAPI(ctx context.Context, pipeline *estypes.IngestPipelineParams, diags *diag.Diagnostics) fwtypes.Object {
	if pipeline == nil {
		return fwtypes.ObjectNull(PipelineModelAttrTypes())
	}
	model := PipelineModel{
		Name:                 fwtypes.StringValue(pipeline.Name),
		ExtractBinaryContent: fwtypes.BoolValue(pipeline.ExtractBinaryContent),
		ReduceWhitespace:     fwtypes.BoolValue(pipeline.ReduceWhitespace),
		RunMlInference:       fwtypes.BoolValue(pipeline.RunMlInference),
	}
	obj, d := fwtypes.ObjectValueFrom(ctx, PipelineModelAttrTypes(), model)
	diags.Append(d...)
	return obj
}

// PopulateSchedulingFromAPI converts an Elasticsearch SchedulingConfiguration
// into a typed Terraform object.
func PopulateSchedulingFromAPI(ctx context.Context, scheduling estypes.SchedulingConfiguration, diags *diag.Diagnostics) fwtypes.Object {
	model := SchedulingModel{
		Full:          ScheduleEntryFromAPI(ctx, scheduling.Full, diags),
		Incremental:   ScheduleEntryFromAPI(ctx, scheduling.Incremental, diags),
		AccessControl: ScheduleEntryFromAPI(ctx, scheduling.AccessControl, diags),
	}
	obj, d := fwtypes.ObjectValueFrom(ctx, SchedulingModelAttrTypes(), model)
	diags.Append(d...)
	return obj
}

// ScheduleEntryFromAPI converts a single Elasticsearch ConnectorScheduling
// entry into a typed Terraform object, returning a null object when the API
// value is nil.
func ScheduleEntryFromAPI(ctx context.Context, entry *estypes.ConnectorScheduling, diags *diag.Diagnostics) fwtypes.Object {
	if entry == nil {
		return fwtypes.ObjectNull(ScheduleEntryModelAttrTypes())
	}
	model := ScheduleEntryModel{
		Enabled:  fwtypes.BoolValue(entry.Enabled),
		Interval: fwtypes.StringValue(entry.Interval),
	}
	obj, d := fwtypes.ObjectValueFrom(ctx, ScheduleEntryModelAttrTypes(), model)
	diags.Append(d...)
	return obj
}

// PopulateFeaturesFromAPI converts an Elasticsearch ConnectorFeatures into a
// typed Terraform object, returning a null object when the API value is nil.
func PopulateFeaturesFromAPI(ctx context.Context, features *estypes.ConnectorFeatures, diags *diag.Diagnostics) fwtypes.Object {
	if features == nil {
		return fwtypes.ObjectNull(FeaturesModelAttrTypes())
	}
	model := FeaturesModel{
		DocumentLevelSecurity:  FeatureFlagFromAPI(ctx, features.DocumentLevelSecurity, diags),
		IncrementalSync:        FeatureFlagFromAPI(ctx, features.IncrementalSync, diags),
		NativeConnectorAPIKeys: FeatureFlagFromAPI(ctx, features.NativeConnectorApiKeys, diags),
		SyncRules:              SyncRulesFromAPI(ctx, features.SyncRules, diags),
	}
	obj, d := fwtypes.ObjectValueFrom(ctx, FeaturesModelAttrTypes(), model)
	diags.Append(d...)
	return obj
}

// FeatureFlagFromAPI converts an Elasticsearch FeatureEnabled into a typed
// Terraform object, returning a null object when the API value is nil.
func FeatureFlagFromAPI(ctx context.Context, flag *estypes.FeatureEnabled, diags *diag.Diagnostics) fwtypes.Object {
	if flag == nil {
		return fwtypes.ObjectNull(FeatureFlagModelAttrTypes())
	}
	model := FeatureFlagModel{Enabled: fwtypes.BoolValue(flag.Enabled)}
	obj, d := fwtypes.ObjectValueFrom(ctx, FeatureFlagModelAttrTypes(), model)
	diags.Append(d...)
	return obj
}

// SyncRulesFromAPI converts an Elasticsearch SyncRulesFeature into a typed
// Terraform object, returning a null object when the API value is nil.
func SyncRulesFromAPI(ctx context.Context, rules *estypes.SyncRulesFeature, diags *diag.Diagnostics) fwtypes.Object {
	if rules == nil {
		return fwtypes.ObjectNull(SyncRulesModelAttrTypes())
	}
	model := SyncRulesModel{
		Basic:    FeatureFlagFromAPI(ctx, rules.Basic, diags),
		Advanced: FeatureFlagFromAPI(ctx, rules.Advanced, diags),
	}
	obj, d := fwtypes.ObjectValueFrom(ctx, SyncRulesModelAttrTypes(), model)
	diags.Append(d...)
	return obj
}
