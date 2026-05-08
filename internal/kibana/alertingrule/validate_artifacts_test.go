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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestValidateArtifactsInvestigationGuideExclusivity_AllowsUnknownContentPath(t *testing.T) {
	ctx := context.Background()

	igObj, diags := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringUnknown(),
		Checksum:    types.StringUnknown(),
	})
	require.False(t, diags.HasError())

	artifactsObj, diags := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{
		Dashboards:         types.ListNull(types.ObjectType{AttrTypes: getDashboardsAttrTypes()}),
		InvestigationGuide: igObj,
	})
	require.False(t, diags.HasError())

	data := alertingRuleModel{Artifacts: artifactsObj}
	var validationDiags diag.Diagnostics
	validateArtifactsInvestigationGuideExclusivity(ctx, &data, &validationDiags)
	require.False(t, validationDiags.HasError(), "unexpected diagnostics: %v", validationDiags)
}

func TestValidateArtifactsInvestigationGuideExclusivity_RejectsBothSet(t *testing.T) {
	ctx := context.Background()

	igObj, diags := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), investigationGuideModel{
		Content:     types.StringValue("inline"),
		ContentPath: types.StringValue("/tmp/guide.md"),
		Checksum:    types.StringNull(),
	})
	require.False(t, diags.HasError())

	artifactsObj, diags := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{
		Dashboards:         types.ListNull(types.ObjectType{AttrTypes: getDashboardsAttrTypes()}),
		InvestigationGuide: igObj,
	})
	require.False(t, diags.HasError())

	data := alertingRuleModel{Artifacts: artifactsObj}
	var validationDiags diag.Diagnostics
	validateArtifactsInvestigationGuideExclusivity(ctx, &data, &validationDiags)
	require.True(t, validationDiags.HasError())
	require.Contains(t, validationDiags.Errors()[0].Summary(), "Invalid investigation_guide configuration")
	require.Contains(t, validationDiags.Errors()[0].Detail(), "Both are present")
}

func TestValidateArtifactsInvestigationGuideExclusivity_AllowsAbsentInvestigationGuide(t *testing.T) {
	ctx := context.Background()

	artifactsObj, diags := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{
		Dashboards:         types.ListNull(types.ObjectType{AttrTypes: getDashboardsAttrTypes()}),
		InvestigationGuide: types.ObjectNull(getInvestigationGuideAttrTypes()),
	})
	require.False(t, diags.HasError())

	data := alertingRuleModel{Artifacts: artifactsObj}
	var validationDiags diag.Diagnostics
	validateArtifactsInvestigationGuideExclusivity(ctx, &data, &validationDiags)
	require.False(t, validationDiags.HasError(), "unexpected diagnostics: %v", validationDiags)
}
