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

package security

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// UsernameAllowedCharsError is the validation error message for usernames
// that contain disallowed characters.
const UsernameAllowedCharsError = "must contain alphanumeric characters (a-z, A-Z, 0-9), spaces, punctuation, and printable symbols " +
	"in the Basic Latin (ASCII) block. Leading or trailing whitespace is not allowed"

var usernameRegexp = regexp.MustCompile(`^[[:graph:]]+$`)

// UsernameValidator returns a string validator that enforces the username character-class rule.
func UsernameValidator() validator.String {
	return stringvalidator.RegexMatches(usernameRegexp, UsernameAllowedCharsError)
}
