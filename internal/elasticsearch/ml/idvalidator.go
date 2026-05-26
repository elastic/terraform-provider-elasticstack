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

// idAllowedCharsMessage describes the allowed characters for ML calendar,
// job, job-group, datafeed, and filter identifiers as enforced by Elasticsearch.
const idAllowedCharsMessage = "must contain lowercase alphanumeric characters (a-z and 0-9), dots, hyphens, " +
	"and underscores; it must start and end with an alphanumeric character"

// pathIDRegexp matches the Elasticsearch ML path identifier rules shared by
// calendars, anomaly detection jobs, job groups, datafeeds, and filters:
// lowercase letters, digits, underscore, hyphen, dot; must start and end with
// alphanumeric. Equivalent to: ^[a-z0-9](?:[a-z0-9_\-\.]*[a-z0-9])?$
var pathIDRegexp = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9_.-]*[a-z0-9])?$`)

// IDValidator returns a single compound validator covering the standard
// Elasticsearch ML identifier rules: length between 1 and 64 characters and
// the [pathIDRegexp] character/anchor pattern. Use for calendar_id, job_id,
// job-group, datafeed, and filter id schema attributes so all ML resources
// share one source of truth instead of duplicating LengthBetween + RegexMatches
// pairs.
func IDValidator() validator.String {
	return stringvalidator.All(
		stringvalidator.LengthBetween(1, 64),
		stringvalidator.RegexMatches(pathIDRegexp, idAllowedCharsMessage),
	)
}
