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

package queryrulesets

import (
	"context"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// MinSupportedVersion is the minimum Elasticsearch version for the Query Rules API (GA in 8.12).
var MinSupportedVersion = version.Must(version.NewVersion("8.12.0"))

var (
	ruleTypeValidator     = stringvalidator.OneOf("pinned", "exclude")
	criteriaTypeValidator = stringvalidator.OneOf(
		"always", "exact", "fuzzy", "prefix", "suffix", "contains", "lt", "lte", "gt", "gte",
	)
)

// schemaFactory returns the schema for the query ruleset resource. The
// elasticsearch_connection block is injected automatically by the envelope.
func schemaFactory(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: queryRulesetResourceMarkdownDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier in the form `<cluster_uuid>/<ruleset_id>`.",
				Computed:            true,
			},
			"ruleset_id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the query ruleset.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rules": rulesListNestedAttributeResource(),
		},
	}
}

func rulesListNestedAttributeResource() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Ordered list of query rules for this ruleset.",
		Required:            true,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: queryRuleNestedObjectResource(),
	}
}

func queryRuleNestedObjectResource() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			queryRuleRuleIDAttrName: schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the rule within the ruleset.",
				Required:            true,
			},
			queryRuleCriteriaTypeAttrName: schema.StringAttribute{
				MarkdownDescription: "Rule type: `pinned` or `exclude`.",
				Required:            true,
				Validators:          []validator.String{ruleTypeValidator},
			},
			queryRulePriorityAttrName: schema.Int64Attribute{
				MarkdownDescription: "Relative priority within the ruleset; omitted from the API when null.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			queryRuleCriteriaAttrName: queryRuleCriteriaListNestedAttributeResource(),
			queryRuleActionsAttrName:  queryRuleActionsSingleNestedAttributeResource(),
		},
	}
}

func queryRuleCriteriaListNestedAttributeResource() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Match criteria for the rule; all criteria must match for the rule to apply.",
		Required:            true,
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: queryRuleCriteriaNestedObjectResource(),
	}
}

func queryRuleCriteriaNestedObjectResource() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Validators: []validator.Object{
			queryRuleCriteriaValidator{},
		},
		Attributes: map[string]schema.Attribute{
			queryRuleCriteriaTypeAttrName: schema.StringAttribute{
				MarkdownDescription: "Criteria type (for example `exact`, `always`, `gt`).",
				Required:            true,
				Validators:          []validator.String{criteriaTypeValidator},
			},
			queryRuleCriteriaMetadataAttrName: schema.StringAttribute{
				MarkdownDescription: "Metadata field to match against; omitted from the API when null.",
				Optional:            true,
			},
			queryRuleCriteriaValuesAttrName: schema.StringAttribute{
				MarkdownDescription: "JSON-encoded array of string or numeric values; required unless `type` is `always`. Empty arrays are not allowed.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				Validators: []validator.String{
					criteriaValuesJSONValidator{},
				},
			},
		},
	}
}

func queryRuleActionsSingleNestedAttributeResource() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Actions to take when the rule matches; exactly one of `ids` or `docs` must be set.",
		Required:            true,
		Validators: []validator.Object{
			queryRuleActionsValidator{},
		},
		Attributes: map[string]schema.Attribute{
			queryRuleActionsIDsAttrName: schema.ListAttribute{
				MarkdownDescription: "Document IDs to pin or exclude.",
				Optional:            true,
				ElementType:         attrTypesString,
			},
			queryRuleActionsDocsAttrName: schema.ListNestedAttribute{
				MarkdownDescription: "Documents to pin or exclude, specified by index and ID.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						queryRuleActionDocIndexAttrName: schema.StringAttribute{
							MarkdownDescription: "Index containing the document.",
							Required:            true,
						},
						queryRuleActionDocIDAttrName: schema.StringAttribute{
							MarkdownDescription: "Unique document ID.",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

// dataSourceSchemaFactory returns the schema for the query ruleset data source.
// The elasticsearch_connection block is injected automatically by the envelope.
func dataSourceSchemaFactory(_ context.Context) dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: queryRulesetDataSourceMarkdownDescription,
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				MarkdownDescription: "Internal identifier in the form `<cluster_uuid>/<ruleset_id>`.",
				Computed:            true,
			},
			"ruleset_id": dschema.StringAttribute{
				MarkdownDescription: "Unique identifier of the query ruleset to look up.",
				Required:            true,
			},
			"rules": rulesListNestedAttributeDataSource(),
		},
	}
}

func rulesListNestedAttributeDataSource() dschema.ListNestedAttribute {
	return dschema.ListNestedAttribute{
		MarkdownDescription: "Ordered list of query rules for this ruleset.",
		Computed:            true,
		NestedObject:        queryRuleNestedObjectDataSource(),
	}
}

func queryRuleNestedObjectDataSource() dschema.NestedAttributeObject {
	return dschema.NestedAttributeObject{
		Attributes: map[string]dschema.Attribute{
			queryRuleRuleIDAttrName: dschema.StringAttribute{
				MarkdownDescription: "Unique identifier for the rule within the ruleset.",
				Computed:            true,
			},
			queryRuleCriteriaTypeAttrName: dschema.StringAttribute{
				MarkdownDescription: "Rule type: `pinned` or `exclude`.",
				Computed:            true,
			},
			queryRulePriorityAttrName: dschema.Int64Attribute{
				MarkdownDescription: "Relative priority within the ruleset.",
				Computed:            true,
			},
			queryRuleCriteriaAttrName: queryRuleCriteriaListNestedAttributeDataSource(),
			queryRuleActionsAttrName:  queryRuleActionsSingleNestedAttributeDataSource(),
		},
	}
}

func queryRuleCriteriaListNestedAttributeDataSource() dschema.ListNestedAttribute {
	return dschema.ListNestedAttribute{
		MarkdownDescription: "Match criteria for the rule.",
		Computed:            true,
		NestedObject:        queryRuleCriteriaNestedObjectDataSource(),
	}
}

func queryRuleCriteriaNestedObjectDataSource() dschema.NestedAttributeObject {
	return dschema.NestedAttributeObject{
		Attributes: map[string]dschema.Attribute{
			queryRuleCriteriaTypeAttrName: dschema.StringAttribute{
				MarkdownDescription: "Criteria type.",
				Computed:            true,
			},
			queryRuleCriteriaMetadataAttrName: dschema.StringAttribute{
				MarkdownDescription: "Metadata field matched against.",
				Computed:            true,
			},
			queryRuleCriteriaValuesAttrName: dschema.StringAttribute{
				MarkdownDescription: "JSON-encoded array of string or numeric values.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
		},
	}
}

func queryRuleActionsSingleNestedAttributeDataSource() dschema.SingleNestedAttribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: "Actions taken when the rule matches.",
		Computed:            true,
		Attributes: map[string]dschema.Attribute{
			queryRuleActionsIDsAttrName: dschema.ListAttribute{
				MarkdownDescription: "Document IDs pinned or excluded.",
				Computed:            true,
				ElementType:         attrTypesString,
			},
			queryRuleActionsDocsAttrName: dschema.ListNestedAttribute{
				MarkdownDescription: "Documents pinned or excluded.",
				Computed:            true,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						queryRuleActionDocIndexAttrName: dschema.StringAttribute{
							MarkdownDescription: "Index containing the document.",
							Computed:            true,
						},
						queryRuleActionDocIDAttrName: dschema.StringAttribute{
							MarkdownDescription: "Unique document ID.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}
