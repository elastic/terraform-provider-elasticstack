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

// StringMatchesHoursRegex checks if the string matches HH:mm format.
func StringMatchesHoursRegex(s string) (matched bool, err error) {
	pattern := `^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$`
	return regexp.MatchString(pattern, s)
}

// StringIsHours validates that a string is in HH:mm format.
var StringIsHours = regexStringValidator{
	description: "a valid time in 24-hour notation (HH:mm)",
	errSummary:  "expected value to be a valid time in 24-hour notation (HH:mm)",
	errDetail:   "This value must be a valid time in 24-hour notation (HH:mm). For example: 09:00, 14:30, 23:59.",
	matchFn:     StringMatchesHoursRegex,
}
