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

package transform

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// stringAttributeType is the element type for string list attributes.
var stringAttributeType attr.Type = types.StringType

// elasticDurationValidator validates Elastic time unit duration strings (d, h, m, s, ms, micros, nanos).
type elasticDurationValidator struct{}

var _ validator.String = elasticDurationValidator{}

func (v elasticDurationValidator) Description(_ context.Context) string {
	return "must be a valid Elastic duration (e.g. 1d, 2h, 30m, 60s, 500ms, 1micros, 1nanos)"
}

func (v elasticDurationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

var elasticDurationRegexp = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?(?:d|h|m|s|ms|micros|nanos)$`)

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

// goDurationValidator validates Go time.Duration strings (e.g. 30s, 1m).
type goDurationValidator struct{}

var _ validator.String = goDurationValidator{}

func (v goDurationValidator) Description(_ context.Context) string {
	return "must be a valid Go duration string (e.g. 30s, 1m, 1h)"
}

func (v goDurationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v goDurationValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	val := req.ConfigValue
	if val.IsNull() || val.IsUnknown() {
		return
	}
	s := val.ValueString()
	if _, err := time.ParseDuration(s); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid duration",
			fmt.Sprintf("%q is not a valid duration: %s", s, err.Error()),
		)
	}
}
