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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type skillModel struct {
	ID                types.String                 `tfsdk:"id"`
	KibanaConnection  types.List                   `tfsdk:"kibana_connection"`
	SkillID           types.String                 `tfsdk:"skill_id"`
	SpaceID           types.String                 `tfsdk:"space_id"`
	Name              types.String                 `tfsdk:"name"`
	Description       types.String                 `tfsdk:"description"`
	Content           types.String                 `tfsdk:"content"`
	ToolIDs           types.Set                    `tfsdk:"tool_ids"`
	ReferencedContent []skillReferencedContentItem `tfsdk:"referenced_content"`
}

type skillDataSourceModel struct {
	entitycore.KibanaConnectionField
	ID                types.String                 `tfsdk:"id"`
	SkillID           types.String                 `tfsdk:"skill_id"`
	SpaceID           types.String                 `tfsdk:"space_id"`
	Name              types.String                 `tfsdk:"name"`
	Description       types.String                 `tfsdk:"description"`
	Content           types.String                 `tfsdk:"content"`
	ToolIDs           types.Set                    `tfsdk:"tool_ids"`
	ReferencedContent []skillReferencedContentItem `tfsdk:"referenced_content"`
}

type skillReferencedContentItem struct {
	Name         types.String `tfsdk:"name"`
	RelativePath types.String `tfsdk:"relative_path"`
	Content      types.String `tfsdk:"content"`
}

// GetVersionRequirements returns the static minimum Kibana version requirements
// for the Agent Builder skill data source. This satisfies the optional
// entitycore.WithVersionRequirements interface, allowing the generic Kibana
// data source envelope to enforce the requirement before invoking the entity
// read callback.
func (model skillDataSourceModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *minKibanaAgentBuilderSkillsAPIVersion,
			ErrorMessage: fmt.Sprintf("Agent Builder skills require Elastic Stack v%s or later.", minKibanaAgentBuilderSkillsAPIVersion),
		},
	}, nil
}

func (model *skillModel) populateFromAPI(ctx context.Context, spaceID string, data *models.Skill) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	if spaceID == "" {
		spaceID = defaultSpaceID
	}

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String())
	model.SkillID = types.StringValue(data.ID)
	model.SpaceID = types.StringValue(spaceID)
	model.Name = types.StringValue(data.Name)
	model.Description = types.StringValue(data.Description)
	model.Content = types.StringValue(data.Content)

	diags.Append(populateStringSet(ctx, data.ToolIDs, &model.ToolIDs)...)
	model.ReferencedContent = referencedContentItemsFromAPI(data.ReferencedContent)

	return diags
}

func (model *skillDataSourceModel) populateFromAPI(ctx context.Context, spaceID string, data *models.Skill) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	if spaceID == "" {
		spaceID = defaultSpaceID
	}

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: data.ID}).String())
	model.SkillID = types.StringValue(data.ID)
	model.SpaceID = types.StringValue(spaceID)
	model.Name = types.StringValue(data.Name)
	model.Description = types.StringValue(data.Description)
	model.Content = types.StringValue(data.Content)

	diags.Append(populateStringSet(ctx, data.ToolIDs, &model.ToolIDs)...)
	model.ReferencedContent = referencedContentItemsFromAPI(data.ReferencedContent)

	return diags
}

// referencedContentItemsFromAPI converts API referenced-content entries into
// TF model rows, preserving order. Returns nil when the input is empty so the
// attribute is stored as null in state.
func referencedContentItemsFromAPI(in []models.SkillReferencedContent) []skillReferencedContentItem {
	if len(in) == 0 {
		return nil
	}
	out := make([]skillReferencedContentItem, 0, len(in))
	for _, entry := range in {
		out = append(out, skillReferencedContentItem{
			Name:         types.StringValue(entry.Name),
			RelativePath: types.StringValue(entry.RelativePath),
			Content:      types.StringValue(entry.Content),
		})
	}
	return out
}

func populateStringSet(ctx context.Context, src []string, dst *types.Set) diag.Diagnostics {
	if len(src) > 0 {
		v, d := types.SetValueFrom(ctx, types.StringType, src)
		*dst = v
		return d
	}
	*dst = types.SetNull(types.StringType)
	return nil
}

func setToStrings(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	if set.IsNull() || set.IsUnknown() {
		return nil, nil
	}
	var out []string
	d := set.ElementsAs(ctx, &out, false)
	return out, d
}

// postReferencedContentItem matches the anonymous struct used by
// kbapi.PostAgentBuilderSkillsJSONBody.ReferencedContent.
type postReferencedContentItem = struct {
	Content      string `json:"content"`
	Name         string `json:"name"`
	RelativePath string `json:"relativePath"`
}

// putReferencedContentItem matches the anonymous struct used by
// kbapi.PutAgentBuilderSkillsSkillidJSONBody.ReferencedContent. The shape is
// identical to the POST variant, but the generated types use distinct anonymous
// structs so we keep both aliases for clarity.
type putReferencedContentItem = struct {
	Content      string `json:"content"`
	Name         string `json:"name"`
	RelativePath string `json:"relativePath"`
}

func (model skillModel) toAPICreateModel(ctx context.Context) (kbapi.PostAgentBuilderSkillsJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PostAgentBuilderSkillsJSONRequestBody{
		Id:          model.SkillID.ValueString(),
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Content:     model.Content.ValueString(),
	}

	toolIDs, d := setToStrings(ctx, model.ToolIDs)
	diags.Append(d...)
	if len(toolIDs) > 0 {
		body.ToolIds = &toolIDs
	}

	if len(model.ReferencedContent) > 0 {
		entries := make([]postReferencedContentItem, 0, len(model.ReferencedContent))
		for _, entry := range model.ReferencedContent {
			entries = append(entries, postReferencedContentItem{
				Content:      entry.Content.ValueString(),
				Name:         entry.Name.ValueString(),
				RelativePath: entry.RelativePath.ValueString(),
			})
		}
		body.ReferencedContent = &entries
	}

	return body, diags
}

func (model skillModel) toAPIUpdateModel(ctx context.Context) (kbapi.PutAgentBuilderSkillsSkillidJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := model.Name.ValueString()
	description := model.Description.ValueString()
	content := model.Content.ValueString()

	body := kbapi.PutAgentBuilderSkillsSkillidJSONRequestBody{
		Name:        &name,
		Description: &description,
		Content:     &content,
	}

	toolIDs, d := setToStrings(ctx, model.ToolIDs)
	diags.Append(d...)
	// Always send tool_ids on update (including empty) so cleared values are
	// propagated to Kibana. The omitempty tag means a nil slice would skip the
	// field; we explicitly allocate an empty slice when the model is null.
	if toolIDs == nil {
		toolIDs = []string{}
	}
	body.ToolIds = &toolIDs

	entries := make([]putReferencedContentItem, 0, len(model.ReferencedContent))
	for _, entry := range model.ReferencedContent {
		entries = append(entries, putReferencedContentItem{
			Content:      entry.Content.ValueString(),
			Name:         entry.Name.ValueString(),
			RelativePath: entry.RelativePath.ValueString(),
		})
	}
	body.ReferencedContent = &entries

	return body, diags
}
