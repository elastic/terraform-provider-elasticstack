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

package ml

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// mlDurationMessage is the canonical error message for ML short-duration fields.
const mlDurationMessage = "must be a valid duration (e.g., 15m, 1h)"

// mlDurationRegexp matches the ML short-duration format: one or more digits
// followed by a single unit character: n (nanos), s (seconds), u (micros),
// m (minutes), d (days), h (hours).
var mlDurationRegexp = regexp.MustCompile(`^\d+[nsumdh]$`)

// Duration returns a validator that accepts ML short-duration strings such as
// "15m", "1h", "150s". This covers the same format accepted by the bucket_span,
// frequency, query_delay, chunking time_span, and check_window attributes.
func Duration() validator.String {
	return stringvalidator.RegexMatches(mlDurationRegexp, mlDurationMessage)
}
