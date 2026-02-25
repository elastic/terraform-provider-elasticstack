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
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func StringMatchesISO8601Regex(s string) (matched bool, err error) {
	pattern := `(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d\.\d+([+-][0-2]\d:[0-5]\d|Z))` +
		`|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))` +
		`|(\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d([+-][0-2]\d:[0-5]\d|Z))`
	return regexp.MatchString(pattern, s)
}

type StringIsISO8601 struct{}

func (s StringIsISO8601) Description(_ context.Context) string {
	return "a valid ISO8601 date and time formatted string"
}

func (s StringIsISO8601) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s StringIsISO8601) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if matched, err := StringMatchesISO8601Regex(req.ConfigValue.ValueString()); err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid ISO8601 string",
			"This value must be a valid ISO8601 date and time formatted string.",
		)
		return
	}
}
