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

package slm

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var validExpandWildcardValues = []string{"all", "open", "closed", "hidden", "none"}

// expandWildcardsValidator validates a comma-separated list of wildcard expansion values.
type expandWildcardsValidator struct{}

func (v expandWildcardsValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Each comma-separated value must be one of: %s", strings.Join(validExpandWildcardValues, ", "))
}

func (v expandWildcardsValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Each comma-separated value must be one of: `%s`", strings.Join(validExpandWildcardValues, "`, `"))
}

func (v expandWildcardsValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	for part := range strings.SplitSeq(val, ",") {
		trimmed := strings.TrimSpace(part)
		if !slices.Contains(validExpandWildcardValues, trimmed) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid expand_wildcards value",
				fmt.Sprintf("%q is not a valid value for expand_wildcards. Valid values are: %s", trimmed, strings.Join(validExpandWildcardValues, ", ")),
			)
		}
	}
}
