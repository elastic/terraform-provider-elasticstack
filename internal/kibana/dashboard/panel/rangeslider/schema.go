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

package rangeslider

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const panelType = "range_slider_control"

// Branch keys for the by_field/by_esql union, exported so callers outside this package (e.g. the
// dashboard resource's v0->v1 state upgrader) can reference them instead of duplicating the
// literal strings.
const (
	BranchByField = "by_field"
	BranchByEsql  = "by_esql"
)

// SchemaAttribute returns the dashboard panel range_slider_control_config block. Exactly one of the
// `by_field` (data view field) or `by_esql` (ES|QL query) nested blocks must be set.
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: "Configuration for a range slider control panel. Provides a min/max range filter sourced from " +
			"either a data view field (`by_field`) or an ES|QL query (`by_esql`). Exactly one of the two must be set.",
		BlockName:  "range_slider_control_config",
		PanelType:  panelType,
		Required:   true,
		Attributes: NestedAttributes(),
		ExtraValidators: []validator.Object{
			ExactlyOneOfBranchValidator(),
		},
	})
}

// ExactlyOneOfBranchValidator enforces that exactly one of `by_field` / `by_esql` is configured
// inside a block using NestedAttributes(). Shared by the regular panel schema and the pinned-panel
// control-bar schema.
func ExactlyOneOfBranchValidator() validator.Object {
	return validators.ExactlyOneOfNestedAttrsValidator(validators.ExactlyOneOfNestedAttrsOpts{
		AttrNames:     []string{BranchByField, BranchByEsql},
		Summary:       "Invalid range_slider_control_config",
		MissingDetail: "Exactly one of `by_field` or `by_esql` must be configured inside `range_slider_control_config`.",
		TooManyDetail: "Exactly one of `by_field` or `by_esql` must be configured inside `range_slider_control_config`, not both.",
		Description:   "Ensures exactly one of `by_field` or `by_esql` is configured inside `range_slider_control_config`.",
	})
}

// NestedAttributes returns the `by_field` / `by_esql` branch attribute map shared by the regular
// panel schema (SchemaAttribute) and the pinned-panel control-bar schema.
func NestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		BranchByField: schema.SingleNestedAttribute{
			MarkdownDescription: "Range slider sourced from a Kibana data view field. Mutually exclusive with `by_esql`.",
			Optional:            true,
			Attributes:          byFieldAttributes(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(BranchByEsql)),
			},
		},
		BranchByEsql: schema.SingleNestedAttribute{
			MarkdownDescription: "Range slider sourced from an ES|QL query. Mutually exclusive with `by_field`.",
			Optional:            true,
			Attributes:          byEsqlAttributes(),
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(BranchByField)),
			},
		},
	}
}

// sharedAttributes returns the attributes common to both `by_field` and `by_esql`.
func sharedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"title": schema.StringAttribute{
			MarkdownDescription: "A human-readable title for the control.",
			Optional:            true,
		},
		"use_global_filters": schema.BoolAttribute{
			MarkdownDescription: "Whether the control respects dashboard-level filters.",
			Optional:            true,
		},
		"ignore_validations": schema.BoolAttribute{
			MarkdownDescription: "Whether to suppress validation errors during intermediate states.",
			Optional:            true,
		},
		"value": schema.ListAttribute{
			MarkdownDescription: "Initial range as a list of exactly 2 strings: [min, max].",
			ElementType:         types.StringType,
			Optional:            true,
			Validators: []validator.List{
				listvalidator.SizeAtLeast(2),
				listvalidator.SizeAtMost(2),
			},
		},
		"step": schema.Float32Attribute{
			MarkdownDescription: "The step size for the range slider. Stored as float32 to match the Kibana API type and avoid refresh drift.",
			Optional:            true,
		},
	}
}

func byFieldAttributes() map[string]schema.Attribute {
	attrs := sharedAttributes()
	attrs["data_view_id"] = schema.StringAttribute{
		MarkdownDescription: "The ID of the data view that the control is tied to.",
		Required:            true,
	}
	attrs["field_name"] = schema.StringAttribute{
		MarkdownDescription: "The name of the field in the data view that the control is tied to.",
		Required:            true,
	}
	return attrs
}

// ByFieldAttributeNames returns the by_field branch's attribute names. Callers that need to
// enumerate them without depending on the full schema (e.g. the dashboard resource's v0->v1 state
// upgrader, which relocates these same names from a flat v0 layout into by_field {}) should use
// this instead of hardcoding a duplicate list that could drift from the schema.
func ByFieldAttributeNames() []string {
	attrs := byFieldAttributes()
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	return names
}

func byEsqlAttributes() map[string]schema.Attribute {
	attrs := sharedAttributes()
	attrs["esql_query"] = schema.StringAttribute{
		MarkdownDescription: "The ES|QL query that produces the min/max range values.",
		Required:            true,
	}
	attrs["values_source"] = schema.StringAttribute{
		MarkdownDescription: "The source of the range values. Must be `esql_query`.",
		Required:            true,
		Validators: []validator.String{
			stringvalidator.OneOf(panelkit.EsqlValuesSourceUserValue),
		},
	}
	return attrs
}
