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

// Package indexname provides shared validators for Elasticsearch index names.
package indexname

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	// AllowedCharsMessage is the validation error message for index names that contain disallowed characters.
	AllowedCharsMessage = "must contain lower case alphanumeric characters and selected punctuation, see: " +
		"https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html#indices-create-api-path-params"
)

// LeadingCharRegexp matches index names that do not start with -, _, or +.
var LeadingCharRegexp = regexp.MustCompile(`^[^-_+]`)

// AllowedCharsRegexp matches index names composed only of allowed characters.
var AllowedCharsRegexp = regexp.MustCompile(`^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$`)

// NameValidators returns the two string validators that enforce Elasticsearch
// index naming rules: no leading -, _, or +; and only allowed characters.
func NameValidators() []validator.String {
	return []validator.String{
		stringvalidator.RegexMatches(LeadingCharRegexp, "cannot start with -, _, +"),
		stringvalidator.RegexMatches(AllowedCharsRegexp, AllowedCharsMessage),
	}
}
