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

package template

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model is the Terraform plan/state shape for the index template resource and data source.
type Model struct {
	entitycore.ElasticsearchConnectionField
	ID                              types.String         `tfsdk:"id"`
	Name                            types.String         `tfsdk:"name"`
	ComposedOf                      types.List           `tfsdk:"composed_of"`
	IgnoreMissingComponentTemplates types.List           `tfsdk:"ignore_missing_component_templates"`
	IndexPatterns                   types.Set            `tfsdk:"index_patterns"`
	Metadata                        jsontypes.Normalized `tfsdk:"metadata"`
	Priority                        types.Int64          `tfsdk:"priority"`
	Version                         types.Int64          `tfsdk:"version"`
	DataStream                      types.Object         `tfsdk:"data_stream"`
	Template                        types.Object         `tfsdk:"template"`
}

// GetID satisfies [entitycore.ElasticsearchResourceModel].
func (m Model) GetID() types.String { return m.ID }

// GetResourceID satisfies [entitycore.ElasticsearchResourceModel].
// For index templates the write identity is the template name.
func (m Model) GetResourceID() types.String { return m.Name }

// DataStreamModel is the inner shape of the data_stream block (for Object.As).
type DataStreamModel struct {
	Hidden             types.Bool `tfsdk:"hidden"`
	AllowCustomRouting types.Bool `tfsdk:"allow_custom_routing"`
}

// TemplateBlockModel is the inner shape of the template block (for Object.As).
//
//nolint:revive // Name documents the template {} block; BlockModel alone is ambiguous in this package.
type TemplateBlockModel struct {
	Alias             types.Set                      `tfsdk:"alias"`
	Mappings          index.MappingsValue            `tfsdk:"mappings"`
	Settings          customtypes.IndexSettingsValue `tfsdk:"settings"`
	Lifecycle         types.Object                   `tfsdk:"lifecycle"`
	DataStreamOptions types.Object                   `tfsdk:"data_stream_options"`
}

// LifecycleModel is the inner shape of template.lifecycle.
type LifecycleModel struct {
	DataRetention types.String `tfsdk:"data_retention"`
}

// DataStreamOptionsModel is the inner shape of template.data_stream_options.
type DataStreamOptionsModel struct {
	FailureStore types.Object `tfsdk:"failure_store"`
}

// FailureStoreModel is the inner shape of template.data_stream_options.failure_store.
type FailureStoreModel struct {
	Enabled   types.Bool   `tfsdk:"enabled"`
	Lifecycle types.Object `tfsdk:"lifecycle"`
}

// FailureStoreLifecycleModel is the inner shape of failure_store.lifecycle.
type FailureStoreLifecycleModel struct {
	DataRetention types.String `tfsdk:"data_retention"`
}

// DataStreamAttrTypes returns attribute types for the data_stream block object.
func DataStreamAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"hidden":               types.BoolType,
		"allow_custom_routing": types.BoolType,
	}
}

// LifecycleAttrTypes returns attribute types for template.lifecycle.
func LifecycleAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"data_retention": types.StringType,
	}
}

// FailureStoreLifecycleAttrTypes returns attribute types for failure_store.lifecycle.
func FailureStoreLifecycleAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"data_retention": types.StringType,
	}
}

// FailureStoreAttrTypes returns attribute types for failure_store.
func FailureStoreAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":   types.BoolType,
		"lifecycle": types.ObjectType{AttrTypes: FailureStoreLifecycleAttrTypes()},
	}
}

// DataStreamOptionsAttrTypes returns attribute types for template.data_stream_options.
func DataStreamOptionsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"failure_store": types.ObjectType{AttrTypes: FailureStoreAttrTypes()},
	}
}

// TemplateAttrTypes returns attribute types for the template block object.
//
//nolint:revive // Name matches OpenSpec task wording (template block attribute types).
func TemplateAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"alias":               types.SetType{ElemType: NewAliasObjectType()},
		"mappings":            index.MappingsType{},
		"settings":            customtypes.IndexSettingsType{},
		"lifecycle":           types.ObjectType{AttrTypes: LifecycleAttrTypes()},
		"data_stream_options": types.ObjectType{AttrTypes: DataStreamOptionsAttrTypes()},
	}
}
