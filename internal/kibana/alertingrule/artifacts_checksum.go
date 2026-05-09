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
	"os"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func persistArtifactsChecksum(ctx context.Context, model *alertingRuleModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(model.Artifacts) || model.Artifacts.IsNull() {
		return diags
	}

	var am artifactsModel
	diags.Append(model.Artifacts.As(ctx, &am, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return diags
	}

	if !typeutils.IsKnown(am.InvestigationGuide) || am.InvestigationGuide.IsNull() {
		return diags
	}

	var igm investigationGuideModel
	diags.Append(am.InvestigationGuide.As(ctx, &igm, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)
	if diags.HasError() {
		return diags
	}

	if igm.ContentPath.IsUnknown() {
		return diags
	}

	if igm.ContentPath.IsNull() || igm.ContentPath.ValueString() == "" {
		if !igm.Checksum.IsUnknown() {
			return diags
		}

		igm.Checksum = types.StringNull()
		igObj, d := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), igm)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		am.InvestigationGuide = igObj
		artifactsObj, d := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), am)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		model.Artifacts = artifactsObj
		return diags
	}

	content, err := os.ReadFile(igm.ContentPath.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("artifacts").AtName("investigation_guide").AtName("content_path"),
			"Failed to read investigation guide file",
			err.Error(),
		)
		return diags
	}

	sum := sha256.Sum256(content)
	igm.Checksum = types.StringValue(hex.EncodeToString(sum[:]))

	igObj, d := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), igm)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	am.InvestigationGuide = igObj
	artifactsObj, d := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), am)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	model.Artifacts = artifactsObj
	return diags
}
