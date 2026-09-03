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

package fieldstatstable

import (
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const panelType = "field_stats_table"

const (
	attrByDataview = "by_dataview"
	attrByEsql     = "by_esql"
)

//go:embed descriptions/field_stats_table_config.md
var fieldStatsTableConfigDescription string

//go:embed descriptions/by_dataview.md
var fieldStatsTableByDataviewDescription string

//go:embed descriptions/by_esql.md
var fieldStatsTableByEsqlDescription string

// fieldStatsTableConfigModeValidator enforces exactly one of `by_dataview` or `by_esql` inside
// `field_stats_table_config`, deferring validation while either branch value is still unknown.
var fieldStatsTableConfigModeValidator = validators.ExactlyOneOfNestedAttrsValidator(validators.ExactlyOneOfNestedAttrsOpts{
	AttrNames:     []string{attrByDataview, attrByEsql},
	Summary:       "Invalid field_stats_table_config",
	MissingDetail: "Exactly one of `by_dataview` or `by_esql` must be set inside `field_stats_table_config`.",
	TooManyDetail: "Exactly one of `by_dataview` or `by_esql` must be set inside `field_stats_table_config`, not both.",
	Description:   "Ensures exactly one of `by_dataview` or `by_esql` is set inside `field_stats_table_config`.",
})

func fieldStatsTableBranchAttributes() map[string]schema.Attribute {
	attrs := panelkit.PanelPresentationAttributes()
	attrs["show_distributions"] = schema.BoolAttribute{
		MarkdownDescription: "When true, shows distribution mini-charts in the field statistics table. Null-preserved on read (REQ-009).",
		Optional:            true,
	}
	attrs["time_range"] = panelkit.TimeRangeSchema(
		"Optional panel time range override (`from`, `to`, optional `mode`). Null-preserved on read: " +
			"when omitted in configuration, this attribute stays null in state even if Kibana returns values (REQ-009).",
	)
	return attrs
}

func fieldStatsTableByDataviewAttributes() map[string]schema.Attribute {
	attrs := fieldStatsTableBranchAttributes()
	attrs["data_view_id"] = schema.StringAttribute{
		MarkdownDescription: "The identifier of the source data view.",
		Required:            true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	}
	return attrs
}

func fieldStatsTableByEsqlAttributes() map[string]schema.Attribute {
	attrs := fieldStatsTableBranchAttributes()
	attrs["query"] = schema.StringAttribute{
		MarkdownDescription: "The ES|QL query string (mapped to `query.esql` in the API).",
		Required:            true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	}
	return attrs
}

// SchemaAttribute returns the Terraform schema for `field_stats_table_config`.
func SchemaAttribute() schema.Attribute {
	return panelkit.PanelConfigBlock(panelkit.PanelConfigBlockOpts{
		Description: fieldStatsTableConfigDescription,
		BlockName:   "field_stats_table_config",
		PanelType:   panelType,
		Required:    true,
		Attributes: map[string]schema.Attribute{
			attrByDataview: schema.SingleNestedAttribute{
				MarkdownDescription: fieldStatsTableByDataviewDescription,
				Optional:            true,
				Attributes:          fieldStatsTableByDataviewAttributes(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(attrByEsql)),
				},
			},
			attrByEsql: schema.SingleNestedAttribute{
				MarkdownDescription: fieldStatsTableByEsqlDescription,
				Optional:            true,
				Attributes:          fieldStatsTableByEsqlAttributes(),
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(attrByDataview)),
				},
			},
		},
		ExtraValidators: []validator.Object{
			fieldStatsTableConfigModeValidator,
		},
	})
}
