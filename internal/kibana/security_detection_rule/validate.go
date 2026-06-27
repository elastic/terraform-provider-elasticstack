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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ValidateConfig validates the configuration for a security detection rule resource.
// It ensures that the configuration meets the following requirements:
//
// - For rule types "esql" and "machine_learning", no additional validation is performed
// - For other rule types, exactly one of 'index' or 'data_view_id' must be specified
// - Both 'index' and 'data_view_id' cannot be set simultaneously
//
// The function adds appropriate error diagnostics if validation fails.
func (r securityDetectionRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data Data

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Type.ValueString() == ruleTypeESQL || data.Type.ValueString() == ruleTypeMachineLearning {
		return
	}

	if typeutils.IsKnown(data.Index) && typeutils.IsKnown(data.DataViewID) {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Both 'index' and 'data_view_id' cannot be set at the same time.",
		)

	}

	if data.Index.IsNull() && data.DataViewID.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"One of 'index' or 'data_view_id' must be set.",
		)
	}
}
