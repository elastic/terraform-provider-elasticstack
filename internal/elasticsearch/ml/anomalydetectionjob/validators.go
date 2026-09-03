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

package anomalydetectionjob

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = resultsIndexNameWithoutCustomPrefixValidator{}

type resultsIndexNameWithoutCustomPrefixValidator struct{}

func resultsIndexNameWithoutCustomPrefix() resultsIndexNameWithoutCustomPrefixValidator {
	return resultsIndexNameWithoutCustomPrefixValidator{}
}

func (v resultsIndexNameWithoutCustomPrefixValidator) Description(_ context.Context) string {
	return "Must not start with \"custom-\"; Elasticsearch automatically adds this prefix to user-defined results index names."
}

func (v resultsIndexNameWithoutCustomPrefixValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v resultsIndexNameWithoutCustomPrefixValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}

	value := req.ConfigValue.ValueString()
	if !strings.HasPrefix(value, "custom-") {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid results_index_name",
		"Do not start the value with \"custom-\"; Elasticsearch automatically adds this prefix. "+
			"The provider strips the prefix when reading job configuration, so including it causes plan/apply drift.",
	)
}
