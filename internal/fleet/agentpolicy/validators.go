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

package agentpolicy

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const policyIDValidatorDescription = `Must be 1-255 characters and must not contain path separators ("/"), traversal sequences (".."), or reserved keys ("__proto__", "constructor", "prototype").`

var (
	_ validator.String = policyIDValidator{}
)

type policyIDValidator struct{}

func (policyIDValidator) Description(_ context.Context) string {
	return policyIDValidatorDescription
}

func (v policyIDValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (policyIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if len(value) < 1 || len(value) > 255 {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid policy_id length",
			"policy_id must be between 1 and 255 characters (inclusive).",
		))
		return
	}

	if strings.Contains(value, "/") {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid policy_id",
			`policy_id must not contain path separators ("/").`,
		))
		return
	}

	if strings.Contains(value, "..") {
		resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			req.Path,
			"Invalid policy_id",
			`policy_id must not contain traversal sequences ("..").`,
		))
		return
	}

	for _, reserved := range []string{"__proto__", "constructor", "prototype"} {
		if strings.Contains(value, reserved) {
			resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				req.Path,
				"Invalid policy_id",
				fmt.Sprintf(`policy_id must not contain reserved keys (%q).`, reserved),
			))
			return
		}
	}
}
