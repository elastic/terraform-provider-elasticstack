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
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var intervalFrequencyRegex = regexp.MustCompile(`^[1-9][0-9]*(?:d|w|M|y)$`)

func StringMatchesIntervalFrequencyRegex(s string) (matched bool, err error) {
	return intervalFrequencyRegex.MatchString(s), nil
}

const maintenanceWindowIntervalFrequencyDescription = "a valid interval/frequency. Allowed values are in the `<integer><unit>` format. " +
	"`<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. " +
	"For example: `15d`, `2w`, `3m`, `1y`."

var StringIsMaintenanceWindowIntervalFrequency = stringvalidator.RegexMatches(
	intervalFrequencyRegex,
	"This value must be "+maintenanceWindowIntervalFrequencyDescription,
)
