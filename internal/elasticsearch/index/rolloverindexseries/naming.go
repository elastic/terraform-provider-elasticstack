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

package rolloverindexseries

import (
	"fmt"
	"strings"
	"unicode/utf8"

	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/indexname"
)

const (
	maxElasticsearchNameBytes = 255
	bootstrapGenerationSuffix = "-000001"
)

func bootstrapIndexName(prefix string) string {
	return prefix + bootstrapGenerationSuffix
}

func validatePrefix(prefix string) error {
	if err := validateLiteralName(prefix, "prefix"); err != nil {
		return err
	}
	bootstrap := bootstrapIndexName(prefix)
	if len(bootstrap) > maxElasticsearchNameBytes {
		return fmt.Errorf(
			"generated bootstrap index name %q is %d bytes, which exceeds the Elasticsearch index name limit of %d",
			bootstrap,
			len(bootstrap),
			maxElasticsearchNameBytes,
		)
	}
	return nil
}

func validateRolloverAlias(alias string) error {
	if err := validateLiteralName(alias, "rollover_alias"); err != nil {
		return err
	}
	if strings.HasPrefix(alias, ".") {
		return fmt.Errorf("rollover_alias %q must not start with '.'; Elasticsearch alias names cannot start with a dot", alias)
	}
	if len(alias) > maxElasticsearchNameBytes {
		return fmt.Errorf(
			"rollover_alias %q is %d bytes, which exceeds the Elasticsearch alias name limit of %d",
			alias,
			len(alias),
			maxElasticsearchNameBytes,
		)
	}
	return nil
}

// ValidateNamingFields checks a bootstrap prefix and rollover alias using the
// same literal-name rules as resource configuration. Import calls this before
// hydration so imported names are not treated as discovery expressions.
func ValidateNamingFields(prefix, rolloverAlias string) error {
	if err := validatePrefix(prefix); err != nil {
		return err
	}
	return validateRolloverAlias(rolloverAlias)
}

func validateLiteralName(name, label string) error {
	if name == "" {
		return fmt.Errorf("%s must not be empty", label)
	}
	if name == "." || name == ".." {
		return fmt.Errorf("%s %q is not a valid Elasticsearch name", label, name)
	}
	if strings.ContainsAny(name, "*?") {
		return fmt.Errorf("%s %q contains wildcard syntax; only a literal name is accepted", label, name)
	}
	if strings.Contains(name, ",") {
		return fmt.Errorf("%s %q contains a comma; only a single literal name is accepted", label, name)
	}
	if strings.Contains(name, ":") {
		return fmt.Errorf("%s %q is remote-cluster qualified; only a local literal name is accepted", label, name)
	}
	if strings.ContainsAny(name, "<>") || esclient.DateMathIndexNameRe.MatchString(name) {
		return fmt.Errorf("%s %q uses date math; only a literal name is accepted", label, name)
	}
	r, _ := utf8.DecodeRuneInString(name)
	if r == '-' || r == '_' || r == '+' {
		return fmt.Errorf("%s %q must not start with '-', '_', or '+'", label, name)
	}
	if !indexname.AllowedCharsRegexp.MatchString(name) {
		return fmt.Errorf("%s %q contains characters that are not allowed in an Elasticsearch name", label, name)
	}
	return nil
}
