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

var alertingDurationPattern = "^[1-9][0-9]*(?:d|h|m|s)$"

func StringMatchesAlertingDurationRegex(s string) (matched bool, err error) {
	return regexp.MatchString(alertingDurationPattern, s)
}

type StringIsAlertingDuration struct{}

func (s StringIsAlertingDuration) Description(_ context.Context) string {
	return "a valid alerting duration in seconds (s), minutes (m), hours (h), or days (d)"
}

func (s StringIsAlertingDuration) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}

func (s StringIsAlertingDuration) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if matched, err := StringMatchesAlertingDurationRegex(req.ConfigValue.ValueString()); err != nil || !matched {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"expected value to be a valid alerting duration",
			"This value must be a valid alerting duration in seconds (s), minutes (m), hours (h), or days (d).",
		)
		return
	}
}
