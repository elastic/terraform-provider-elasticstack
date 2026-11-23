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

package alertingrule

import (
	"context"
	"fmt"
	"os"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/fileutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// artifactsModel is the Terraform model for the rule's linked artifacts.
type artifactsModel struct {
	Dashboards         types.List   `tfsdk:"dashboards"`
	InvestigationGuide types.Object `tfsdk:"investigation_guide"`
}

// investigationGuideModel is the Terraform model for an investigation guide.
type investigationGuideModel struct {
	Content     types.String `tfsdk:"content"`
	ContentPath types.String `tfsdk:"content_path"`
	Checksum    types.String `tfsdk:"checksum"`
}

// dashboardModel is the Terraform model for a single linked dashboard.
type dashboardModel struct {
	ID types.String `tfsdk:"id"`
}

// artifactsModelFrom decodes the full artifacts nested object from m.Artifacts
// (from plan or state, depending on the caller), or nil when unset/unknown.
func (m alertingRuleModel) artifactsModelFrom(ctx context.Context) (*artifactsModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	am := typeutils.ObjectTypeAs[artifactsModel](ctx, m.Artifacts, path.Root("artifacts"), &diags)
	if diags.HasError() {
		return nil, diags
	}
	return am, diags
}

// investigationGuideFrom returns the investigation guide nested under
// m.Artifacts (from plan or state, depending on the caller), or nil when unset.
func (m alertingRuleModel) investigationGuideFrom(ctx context.Context) (*investigationGuideModel, diag.Diagnostics) {
	am, diags := m.artifactsModelFrom(ctx)
	if diags.HasError() || am == nil {
		return nil, diags
	}
	ig := investigationGuideFromArtifacts(ctx, am, &diags)
	return ig, diags
}

// investigationGuideFromArtifacts decodes the investigation_guide nested object
// from an already-decoded artifactsModel, or nil when unset.
func investigationGuideFromArtifacts(ctx context.Context, am *artifactsModel, diags *diag.Diagnostics) *investigationGuideModel {
	if am == nil {
		return nil
	}
	ig := typeutils.ObjectTypeAs[investigationGuideModel](ctx, am.InvestigationGuide, path.Root("artifacts").AtName("investigation_guide"), diags)
	if diags.HasError() {
		return nil
	}
	return ig
}

// buildArtifactsObject builds the artifacts nested object from an artifactsModel,
// ensuring every field (dashboards + investigation_guide) is preserved. Callers
// that only touch one field must read the current artifactsModel first so the
// other field is carried through unchanged.
func buildArtifactsObject(ctx context.Context, am artifactsModel) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), am)
}

// resolveArtifactsUnknowns replaces any unknown nested artifacts attribute with
// its null value. Because `investigation_guide` and `dashboards` are both
// Optional+Computed, the one the practitioner did not configure is unknown in
// the plan; on stacks that do not return artifacts on read the value would
// otherwise remain unknown in state after apply.
func resolveArtifactsUnknowns(am artifactsModel) artifactsModel {
	if am.InvestigationGuide.IsUnknown() {
		am.InvestigationGuide = types.ObjectNull(getInvestigationGuideAttrTypes())
	}
	if am.Dashboards.IsUnknown() {
		am.Dashboards = types.ListNull(getDashboardsElementType())
	}
	return am
}

// populateArtifactsFromAPI maps the API artifacts (investigation guide and
// dashboards) back into the model while preserving the practitioner's chosen
// investigation-guide source. The API returns only the blob, never a checksum,
// so for the file-based (`content_path`) case the prior path and checksum are
// preserved and the blob is not surfaced as `content`.
//
// On stacks older than 9.5.0 the Kibana GET API does not return artifacts
// (elastic/kibana#247279); when the entire artifacts key is omitted the
// prior (plan/state) value is preserved so write-only management does not
// thrash state. When the API does return artifacts, omitted siblings are
// cleared so out-of-band deletion of just the guide or just dashboards is
// visible on refresh.
func (m *alertingRuleModel) populateArtifactsFromAPI(ctx context.Context, rule *models.AlertingRule) diag.Diagnostics {
	var diags diag.Diagnostics

	artifactsAttrTypes := getArtifactsAttrTypes()

	prior, d := m.artifactsModelFrom(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	priorIG := investigationGuideFromArtifacts(ctx, prior, &diags)
	if diags.HasError() {
		return diags
	}

	if rule.Artifacts == nil {
		// Nothing returned by the API (nothing configured, or a pre-9.5.0 stack
		// that does not return artifacts). Preserve configured values only when
		// the entire artifacts key is omitted — not per-field — so that on
		// 9.5.0+ an out-of-band deletion of just the guide or just dashboards
		// is visible on refresh.
		if prior == nil {
			// Nothing configured: resolve a lingering unknown object to null.
			if m.Artifacts.IsUnknown() {
				m.Artifacts = types.ObjectNull(artifactsAttrTypes)
			}
			return diags
		}
		// Artifacts were configured but the API returned none (write-only on
		// pre-9.5.0 stacks). Preserve the configured values, but resolve any
		// nested attribute the practitioner did not set (which is Computed and
		// therefore unknown in the plan) to null so state has no unknown after
		// apply.
		resolved := resolveArtifactsUnknowns(*prior)
		artObj, d := buildArtifactsObject(ctx, resolved)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.Artifacts = artObj
		return diags
	}

	am := artifactsModel{
		InvestigationGuide: types.ObjectNull(getInvestigationGuideAttrTypes()),
		Dashboards:         types.ListNull(getDashboardsElementType()),
	}

	// Investigation guide. Omitted sibling stays null (do not preserve prior).
	if rule.Artifacts.InvestigationGuide != nil {
		blob := rule.Artifacts.InvestigationGuide.Blob
		ig := investigationGuideModel{}
		if priorIG != nil && typeutils.IsKnown(priorIG.ContentPath) && !priorIG.ContentPath.IsNull() {
			// File-based source: preserve path + checksum, do not surface blob.
			ig.Content = types.StringNull()
			ig.ContentPath = priorIG.ContentPath
			ig.Checksum = priorIG.Checksum
		} else {
			// Inline content (or first read / import): store the blob as content.
			ig.Content = types.StringValue(blob)
			ig.ContentPath = types.StringNull()
			ig.Checksum = types.StringNull()
		}
		igObj, d := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), ig)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		am.InvestigationGuide = igObj
	}

	// Dashboards. Omitted sibling stays null (do not preserve prior).
	if len(rule.Artifacts.Dashboards) > 0 {
		dashboards := make([]dashboardModel, len(rule.Artifacts.Dashboards))
		for i, dash := range rule.Artifacts.Dashboards {
			dashboards[i] = dashboardModel{ID: types.StringValue(dash.ID)}
		}
		dashList, d := types.ListValueFrom(ctx, getDashboardsElementType(), dashboards)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		am.Dashboards = dashList
	}

	artObj, d := buildArtifactsObject(ctx, am)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	m.Artifacts = artObj
	return diags
}

// artifactsToAPI converts the model's artifacts object into the API model. For
// the file-based investigation guide the file at `content_path` is read at
// apply time and its contents are sent as the blob. Returns nil when neither an
// investigation guide nor any dashboards are configured.
func (m alertingRuleModel) artifactsToAPI(ctx context.Context) (*models.AlertingRuleArtifacts, diag.Diagnostics) {
	am, diags := m.artifactsModelFrom(ctx)
	if diags.HasError() || am == nil {
		return nil, diags
	}

	out := &models.AlertingRuleArtifacts{}

	if ig := investigationGuideFromArtifacts(ctx, am, &diags); ig != nil {
		if diags.HasError() {
			return nil, diags
		}
		var blob string
		switch {
		case typeutils.IsKnown(ig.Content) && !ig.Content.IsNull():
			blob = ig.Content.ValueString()
		case typeutils.IsKnown(ig.ContentPath) && !ig.ContentPath.IsNull():
			data, err := os.ReadFile(ig.ContentPath.ValueString())
			if err != nil {
				diags.AddError(
					"Cannot read investigation guide file",
					fmt.Sprintf("Failed to read content_path %q: %s", ig.ContentPath.ValueString(), err),
				)
				return nil, diags
			}
			blob = string(data)
		}
		// Send the guide whenever it is configured, including an explicit
		// empty `content = ""`. Dropping empty blobs omitted the artifacts
		// key entirely and produced a silent no-op (or inconsistent-result
		// after apply) instead of writing the empty guide.
		out.InvestigationGuide = &models.AlertingRuleInvestigationGuide{Blob: blob}
	}

	if typeutils.IsKnown(am.Dashboards) && !am.Dashboards.IsNull() {
		var dashboards []dashboardModel
		diags.Append(am.Dashboards.ElementsAs(ctx, &dashboards, false)...)
		if diags.HasError() {
			return nil, diags
		}
		for _, d := range dashboards {
			out.Dashboards = append(out.Dashboards, models.AlertingRuleArtifactDashboard{ID: d.ID.ValueString()})
		}
	}

	if out.InvestigationGuide == nil && len(out.Dashboards) == 0 {
		return nil, diags
	}
	return out, diags
}

// applyInvestigationGuideChecksum recomputes and stores the SHA-256 checksum of
// the file referenced by `content_path`. For the inline `content` case the
// checksum is cleared. It is invoked from Create/Update after a successful API
// write so the computed `checksum` in state reflects the applied file, and
// mirrors the drift detection performed by ModifyPlan.
func (m *alertingRuleModel) applyInvestigationGuideChecksum(ctx context.Context) diag.Diagnostics {
	am, diags := m.artifactsModelFrom(ctx)
	if diags.HasError() || am == nil {
		return diags
	}
	ig := investigationGuideFromArtifacts(ctx, am, &diags)
	if diags.HasError() || ig == nil {
		return diags
	}

	if typeutils.IsKnown(ig.ContentPath) && !ig.ContentPath.IsNull() {
		sum, err := fileutil.SHA256HexDigest(ig.ContentPath.ValueString())
		if err != nil {
			diags.AddError(
				"Cannot read investigation guide file",
				fmt.Sprintf("Failed to compute checksum of content_path %q: %s", ig.ContentPath.ValueString(), err),
			)
			return diags
		}
		ig.Checksum = types.StringValue(sum)
	} else {
		ig.Checksum = types.StringNull()
	}

	igObj, d := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), *ig)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	am.InvestigationGuide = igObj
	artObj, d := buildArtifactsObject(ctx, *am)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.Artifacts = artObj
	return diags
}
