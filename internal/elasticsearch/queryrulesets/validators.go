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

package queryrulesets

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// queryRuleActionsValidator returns an object validator that enforces exactly
// one of `ids` or `docs` being set on a query rule's `actions` block.
func queryRuleActionsValidator() validator.Object {
	return validators.ExactlyOneOfNestedAttrsValidator(validators.ExactlyOneOfNestedAttrsOpts{
		AttrNames: []string{queryRuleActionsIDsAttrName, queryRuleActionsDocsAttrName},
		// ids and docs are both list-typed; an empty known list (e.g. `ids = []`)
		// must count as unset to match the prior hand-rolled validator's
		// `len(list.Elements()) > 0` rule. Non-list attribute types are treated as set.
		IsSet: func(v attr.Value) bool {
			list, ok := v.(types.List)
			if !ok {
				return true
			}
			return len(list.Elements()) > 0
		},
		Summary:       "Invalid actions configuration",
		MissingDetail: "Exactly one of `ids` or `docs` must be set in `actions`.",
		TooManyDetail: "Exactly one of `ids` or `docs` must be set in `actions`; both cannot be set.",
		Description:   "Exactly one of ids or docs must be set in actions",
	})
}

type queryRuleCriteriaValidator struct{}

func (queryRuleCriteriaValidator) Description(_ context.Context) string {
	return "Validates criteria values against the criteria type"
}

func (v queryRuleCriteriaValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v queryRuleCriteriaValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	if attrs["type"].IsUnknown() {
		return
	}
	if attrs["values"].IsUnknown() {
		return
	}

	criteriaType := stringAttributeValue(attrs["type"])
	valuesAttr := attrs["values"]
	if normalized, ok := valuesAttr.(jsontypes.Normalized); ok && normalized.IsUnknown() {
		return
	}
	valuesSet := valuesAttributeIsSet(valuesAttr)

	switch {
	case criteriaType == "always" && valuesSet:
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("values"),
			"Invalid criteria values",
			"`values` must be omitted or null when `type` is `always`.",
		)
	case criteriaType != "always" && criteriaType != "" && !valuesSet:
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("values"),
			"Invalid criteria values",
			"`values` is required when `type` is not `always`.",
		)
	}
}

type criteriaValuesJSONValidator struct{}

func (criteriaValuesJSONValidator) Description(_ context.Context) string {
	return "Must be a valid non-empty JSON array string"
}

func (v criteriaValuesJSONValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v criteriaValuesJSONValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	rawInput := strings.TrimSpace(req.ConfigValue.ValueString())
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(rawInput), &raw); err != nil {
		resp.Diagnostics.Append(criteriaValuesJSONDiagnostic(req.Path))
		return
	}

	if len(raw) == 0 || raw[0] != '[' {
		resp.Diagnostics.Append(criteriaValuesJSONDiagnostic(req.Path))
		return
	}

	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		resp.Diagnostics.Append(criteriaValuesJSONDiagnostic(req.Path))
		return
	}

	if len(values) == 0 {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid criteria values",
			"`values` must be a non-empty JSON array string; empty arrays are not allowed.",
		))
	}
}

func criteriaValuesJSONDiagnostic(p path.Path) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		p,
		"Invalid criteria values",
		"`values` must be a valid JSON array string.",
	)
}

func stringAttributeValue(val attr.Value) string {
	if val == nil || val.IsNull() || val.IsUnknown() {
		return ""
	}

	str, ok := val.(types.String)
	if !ok {
		return ""
	}

	return str.ValueString()
}

func valuesAttributeIsSet(val attr.Value) bool {
	if val == nil || val.IsNull() || val.IsUnknown() {
		return false
	}

	normalized, ok := val.(jsontypes.Normalized)
	if !ok {
		return true
	}

	return typeutils.IsKnown(normalized)
}

// Ensure validators satisfy interfaces at compile time.
var (
	_ validator.Object = queryRuleCriteriaValidator{}
	_ validator.String = criteriaValuesJSONValidator{}
)
