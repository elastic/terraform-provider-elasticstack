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
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// elasticDurationRegexp matches the Elastic time-unit duration format
// documented at
// https://www.elastic.co/guide/en/elasticsearch/reference/current/api-conventions.html#time-units
// (d, h, m, s, ms, micros, nanos).
var elasticDurationRegexp = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?(?:d|h|m|s|ms|micros|nanos)$`)

// ElasticDuration returns a Plugin Framework string validator that ensures the
// value is a valid Elastic time-unit duration (d, h, m, s, ms, micros, nanos).
// Null and unknown values are skipped.
func ElasticDuration() validator.String {
	return elasticDurationValidator{}
}

type elasticDurationValidator struct{}

var _ validator.String = elasticDurationValidator{}

func (v elasticDurationValidator) Description(_ context.Context) string {
	return "must be a valid Elastic duration (e.g. 1d, 2h, 30m, 60s, 500ms, 1micros, 1nanos)"
}

func (v elasticDurationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v elasticDurationValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	val := req.ConfigValue
	if val.IsNull() || val.IsUnknown() {
		return
	}
	s := val.ValueString()
	if s == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Elastic duration", "duration must not be empty")
		return
	}
	if !elasticDurationRegexp.MatchString(s) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Elastic duration",
			fmt.Sprintf("%q is not a valid Elastic time-unit duration", s),
		)
	}
}
