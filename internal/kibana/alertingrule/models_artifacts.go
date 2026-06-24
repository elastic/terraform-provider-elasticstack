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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const (
	attrInvestigationGuideContent     = "content"
	attrInvestigationGuideContentPath = "content_path"
	attrInvestigationGuideChecksum    = "checksum"
)

// artifactsModel is the Terraform model for rule-level artifacts.
type artifactsModel struct {
	Dashboards         types.List   `tfsdk:"dashboards"`
	InvestigationGuide types.Object `tfsdk:"investigation_guide"`
}

type artifactDashboardModel struct {
	ID types.String `tfsdk:"id"`
}

func investigationGuideAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrInvestigationGuideContent:     types.StringType,
		attrInvestigationGuideContentPath: types.StringType,
		attrInvestigationGuideChecksum:    types.StringType,
	}
}

type investigationGuideModel struct {
	Content     types.String `tfsdk:"content"`
	ContentPath types.String `tfsdk:"content_path"`
	Checksum    types.String `tfsdk:"checksum"`
}

var (
	artifactsMinSupportedVersion8 = version.Must(version.NewVersion("8.19.0"))
	artifactsMinSupportedVersion9 = version.Must(version.NewVersion("9.1.0"))
)

func artifactsVersionSupported(sv *version.Version) bool {
	if sv == nil {
		return false
	}
	segments := sv.Segments()
	if len(segments) > 0 && segments[0] >= 9 {
		return sv.GreaterThanOrEqual(artifactsMinSupportedVersion9)
	}
	return sv.GreaterThanOrEqual(artifactsMinSupportedVersion8)
}

// ArtifactsVersionSupported reports whether the connected Kibana version supports
// configuring alerting rule artifacts.
func ArtifactsVersionSupported(sv *version.Version) bool {
	return artifactsVersionSupported(sv)
}

func enforceArtifactsVersion(ctx context.Context, client *clients.KibanaScopedClient, m alertingRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(m.Artifacts) || m.Artifacts.IsNull() {
		return diags
	}
	supported, vDiags := client.EnforceVersionCheck(ctx, artifactsVersionSupported)
	diags.Append(vDiags...)
	if diags.HasError() {
		return diags
	}
	if !supported {
		diags.AddError(
			"Unsupported server version",
			"artifacts is only supported for Kibana 8.19 or higher on the 8.x line and 9.1 or higher on the 9.x line",
		)
	}
	return diags
}

func (m *alertingRuleModel) populateArtifactsFromAPI(ctx context.Context, rule *models.AlertingRule) diag.Diagnostics {
	var diags diag.Diagnostics

	if rule.Artifacts == nil {
		if typeutils.IsKnown(m.Artifacts) && !m.Artifacts.IsNull() {
			diags.Append(finalizeArtifactsChecksumInModel(ctx, m)...)
			return diags
		}
		if m.Artifacts.IsUnknown() {
			m.Artifacts = types.ObjectNull(getArtifactsAttrTypes())
		}
		return diags
	}

	if !artifactsAPIHasContent(rule.Artifacts) {
		if typeutils.IsKnown(m.Artifacts) && !m.Artifacts.IsNull() {
			diags.Append(finalizeArtifactsChecksumInModel(ctx, m)...)
			return diags
		}
		m.Artifacts = types.ObjectNull(getArtifactsAttrTypes())
		return diags
	}

	var prior artifactsModel
	if typeutils.IsKnown(m.Artifacts) && !m.Artifacts.IsNull() {
		diags.Append(m.Artifacts.As(ctx, &prior, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return diags
		}
	}

	priorIG := investigationGuideModel{}
	if typeutils.IsKnown(prior.InvestigationGuide) && !prior.InvestigationGuide.IsNull() {
		diags.Append(prior.InvestigationGuide.As(ctx, &priorIG, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return diags
		}
	}

	out := artifactsModel{
		Dashboards:         types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType}}),
		InvestigationGuide: types.ObjectNull(investigationGuideAttrTypes()),
	}

	if len(rule.Artifacts.Dashboards) > 0 {
		rows := make([]artifactDashboardModel, 0, len(rule.Artifacts.Dashboards))
		for _, d := range rule.Artifacts.Dashboards {
			rows = append(rows, artifactDashboardModel{ID: types.StringValue(d.ID)})
		}
		dl, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType}}, rows)
		diags.Append(d...)
		out.Dashboards = dl
	} else if typeutils.IsKnown(prior.Dashboards) && !prior.Dashboards.IsNull() {
		out.Dashboards = prior.Dashboards
	}

	if rule.Artifacts.InvestigationGuide != nil && rule.Artifacts.InvestigationGuide.Blob != "" {
		ig := investigationGuideModel{}
		if typeutils.IsKnown(priorIG.ContentPath) && !priorIG.ContentPath.IsNull() {
			ig.ContentPath = priorIG.ContentPath
			ig.Content = types.StringNull()
			ig.Checksum = priorIG.Checksum
		} else {
			ig.Content = types.StringValue(rule.Artifacts.InvestigationGuide.Blob)
			ig.ContentPath = types.StringNull()
			ig.Checksum = types.StringNull()
		}
		igObj, d := types.ObjectValueFrom(ctx, investigationGuideAttrTypes(), ig)
		diags.Append(d...)
		out.InvestigationGuide = igObj
	} else if typeutils.IsKnown(prior.InvestigationGuide) && !prior.InvestigationGuide.IsNull() {
		out.InvestigationGuide = prior.InvestigationGuide
	}

	obj, d := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), out)
	diags.Append(d...)
	m.Artifacts = obj
	diags.Append(finalizeArtifactsChecksumInModel(ctx, m)...)
	return diags
}

func artifactsToAPIModel(ctx context.Context, obj types.Object) (*models.AlertingRuleArtifacts, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(obj) || obj.IsNull() {
		return nil, diags
	}

	var am artifactsModel
	diags.Append(obj.As(ctx, &am, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	out := &models.AlertingRuleArtifacts{}

	if typeutils.IsKnown(am.Dashboards) && !am.Dashboards.IsNull() {
		var rows []artifactDashboardModel
		diags.Append(am.Dashboards.ElementsAs(ctx, &rows, false)...)
		if diags.HasError() {
			return nil, diags
		}
		for _, row := range rows {
			out.Dashboards = append(out.Dashboards, models.AlertingRuleArtifactDashboard{
				ID: row.ID.ValueString(),
			})
		}
	}

	if typeutils.IsKnown(am.InvestigationGuide) && !am.InvestigationGuide.IsNull() {
		var ig investigationGuideModel
		diags.Append(am.InvestigationGuide.As(ctx, &ig, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		blob, blobDiags := investigationGuideBlobFromModel(ig)
		diags.Append(blobDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if blob != "" {
			out.InvestigationGuide = &models.AlertingRuleArtifactInvestigationGuide{Blob: blob}
		}
	}

	return out, diags
}

func investigationGuideBlobFromModel(ig investigationGuideModel) (string, diag.Diagnostics) {
	var diags diag.Diagnostics

	if typeutils.IsKnown(ig.Content) && !ig.Content.IsNull() {
		return ig.Content.ValueString(), diags
	}

	if typeutils.IsKnown(ig.ContentPath) && !ig.ContentPath.IsNull() {
		filePath := ig.ContentPath.ValueString()
		data, err := os.ReadFile(filePath)
		if err != nil {
			diags.AddAttributeError(
				fwpath.Root("artifacts").AtName("investigation_guide").AtName(attrInvestigationGuideContentPath),
				"Cannot read investigation guide file",
				fmt.Sprintf("Failed to read content_path %q: %v", filePath, err),
			)
			return "", diags
		}
		return string(data), diags
	}

	return "", diags
}

func applyArtifactsChecksumToModel(ctx context.Context, m *alertingRuleModel) diag.Diagnostics {
	return finalizeArtifactsChecksumInModel(ctx, m)
}

func finalizeArtifactsChecksumInModel(ctx context.Context, m *alertingRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if !typeutils.IsKnown(m.Artifacts) || m.Artifacts.IsNull() {
		return diags
	}

	var am artifactsModel
	diags.Append(m.Artifacts.As(ctx, &am, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return diags
	}
	if !typeutils.IsKnown(am.InvestigationGuide) || am.InvestigationGuide.IsNull() {
		return diags
	}

	var ig investigationGuideModel
	diags.Append(am.InvestigationGuide.As(ctx, &ig, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return diags
	}

	if typeutils.IsKnown(ig.ContentPath) && !ig.ContentPath.IsNull() {
		checksum, err := sha256HexFile(ig.ContentPath.ValueString())
		if err != nil {
			diags.AddAttributeError(
				fwpath.Root("artifacts").AtName("investigation_guide").AtName(attrInvestigationGuideContentPath),
				"Cannot read investigation guide file",
				err.Error(),
			)
			return diags
		}
		ig.Checksum = types.StringValue(checksum)
	} else {
		ig.Checksum = types.StringNull()
	}

	igObj, d := types.ObjectValueFrom(ctx, investigationGuideAttrTypes(), ig)
	diags.Append(d...)
	am.InvestigationGuide = igObj

	obj, d := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), am)
	diags.Append(d...)
	m.Artifacts = obj
	return diags
}

func sha256HexFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open %q: %w", filePath, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to read %q: %w", filePath, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func artifactsAPIHasContent(a *models.AlertingRuleArtifacts) bool {
	if a == nil {
		return false
	}
	if len(a.Dashboards) > 0 {
		return true
	}
	return a.InvestigationGuide != nil && a.InvestigationGuide.Blob != ""
}
