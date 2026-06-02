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

// Package connector holds shared model types, attribute-type helpers, and
// API→state converters used by both the connector resource and data source
// subpackages. The resource (registered as elasticstack_elasticsearch_connector)
// lives in connector/resource, the data source in connector/data_source, and
// the provider-defined sync-job action in connector/sync_job_create. Nothing
// in this base package registers a Terraform entity directly — it exists only
// so the entities can share model shape and deserialization.
package connector

import (
	"context"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// MinSupportedVersion is the minimum Elasticsearch version supported by the
// connector resource and data source.
//
// The connector APIs are GA from Elasticsearch 8.12.0, but the request body
// shapes used by the typed go-elasticsearch client (specifically the
// `connector_id` field on POST /_connector and the `rules` field on
// PUT /_connector/{id}/_filtering) only stabilized in 8.16.0. Older 8.12.x–
// 8.15.x clusters reject those payloads, so the provider pins both this
// resource and the data source to 8.16.0 as the minimum supported floor.
var MinSupportedVersion = version.Must(version.NewVersion("8.16.0"))

// VersionGate is a zero-size embedded struct that satisfies
// entitycore.WithVersionRequirements for connector resource and data source models.
type VersionGate struct{}

func (VersionGate) GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *MinSupportedVersion,
		ErrorMessage: "elasticstack_elasticsearch_connector requires Elasticsearch 8.16.0 or later (the connector request bodies the typed client sends are rejected on 8.12.x–8.15.x).",
	}}, nil
}

// Shared attribute names used by the resource and data source schemas plus
// the API↔state converters. Exported so the resource/data_source subpackages
// can reference them without hardcoding string literals.
const (
	NameAttr     = "name"
	EnabledAttr  = "enabled"
	IntervalAttr = "interval"

	ExtractBinaryContentAttr = "extract_binary_content"
	ReduceWhitespaceAttr     = "reduce_whitespace"
	RunMlInferenceAttr       = "run_ml_inference"

	FullScheduleAttr          = "full"
	IncrementalScheduleAttr   = "incremental"
	AccessControlScheduleAttr = "access_control"

	BasicSyncRulesAttr    = "basic"
	AdvancedSyncRulesAttr = "advanced"

	DocumentLevelSecurityAttr  = "document_level_security"
	IncrementalSyncAttr        = "incremental_sync"
	NativeConnectorAPIKeysAttr = "native_connector_api_keys"
	SyncRulesAttr              = "sync_rules"

	StringBranchAttr      = "string"
	NumberBranchAttr      = "number"
	BoolBranchAttr        = "bool"
	JSONBranchAttr        = "json"
	SecretValueBranchAttr = "secret_value"

	// JSONNullLiteral is the JSON `null` byte sequence used both by the
	// resource's configuration-value decoder and the data source's
	// jsontypes.Normalized field encoders to distinguish "value absent"
	// from "value present and JSON-null".
	JSONNullLiteral = "null"
)

// ConfigurationValueBranchAttrNames lists the configuration_value branches in
// the canonical schema order. The branch validator and converters iterate this
// slice so adding a new branch only requires touching one place.
var ConfigurationValueBranchAttrNames = []string{
	StringBranchAttr,
	NumberBranchAttr,
	BoolBranchAttr,
	JSONBranchAttr,
	SecretValueBranchAttr,
}

// PipelineModel represents the connector ingest pipeline settings.
type PipelineModel struct {
	Name                 fwtypes.String `tfsdk:"name"`
	ExtractBinaryContent fwtypes.Bool   `tfsdk:"extract_binary_content"`
	ReduceWhitespace     fwtypes.Bool   `tfsdk:"reduce_whitespace"`
	RunMlInference       fwtypes.Bool   `tfsdk:"run_ml_inference"`
}

// PipelineModelAttrTypes is the attribute-type map describing PipelineModel.
func PipelineModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		NameAttr:                 fwtypes.StringType,
		ExtractBinaryContentAttr: fwtypes.BoolType,
		ReduceWhitespaceAttr:     fwtypes.BoolType,
		RunMlInferenceAttr:       fwtypes.BoolType,
	}
}

// ScheduleEntryModel represents a single scheduling sub-block (full, incremental, or access_control).
type ScheduleEntryModel struct {
	Enabled  fwtypes.Bool   `tfsdk:"enabled"`
	Interval fwtypes.String `tfsdk:"interval"`
}

// ScheduleEntryModelAttrTypes is the attribute-type map describing ScheduleEntryModel.
func ScheduleEntryModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		EnabledAttr:  fwtypes.BoolType,
		IntervalAttr: fwtypes.StringType,
	}
}

// SchedulingModel represents connector sync scheduling.
type SchedulingModel struct {
	Full          fwtypes.Object `tfsdk:"full"`
	Incremental   fwtypes.Object `tfsdk:"incremental"`
	AccessControl fwtypes.Object `tfsdk:"access_control"`
}

// SchedulingModelAttrTypes is the attribute-type map describing SchedulingModel.
func SchedulingModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		FullScheduleAttr:          fwtypes.ObjectType{AttrTypes: ScheduleEntryModelAttrTypes()},
		IncrementalScheduleAttr:   fwtypes.ObjectType{AttrTypes: ScheduleEntryModelAttrTypes()},
		AccessControlScheduleAttr: fwtypes.ObjectType{AttrTypes: ScheduleEntryModelAttrTypes()},
	}
}

// FeatureFlagModel represents a feature toggle with a single enabled flag.
type FeatureFlagModel struct {
	Enabled fwtypes.Bool `tfsdk:"enabled"`
}

// FeatureFlagModelAttrTypes is the attribute-type map describing FeatureFlagModel.
func FeatureFlagModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		EnabledAttr: fwtypes.BoolType,
	}
}

// SyncRulesModel represents sync rules feature flags.
type SyncRulesModel struct {
	Basic    fwtypes.Object `tfsdk:"basic"`
	Advanced fwtypes.Object `tfsdk:"advanced"`
}

// SyncRulesModelAttrTypes is the attribute-type map describing SyncRulesModel.
func SyncRulesModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		BasicSyncRulesAttr:    fwtypes.ObjectType{AttrTypes: FeatureFlagModelAttrTypes()},
		AdvancedSyncRulesAttr: fwtypes.ObjectType{AttrTypes: FeatureFlagModelAttrTypes()},
	}
}

// FeaturesModel represents connector feature flags.
type FeaturesModel struct {
	DocumentLevelSecurity  fwtypes.Object `tfsdk:"document_level_security"`
	IncrementalSync        fwtypes.Object `tfsdk:"incremental_sync"`
	NativeConnectorAPIKeys fwtypes.Object `tfsdk:"native_connector_api_keys"`
	SyncRules              fwtypes.Object `tfsdk:"sync_rules"`
}

// FeaturesModelAttrTypes is the attribute-type map describing FeaturesModel.
func FeaturesModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		DocumentLevelSecurityAttr:  fwtypes.ObjectType{AttrTypes: FeatureFlagModelAttrTypes()},
		IncrementalSyncAttr:        fwtypes.ObjectType{AttrTypes: FeatureFlagModelAttrTypes()},
		NativeConnectorAPIKeysAttr: fwtypes.ObjectType{AttrTypes: FeatureFlagModelAttrTypes()},
		SyncRulesAttr:              fwtypes.ObjectType{AttrTypes: SyncRulesModelAttrTypes()},
	}
}

// ConfigurationValueModel is a branch-typed configuration value element.
// Exactly one of String, Number, Bool, JSON, or SecretValue must be set —
// the resource's configurationValueBranchValidator enforces this at plan time.
type ConfigurationValueModel struct {
	String      fwtypes.String       `tfsdk:"string"`
	Number      fwtypes.Number       `tfsdk:"number"`
	Bool        fwtypes.Bool         `tfsdk:"bool"`
	JSON        jsontypes.Normalized `tfsdk:"json"`
	SecretValue fwtypes.String       `tfsdk:"secret_value"`
}

// ConfigurationValueModelAttrTypes is the attribute-type map describing
// ConfigurationValueModel.
func ConfigurationValueModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		StringBranchAttr:      fwtypes.StringType,
		NumberBranchAttr:      fwtypes.NumberType,
		BoolBranchAttr:        fwtypes.BoolType,
		JSONBranchAttr:        jsontypes.NormalizedType{},
		SecretValueBranchAttr: fwtypes.StringType,
	}
}
