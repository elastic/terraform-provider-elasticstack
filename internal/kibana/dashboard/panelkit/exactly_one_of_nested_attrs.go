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

package panelkit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// ExactlyOneOfNestedAttrsOpts configures ExactlyOneOfNestedAttrsValidator.
type ExactlyOneOfNestedAttrsOpts struct {
	// AttrNames lists the mutually exclusive nested attribute names; must contain at least two.
	AttrNames []string
	// Summary is the error summary used for both "missing" and "too many" diagnostics.
	Summary string
	// MissingDetail is the diagnostic detail when zero of AttrNames are set.
	MissingDetail string
	// TooManyDetail is the diagnostic detail when more than one of AttrNames is set.
	TooManyDetail string
	// Description is an optional MarkdownDescription returned by the validator (defaults to a generic phrase).
	Description string
}

// ExactlyOneOfNestedAttrsValidator returns an object validator that enforces exactly one of the
// nested attribute names being set on the validated object. Unknown values defer the check so
// plan-time references resolve cleanly.
func ExactlyOneOfNestedAttrsValidator(opts ExactlyOneOfNestedAttrsOpts) validator.Object {
	return exactlyOneOfNestedAttrsValidator{opts: opts}
}

type exactlyOneOfNestedAttrsValidator struct {
	opts ExactlyOneOfNestedAttrsOpts
}

func (v exactlyOneOfNestedAttrsValidator) Description(_ context.Context) string {
	if v.opts.Description != "" {
		return v.opts.Description
	}
	return "Ensures exactly one of the listed nested attributes is set."
}

func (v exactlyOneOfNestedAttrsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v exactlyOneOfNestedAttrsValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	setCount := 0
	hasUnknown := false
	for _, name := range v.opts.AttrNames {
		av, ok := attrs[name]
		if !ok || av == nil {
			continue
		}
		switch {
		case av.IsUnknown():
			hasUnknown = true
		case av.IsNull():
		default:
			setCount++
		}
	}
	if setCount > 1 {
		resp.Diagnostics.AddAttributeError(req.Path, v.opts.Summary, v.opts.TooManyDetail)
		return
	}
	if hasUnknown {
		return
	}
	if setCount == 0 {
		resp.Diagnostics.AddAttributeError(req.Path, v.opts.Summary, v.opts.MissingDetail)
	}
}
