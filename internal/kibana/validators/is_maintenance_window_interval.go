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
)

func StringMatchesIntervalFrequencyRegex(s string) (matched bool, err error) {
	pattern := `^[1-9][0-9]*(?:d|w|M|y)$`
	return regexp.MatchString(pattern, s)
}

const maintenanceWindowIntervalFrequencyDescription = "a valid interval/frequency. Allowed values are in the `<integer><unit>` format. " +
	"`<unit>` is one of `d`, `w`, `M`, or `y` for days, weeks, months, years. " +
	"For example: `15d`, `2w`, `3m`, `1y`."

var StringIsMaintenanceWindowIntervalFrequency = regexStringValidator{
	description: maintenanceWindowIntervalFrequencyDescription,
	errSummary:  "expected value to be a valid interval/frequency",
	errDetail:   "This value must be " + maintenanceWindowIntervalFrequencyDescription,
	matchFn:     StringMatchesIntervalFrequencyRegex,
}
