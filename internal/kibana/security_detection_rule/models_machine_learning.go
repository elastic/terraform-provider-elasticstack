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

package securitydetectionrule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MachineLearningRuleProcessor struct {
	baseRuleProcessor[kbapi.SecurityDetectionsAPIMachineLearningRule]
}

func newMachineLearningRuleProcessor() MachineLearningRuleProcessor {
	return MachineLearningRuleProcessor{
		updateFn: func(ctx context.Context, v *kbapi.SecurityDetectionsAPIMachineLearningRule, d *Data) diag.Diagnostics {
			return d.updateFromMachineLearningRule(ctx, v)
		},
		idFn: func(v kbapi.SecurityDetectionsAPIMachineLearningRule) string {
			return v.Id.String()
		},
	}
}

func (m MachineLearningRuleProcessor) HandlesRuleType(t string) bool {
	return t == ruleTypeMachineLearning
}

func (m MachineLearningRuleProcessor) ToCreateProps(
	ctx context.Context,
	d Data,
) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	return d.toMachineLearningRuleCreateProps(ctx)
}

func (m MachineLearningRuleProcessor) ToUpdateProps(
	ctx context.Context,
	d Data,
) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	return d.toMachineLearningRuleUpdateProps(ctx)
}

func (d Data) toMachineLearningRuleCreateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleCreateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createProps kbapi.SecurityDetectionsAPIRuleCreateProps

	mlRule := kbapi.SecurityDetectionsAPIMachineLearningRuleCreateProps{
		Name:             d.Name.ValueString(),
		Description:      d.Description.ValueString(),
		Type:             kbapi.SecurityDetectionsAPIMachineLearningRuleCreatePropsType(ruleTypeMachineLearning),
		AnomalyThreshold: kbapi.SecurityDetectionsAPIAnomalyThreshold(d.AnomalyThreshold.ValueInt64()),
		RiskScore:        kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:         kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// Set ML job ID(s) - can be single string or array
	if typeutils.IsKnown(d.MachineLearningJobID) {
		jobIDs := typeutils.ListTypeAs[string](ctx, d.MachineLearningJobID, path.Root("machine_learning_job_id"), &diags)
		if !diags.HasError() {
			var mlJobID kbapi.SecurityDetectionsAPIMachineLearningJobId
			err := mlJobID.FromSecurityDetectionsAPIMachineLearningJobId1(jobIDs)
			if err != nil {
				diags.AddError("Error setting ML job IDs", err.Error())
			} else {
				mlRule.MachineLearningJobId = mlJobID
			}
		}
	}

	// ML rules don't use index patterns or query, and don't support DataViewID or Filters;
	// buildCommonRuleProps leaves those nil automatically since mlRule has no such fields.
	d.setCommonCreateProps(ctx, buildCommonRuleProps(&mlRule), &diags)

	// Convert to union type
	err := createProps.FromSecurityDetectionsAPIMachineLearningRuleCreateProps(mlRule)
	if err != nil {
		diags.AddError(
			"Error building create properties",
			"Could not convert ML rule properties: "+err.Error(),
		)
	}

	return createProps, diags
}
func (d Data) toMachineLearningRuleUpdateProps(ctx context.Context) (kbapi.SecurityDetectionsAPIRuleUpdateProps, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateProps kbapi.SecurityDetectionsAPIRuleUpdateProps

	uid, ok := d.parseResourceUUID(&diags)
	if !ok {
		return updateProps, diags
	}

	mlRule := kbapi.SecurityDetectionsAPIMachineLearningRuleUpdateProps{
		Id:               &uid,
		Name:             d.Name.ValueString(),
		Description:      d.Description.ValueString(),
		Type:             kbapi.SecurityDetectionsAPIMachineLearningRuleUpdatePropsType(ruleTypeMachineLearning),
		AnomalyThreshold: kbapi.SecurityDetectionsAPIAnomalyThreshold(d.AnomalyThreshold.ValueInt64()),
		RiskScore:        kbapi.SecurityDetectionsAPIRiskScore(d.RiskScore.ValueInt64()),
		Severity:         kbapi.SecurityDetectionsAPISeverity(d.Severity.ValueString()),
	}

	// For updates, we need to include the rule_id if it's set
	if typeutils.IsKnown(d.RuleID) {
		ruleID := d.RuleID.ValueString()
		mlRule.RuleId = &ruleID
		mlRule.Id = nil // if rule_id is set, we cant send id
	}

	// Set ML job ID(s) - can be single string or array
	if typeutils.IsKnown(d.MachineLearningJobID) {
		jobIDs := typeutils.ListTypeAs[string](ctx, d.MachineLearningJobID, path.Root("machine_learning_job_id"), &diags)
		if !diags.HasError() {
			var mlJobID kbapi.SecurityDetectionsAPIMachineLearningJobId
			err := mlJobID.FromSecurityDetectionsAPIMachineLearningJobId1(jobIDs)
			if err != nil {
				diags.AddError("Error setting ML job IDs", err.Error())
			} else {
				mlRule.MachineLearningJobId = mlJobID
			}
		}
	}

	// ML rules don't use index patterns or query, and don't support DataViewID or Filters;
	// buildCommonRuleProps leaves those nil automatically since mlRule has no such fields.
	d.setCommonUpdateProps(ctx, buildCommonRuleProps(&mlRule), &diags)

	// Convert to union type
	err := updateProps.FromSecurityDetectionsAPIMachineLearningRuleUpdateProps(mlRule)
	if err != nil {
		diags.AddError(
			"Error building update properties",
			"Could not convert ML rule properties: "+err.Error(),
		)
	}

	return updateProps, diags
}

func (d *Data) updateFromMachineLearningRule(ctx context.Context, rule *kbapi.SecurityDetectionsAPIMachineLearningRule) diag.Diagnostics {
	var diags diag.Diagnostics

	// ML rules don't support DataViewId or Index; buildCommonAPIRuleFields leaves those
	// zero-valued automatically since rule has no such fields.
	diags.Append(d.updateCommonRuleFieldsFromAPI(ctx, buildCommonAPIRuleFields(rule.Id.String(), rule))...)

	// ML rules don't have query or language
	d.Query = types.StringNull()
	d.Language = types.StringNull()

	// ML-specific fields
	d.AnomalyThreshold = types.Int64Value(int64(rule.AnomalyThreshold))

	if singleJobID, err := rule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId0(); err == nil {
		d.MachineLearningJobID = typeutils.ListValueFrom(ctx, []string{singleJobID}, types.StringType, path.Root("machine_learning_job_id"), &diags)
	} else if multipleJobIDs, err := rule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId1(); err == nil {
		jobIDStrings := make([]string, len(multipleJobIDs))
		copy(jobIDStrings, multipleJobIDs)
		d.MachineLearningJobID = typeutils.ListValueFrom(ctx, jobIDStrings, types.StringType, path.Root("machine_learning_job_id"), &diags)
	} else {
		d.MachineLearningJobID = typeutils.StringsToListMust(nil)
	}

	return diags
}
