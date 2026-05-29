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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	nameAttrName     = "name"
	enabledAttrName  = "enabled"
	intervalAttrName = "interval"

	stringBranchAttrName      = "string"
	numberBranchAttrName      = "number"
	boolBranchAttrName        = "bool"
	jsonBranchAttrName        = "json"
	secretValueBranchAttrName = "secret_value"
)

var configurationValueBranchAttrNames = []string{
	stringBranchAttrName,
	numberBranchAttrName,
	boolBranchAttrName,
	jsonBranchAttrName,
	secretValueBranchAttrName,
}

// ContentConnectorData is the Terraform state model for the content connector resource.
type ContentConnectorData struct {
	entitycore.ElasticsearchConnectionField
	ID                  fwtypes.String `tfsdk:"id"`
	ConnectorID         fwtypes.String `tfsdk:"connector_id"`
	ServiceType         fwtypes.String `tfsdk:"service_type"`
	Name                fwtypes.String `tfsdk:"name"`
	Description         fwtypes.String `tfsdk:"description"`
	IndexName           fwtypes.String `tfsdk:"index_name"`
	IsNative            fwtypes.Bool   `tfsdk:"is_native"`
	Language            fwtypes.String `tfsdk:"language"`
	APIKeyID            fwtypes.String `tfsdk:"api_key_id"`
	APIKeySecretID      fwtypes.String `tfsdk:"api_key_secret_id"`
	Pipeline            fwtypes.Object `tfsdk:"pipeline"`
	Scheduling          fwtypes.Object `tfsdk:"scheduling"`
	Features            fwtypes.Object `tfsdk:"features"`
	ConfigurationValues fwtypes.Map    `tfsdk:"configuration_values"`
}

func (data ContentConnectorData) GetID() fwtypes.String         { return data.ID }
func (data ContentConnectorData) GetResourceID() fwtypes.String { return data.ConnectorID }
func (data ContentConnectorData) GetElasticsearchConnection() fwtypes.List {
	return data.ElasticsearchConnection
}

var (
	_ entitycore.ElasticsearchResourceModel = ContentConnectorData{}
	_ entitycore.WithVersionRequirements    = ContentConnectorData{}
)

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements].
func (data ContentConnectorData) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *MinSupportedVersion,
		ErrorMessage: "elasticstack_elasticsearch_connector requires Elasticsearch 8.12.0 or later (connector APIs GA).",
	}}, nil
}

// PipelineModel represents the connector ingest pipeline settings.
type PipelineModel struct {
	Name                 fwtypes.String `tfsdk:"name"`
	ExtractBinaryContent fwtypes.Bool   `tfsdk:"extract_binary_content"`
	ReduceWhitespace     fwtypes.Bool   `tfsdk:"reduce_whitespace"`
	RunMlInference       fwtypes.Bool   `tfsdk:"run_ml_inference"`
}

func pipelineModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		nameAttrName:             fwtypes.StringType,
		"extract_binary_content": fwtypes.BoolType,
		"reduce_whitespace":      fwtypes.BoolType,
		"run_ml_inference":       fwtypes.BoolType,
	}
}

// ScheduleEntryModel represents a single scheduling sub-block (full, incremental, or access_control).
type ScheduleEntryModel struct {
	Enabled  fwtypes.Bool   `tfsdk:"enabled"`
	Interval fwtypes.String `tfsdk:"interval"`
}

func scheduleEntryModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		enabledAttrName:  fwtypes.BoolType,
		intervalAttrName: fwtypes.StringType,
	}
}

// SchedulingModel represents connector sync scheduling.
type SchedulingModel struct {
	Full          fwtypes.Object `tfsdk:"full"`
	Incremental   fwtypes.Object `tfsdk:"incremental"`
	AccessControl fwtypes.Object `tfsdk:"access_control"`
}

func schedulingModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"full":           fwtypes.ObjectType{AttrTypes: scheduleEntryModelAttrTypes()},
		"incremental":    fwtypes.ObjectType{AttrTypes: scheduleEntryModelAttrTypes()},
		"access_control": fwtypes.ObjectType{AttrTypes: scheduleEntryModelAttrTypes()},
	}
}

// FeatureFlagModel represents a feature toggle with a single enabled flag.
type FeatureFlagModel struct {
	Enabled fwtypes.Bool `tfsdk:"enabled"`
}

func featureFlagModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		enabledAttrName: fwtypes.BoolType,
	}
}

// SyncRulesModel represents sync rules feature flags.
type SyncRulesModel struct {
	Basic    fwtypes.Object `tfsdk:"basic"`
	Advanced fwtypes.Object `tfsdk:"advanced"`
}

func syncRulesModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"basic":    fwtypes.ObjectType{AttrTypes: featureFlagModelAttrTypes()},
		"advanced": fwtypes.ObjectType{AttrTypes: featureFlagModelAttrTypes()},
	}
}

// FeaturesModel represents connector feature flags.
type FeaturesModel struct {
	DocumentLevelSecurity  fwtypes.Object `tfsdk:"document_level_security"`
	IncrementalSync        fwtypes.Object `tfsdk:"incremental_sync"`
	NativeConnectorAPIKeys fwtypes.Object `tfsdk:"native_connector_api_keys"`
	SyncRules              fwtypes.Object `tfsdk:"sync_rules"`
}

func featuresModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"document_level_security":   fwtypes.ObjectType{AttrTypes: featureFlagModelAttrTypes()},
		"incremental_sync":          fwtypes.ObjectType{AttrTypes: featureFlagModelAttrTypes()},
		"native_connector_api_keys": fwtypes.ObjectType{AttrTypes: featureFlagModelAttrTypes()},
		"sync_rules":                fwtypes.ObjectType{AttrTypes: syncRulesModelAttrTypes()},
	}
}

// ConfigurationValueModel is a branch-typed configuration value element.
type ConfigurationValueModel struct {
	String      fwtypes.String       `tfsdk:"string"`
	Number      fwtypes.Number       `tfsdk:"number"`
	Bool        fwtypes.Bool         `tfsdk:"bool"`
	JSON        jsontypes.Normalized `tfsdk:"json"`
	SecretValue fwtypes.String       `tfsdk:"secret_value"`
}

func configurationValueModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		stringBranchAttrName:      fwtypes.StringType,
		numberBranchAttrName:      fwtypes.NumberType,
		boolBranchAttrName:        fwtypes.BoolType,
		jsonBranchAttrName:        jsontypes.NormalizedType{},
		secretValueBranchAttrName: fwtypes.StringType,
	}
}
