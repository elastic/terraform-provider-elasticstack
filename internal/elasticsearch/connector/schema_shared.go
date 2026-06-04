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
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// LeafAttrsForResource converts shared LeafAttr definitions into resource schema attributes.
// All leaf attributes are marked Required.
func LeafAttrsForResource(leaves []LeafAttr) map[string]rschema.Attribute {
	result := make(map[string]rschema.Attribute, len(leaves))
	for _, a := range leaves {
		if a.IsString {
			result[a.Name] = rschema.StringAttribute{
				MarkdownDescription: a.Description,
				Required:            true,
			}
		} else {
			result[a.Name] = rschema.BoolAttribute{
				MarkdownDescription: a.Description,
				Required:            true,
			}
		}
	}
	return result
}

// LeafAttrsForDataSource converts shared LeafAttr definitions into data source schema attributes.
// All leaf attributes are marked Computed.
func LeafAttrsForDataSource(leaves []LeafAttr) map[string]dschema.Attribute {
	result := make(map[string]dschema.Attribute, len(leaves))
	for _, a := range leaves {
		if a.IsString {
			result[a.Name] = dschema.StringAttribute{
				MarkdownDescription: a.Description,
				Computed:            true,
			}
		} else {
			result[a.Name] = dschema.BoolAttribute{
				MarkdownDescription: a.Description,
				Computed:            true,
			}
		}
	}
	return result
}

// PipelineAttrForResource returns the pipeline nested attribute for the connector resource schema.
func PipelineAttrForResource() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		MarkdownDescription: PipelineNestedDesc + " Changes trigger `PUT /_connector/{id}/_pipeline`.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: LeafAttrsForResource(PipelineLeafAttrs()),
	}
}

// PipelineAttrForDataSource returns the pipeline nested attribute for the connector data source schema.
func PipelineAttrForDataSource() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: PipelineNestedDesc,
		Computed:            true,
		Attributes:          LeafAttrsForDataSource(PipelineLeafAttrs()),
	}
}

// SchedulingAttrForResource returns the scheduling nested attribute for the connector resource schema.
func SchedulingAttrForResource() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		MarkdownDescription: SchedulingNestedDesc + " Changes trigger `PUT /_connector/{id}/_scheduling`.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]rschema.Attribute{
			FullScheduleAttr:          ScheduleEntryAttrForResource(FullScheduleAttr),
			IncrementalScheduleAttr:   ScheduleEntryAttrForResource(IncrementalScheduleAttr),
			AccessControlScheduleAttr: ScheduleEntryAttrForResource(AccessControlScheduleAttr),
		},
	}
}

// SchedulingAttrForDataSource returns the scheduling nested attribute for the connector data source schema.
func SchedulingAttrForDataSource() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: SchedulingNestedDesc,
		Computed:            true,
		Attributes: map[string]dschema.Attribute{
			FullScheduleAttr:          ScheduleEntryAttrForDataSource(FullScheduleAttr),
			IncrementalScheduleAttr:   ScheduleEntryAttrForDataSource(IncrementalScheduleAttr),
			AccessControlScheduleAttr: ScheduleEntryAttrForDataSource(AccessControlScheduleAttr),
		},
	}
}

// ScheduleEntryAttrForResource returns a schedule entry nested attribute for the connector resource schema.
func ScheduleEntryAttrForResource(jobKind string) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		MarkdownDescription: "Schedule for the `" + jobKind + "` sync job type.",
		Optional:            true,
		Attributes:          LeafAttrsForResource(ScheduleEntryLeafAttrs()),
	}
}

// ScheduleEntryAttrForDataSource returns a schedule entry nested attribute for the connector data source schema.
func ScheduleEntryAttrForDataSource(jobKind string) dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: "Schedule for the `" + jobKind + "` sync job type.",
		Computed:            true,
		Attributes:          LeafAttrsForDataSource(ScheduleEntryLeafAttrs()),
	}
}

// FeaturesAttrForResource returns the features nested attribute for the connector resource schema.
func FeaturesAttrForResource() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		MarkdownDescription: FeaturesNestedDesc + " Changes trigger `PUT /_connector/{id}/_features`.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]rschema.Attribute{
			DocumentLevelSecurityAttr:  FeatureFlagAttrForResource(DocumentLevelSecurityAttr),
			IncrementalSyncAttr:        FeatureFlagAttrForResource(IncrementalSyncAttr),
			NativeConnectorAPIKeysAttr: FeatureFlagAttrForResource(NativeConnectorAPIKeysAttr),
			SyncRulesAttr:              SyncRulesAttrForResource(),
		},
	}
}

// FeaturesAttrForDataSource returns the features nested attribute for the connector data source schema.
func FeaturesAttrForDataSource() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: FeaturesNestedDesc,
		Computed:            true,
		Attributes: map[string]dschema.Attribute{
			DocumentLevelSecurityAttr:  FeatureFlagAttrForDataSource(DocumentLevelSecurityAttr),
			IncrementalSyncAttr:        FeatureFlagAttrForDataSource(IncrementalSyncAttr),
			NativeConnectorAPIKeysAttr: FeatureFlagAttrForDataSource(NativeConnectorAPIKeysAttr),
			SyncRulesAttr:              SyncRulesAttrForDataSource(),
		},
	}
}

// FeatureFlagAttrForResource returns a feature flag nested attribute for the connector resource schema.
func FeatureFlagAttrForResource(featureName string) rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		MarkdownDescription: "Feature flag for `" + featureName + "`.",
		Optional:            true,
		Attributes:          LeafAttrsForResource(FeatureFlagLeafAttrs()),
	}
}

// FeatureFlagAttrForDataSource returns a feature flag nested attribute for the connector data source schema.
func FeatureFlagAttrForDataSource(featureName string) dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: "Feature flag for `" + featureName + "`.",
		Computed:            true,
		Attributes:          LeafAttrsForDataSource(FeatureFlagLeafAttrs()),
	}
}

// SyncRulesAttrForResource returns the sync rules nested attribute for the connector resource schema.
func SyncRulesAttrForResource() rschema.SingleNestedAttribute {
	return rschema.SingleNestedAttribute{
		MarkdownDescription: SyncRulesNestedDesc,
		Optional:            true,
		Attributes: map[string]rschema.Attribute{
			BasicSyncRulesAttr:    FeatureFlagAttrForResource(BasicSyncRulesAttr),
			AdvancedSyncRulesAttr: FeatureFlagAttrForResource(AdvancedSyncRulesAttr),
		},
	}
}

// SyncRulesAttrForDataSource returns the sync rules nested attribute for the connector data source schema.
func SyncRulesAttrForDataSource() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: SyncRulesNestedDesc,
		Computed:            true,
		Attributes: map[string]dschema.Attribute{
			BasicSyncRulesAttr:    FeatureFlagAttrForDataSource(BasicSyncRulesAttr),
			AdvancedSyncRulesAttr: FeatureFlagAttrForDataSource(AdvancedSyncRulesAttr),
		},
	}
}
