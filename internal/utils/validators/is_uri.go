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
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type uriValidator struct {
	allowedSchemes []string
	requireHost   bool
}

// IsURI returns a string validator that ensures the value is a parseable
// absolute URI with a non-empty scheme. If allowedSchemes is non-empty, the
// URI's scheme must match (case-insensitively) one of the supplied schemes.
func IsURI(allowedSchemes ...string) validator.String {
	normalized := make([]string, len(allowedSchemes))
	for i, s := range allowedSchemes {
		normalized[i] = strings.ToLower(s)
	}
	return uriValidator{allowedSchemes: normalized}
}

// IsURL is like IsURI but additionally requires a non-empty host, ensuring
// values like "https://" or "http:example" are rejected.
func IsURL(allowedSchemes ...string) validator.String {
	normalized := make([]string, len(allowedSchemes))
	for i, s := range allowedSchemes {
		normalized[i] = strings.ToLower(s)
	}
	return uriValidator{allowedSchemes: normalized, requireHost: true}
}

func (v uriValidator) Description(_ context.Context) string {
	if len(v.allowedSchemes) == 0 {
		return "value must be a valid URI with a scheme"
	}
	return fmt.Sprintf("value must be a valid URI using one of the following schemes: %s", strings.Join(v.allowedSchemes, ", "))
}

func (v uriValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v uriValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if !typeutils.IsKnown(req.ConfigValue) {
		return
	}

	val := req.ConfigValue.ValueString()
	if val == "" {
		return
	}

	parsed, err := url.Parse(val)
	if err != nil || parsed.Scheme == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid URI",
			fmt.Sprintf("%q must be a valid URI with a scheme.", val),
		)
		return
	}

	if len(v.allowedSchemes) > 0 && !slices.Contains(v.allowedSchemes, strings.ToLower(parsed.Scheme)) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid URI scheme",
			fmt.Sprintf("%q uses scheme %q; allowed schemes are: %s.", val, parsed.Scheme, strings.Join(v.allowedSchemes, ", ")),
		)
		return
	}

	if v.requireHost && parsed.Host == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid URL",
			fmt.Sprintf("%q must be a valid URL with a non-empty host.", val),
		)
	}
}
