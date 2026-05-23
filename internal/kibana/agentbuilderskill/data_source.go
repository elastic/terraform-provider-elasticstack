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

package agentbuilderskill

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = NewDataSource()
	_ datasource.DataSourceWithConfigure = NewDataSource().(datasource.DataSourceWithConfigure)
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return entitycore.NewKibanaDataSource[skillDataSourceModel](
		entitycore.ComponentKibana,
		"agentbuilder_skill",
		getDataSourceSchema,
		readSkillDataSource,
	)
}

// getDataSourceSchema returns the schema for the skill data source. The
// kibana_connection block is injected by the envelope and must not be declared
// here.
func getDataSourceSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		Description:         "Export an Agent Builder skill by ID. See https://www.elastic.co/docs/api/doc/kibana/operation/operation-get-agent-builder-skills-skillid",
		MarkdownDescription: "Export an Agent Builder skill by ID. See the [Agent Builder API documentation](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-agent-builder).",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the skill: `<space_id>/<skill_id>`.",
			},
			"skill_id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The skill ID to look up. Accepts either a bare skill id or a composite `<space_id>/<skill_id>` string.",
			},
			"space_id": dsschema.StringAttribute{
				MarkdownDescription: "An identifier for the Kibana space. If not provided, the default space is used unless the `skill_id` argument supplies a composite space.",
				Optional:            true,
				Computed:            true,
			},
			"name": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Human-readable name for the skill.",
			},
			"description": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of what the skill does.",
			},
			"content": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Skill instructions content as markdown.",
			},
			"tool_ids": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of tool IDs from the tool registry that this skill references.",
			},
			"referenced_content": dsschema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Ordered list of referenced-content entries. Order is preserved as returned by the API.",
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"name": dsschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the referenced content.",
						},
						"relative_path": dsschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Relative path of the referenced content.",
						},
						"content": dsschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Content of the reference.",
						},
					},
				},
			},
		},
	}
}

// readSkillDataSource is the envelope read callback for the skill data source.
// The envelope owns config decode, GetKibanaClient, static version enforcement
// via GetVersionRequirements, and resp.State.Set. This function only contains
// entity-specific logic.
func readSkillDataSource(ctx context.Context, kbClient *clients.KibanaScopedClient, config skillDataSourceModel) (skillDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(config.SkillID) || config.SkillID.ValueString() == "" {
		diags.AddError("Invalid configuration", "skill_id must be set.")
		return config, diags
	}

	oapiClient, d := kbClient.GetKibanaOapiClient()
	diags.Append(d...)
	if diags.HasError() {
		return config, diags
	}

	spaceID := defaultSpaceID
	spaceExplicit := typeutils.IsKnown(config.SpaceID) && config.SpaceID.ValueString() != ""
	if spaceExplicit {
		spaceID = config.SpaceID.ValueString()
	}

	skillID := config.SkillID.ValueString()
	if compID, idDiags := clients.CompositeIDFromStr(skillID); !idDiags.HasError() {
		skillID = compID.ResourceID
		if !spaceExplicit {
			spaceID = compID.ClusterID
		}
	}

	skill, skillDiags := kibanaoapi.GetSkill(ctx, oapiClient, spaceID, skillID)
	diags.Append(skillDiags...)
	if diags.HasError() {
		return config, diags
	}
	if skill == nil {
		diags.AddError("Skill not found", fmt.Sprintf("Unable to fetch skill with ID %s", skillID))
		return config, diags
	}

	populateDiags := (&config).populateFromAPI(ctx, spaceID, skill)
	diags.Append(populateDiags...)
	if diags.HasError() {
		return config, diags
	}

	// Ensure SkillID is normalized back to just the resource id (in case input
	// was composite).
	config.SkillID = types.StringValue(skill.ID)

	return config, diags
}
