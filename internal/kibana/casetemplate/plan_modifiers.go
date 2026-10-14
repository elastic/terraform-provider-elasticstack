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

package casetemplate

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
)

// suppressSemanticYAMLDiff keeps the prior state value when the planned YAML
// parses to the same structure. Every update creates a new template version
// server-side, so a reformat-only edit must not plan an update at all.
func suppressSemanticYAMLDiff() planmodifier.String {
	return semanticYAMLDiffSuppressor{}
}

type semanticYAMLDiffSuppressor struct{}

func (semanticYAMLDiffSuppressor) Description(context.Context) string {
	return "Keeps the prior state value when the planned YAML is semantically equal to it."
}

func (s semanticYAMLDiffSuppressor) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (semanticYAMLDiffSuppressor) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	equal, diags := customtypes.NewNormalizedYamlValue(req.StateValue.ValueString()).
		StringSemanticEquals(ctx, customtypes.NewNormalizedYamlValue(req.PlanValue.ValueString()))
	resp.Diagnostics.Append(diags...)
	if equal {
		resp.PlanValue = types.StringValue(req.StateValue.ValueString())
	}
}
