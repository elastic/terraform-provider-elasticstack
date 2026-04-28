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

package slo

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// kqlObjectFormMeaningful requires that a configured (known) KQL object form includes a non-blank
// kql_query and/or a non-empty filters list. Unknown attribute values are accepted so read/computed
// refresh does not spuriously fail.
type kqlObjectFormMeaningful struct{}

const kqlObjectFormMeaningfulDescription = "when set, the object form must include kql_query and/or a non-empty filters list"

func (kqlObjectFormMeaningful) Description(_ context.Context) string {
	return kqlObjectFormMeaningfulDescription
}

func (kqlObjectFormMeaningful) MarkdownDescription(_ context.Context) string {
	return kqlObjectFormMeaningfulDescription
}

func (kqlObjectFormMeaningful) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	o := req.ConfigValue
	if o.IsNull() || o.IsUnknown() {
		return
	}
	attrs := o.Attributes()
	kq, hasKq := attrs["kql_query"].(types.String)
	filters, hasFilters := attrs["filters"].(types.List)
	if hasKq && kq.IsUnknown() {
		return
	}
	if hasFilters && filters.IsUnknown() {
		return
	}
	kqlNonBlank := hasKq && !kq.IsNull() && !kq.IsUnknown() && strings.TrimSpace(kq.ValueString()) != ""
	filtersNonEmpty := hasFilters && !filters.IsNull() && !filters.IsUnknown() && len(filters.Elements()) > 0
	if kqlNonBlank || filtersNonEmpty {
		return
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Configuration",
		"When using the KQL object form, set a non-blank kql_query and/or a non-empty filters list.",
	)
}
