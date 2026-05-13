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

var _ validator.Object = drilldownItemModeValidator{}

type drilldownItemModeValidator struct{}

func (drilldownItemModeValidator) Description(_ context.Context) string {
	return "Ensures exactly one of `dashboard`, `discover`, or `url` is set inside each drilldown list item."
}

func (v drilldownItemModeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (drilldownItemModeValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	setCount := func(name string) bool {
		av, ok := attrs[name]
		if !ok || av == nil {
			return false
		}
		return !av.IsNull() && !av.IsUnknown()
	}
	dashboard := attrs["dashboard"]
	discover := attrs["discover"]
	url := attrs["url"]
	hasUnknown :=
		dashboard != nil && dashboard.IsUnknown() ||
			discover != nil && discover.IsUnknown() ||
			url != nil && url.IsUnknown()
	if hasUnknown {
		return
	}
	count := 0
	if setCount("dashboard") {
		count++
	}
	if setCount("discover") {
		count++
	}
	if setCount("url") {
		count++
	}
	if count == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid drilldown entry",
			"Set exactly one of `dashboard`, `discover`, or `url` on each drilldown list item.",
		)
		return
	}
	if count > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid drilldown entry",
			"`dashboard`, `discover`, and `url` are mutually exclusive; set exactly one per drilldown list item.",
		)
	}
}
