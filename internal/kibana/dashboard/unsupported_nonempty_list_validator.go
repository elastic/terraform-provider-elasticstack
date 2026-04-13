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

package dashboard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.List = unsupportedNonEmptyListValidator{}

type unsupportedNonEmptyListValidator struct {
	summary string
	detail  string
}

func unsupportedNonEmptyList(summary, detail string) validator.List {
	return unsupportedNonEmptyListValidator{
		summary: summary,
		detail:  detail,
	}
}

func (v unsupportedNonEmptyListValidator) Description(_ context.Context) string {
	return v.detail
}

func (v unsupportedNonEmptyListValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v unsupportedNonEmptyListValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if len(req.ConfigValue.Elements()) == 0 {
		return
	}

	resp.Diagnostics.AddAttributeError(req.Path, v.summary, v.detail)
}
