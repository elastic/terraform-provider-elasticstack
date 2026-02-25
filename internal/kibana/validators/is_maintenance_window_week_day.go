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

func StringMatchesOnWeekDayRegex(s string) (matched bool, err error) {
	pattern := `^(((\+|-)[1-5])?(MO|TU|WE|TH|FR|SA|SU))$`
	return regexp.MatchString(pattern, s)
}

type StringIsMaintenanceWindowOnWeekDay struct{}

func (s StringIsMaintenanceWindowOnWeekDay) Description(_ context.Context) string {
	return "a valid OnWeekDay. Accepted values are specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`)."
}

func (s StringIsMaintenanceWindowOnWeekDay) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s StringIsMaintenanceWindowOnWeekDay) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if matched, err := StringMatchesOnWeekDayRegex(req.ConfigValue.ValueString()); err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid OnWeekDay",
			"This value must be a valid OnWeekDay. Accepted values are specific days of the week (`[MO,TU,WE,TH,FR,SA,SU]`) or nth day of month (`[+1MO, -3FR, +2WE, -4SA, -5SU]`).",
		)
		return
	}
}
