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

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// AtLeastOneOfNestedAttrsOpts configures AtLeastOneOfNestedAttrsValidator.
type AtLeastOneOfNestedAttrsOpts struct {
	// AttrNames lists the nested attribute names, at least one of which must be set.
	AttrNames []string
	// Summary is the error summary used for the "none set" diagnostic.
	Summary string
	// Detail is the diagnostic detail when none of AttrNames are set.
	Detail string
	// Description is an optional MarkdownDescription returned by the validator (defaults to a generic phrase).
	Description string
}

// AtLeastOneOfNestedAttrsValidator returns an object validator that enforces at least one of the
// nested attribute names being set on the validated object. Unknown values (the whole object or
// any named child) defer the check so plan-time references resolve cleanly.
func AtLeastOneOfNestedAttrsValidator(opts AtLeastOneOfNestedAttrsOpts) validator.Object {
	return atLeastOneOfNestedAttrsValidator{opts: opts}
}

type atLeastOneOfNestedAttrsValidator struct {
	opts AtLeastOneOfNestedAttrsOpts
}

func (v atLeastOneOfNestedAttrsValidator) Description(_ context.Context) string {
	if v.opts.Description != "" {
		return v.opts.Description
	}
	return "Ensures at least one of the listed nested attributes is set."
}

func (v atLeastOneOfNestedAttrsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v atLeastOneOfNestedAttrsValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	setCount, unknownCount, _ := CountNestedAttrs(attrs, v.opts.AttrNames, func(attr.Value) bool { return true })
	if setCount > 0 || unknownCount > 0 {
		return
	}

	resp.Diagnostics.AddAttributeError(req.Path, v.opts.Summary, v.opts.Detail)
}
