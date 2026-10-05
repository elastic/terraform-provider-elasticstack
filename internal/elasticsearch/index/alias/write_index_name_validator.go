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

package alias

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type writeIndexNameValidator struct{}

func (writeIndexNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}

	name := req.ConfigValue.ValueString()
	if strings.ContainsAny(name, "*?,") || strings.HasPrefix(name, "-") || strings.HasPrefix(name, "<") || name == "_all" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid write index name",
			fmt.Sprintf("%q must name a single index", name),
		)
	}
}

func (writeIndexNameValidator) Description(context.Context) string {
	return "value must name a single index"
}

func (v writeIndexNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
