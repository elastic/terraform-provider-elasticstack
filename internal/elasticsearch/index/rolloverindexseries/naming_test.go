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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNaming_wildcardPrefixRejected(t *testing.T) {
	err := validatePrefix("orders-*")
	require.Error(t, err)
	require.Error(t, ValidateNamingFields("orders-*", "orders-write"))
	require.Contains(t, err.Error(), "orders-*")
	require.Contains(t, strings.ToLower(err.Error()), "wildcard")
}

func TestNaming_bootstrapNameLengthValidatedInFull(t *testing.T) {
	validPrefix := strings.Repeat("a", maxElasticsearchNameBytes-len(bootstrapGenerationSuffix))
	require.NoError(t, validatePrefix(validPrefix))

	tooLongPrefix := validPrefix + "b"
	bootstrap := bootstrapIndexName(tooLongPrefix)
	require.Greater(t, len(bootstrap), maxElasticsearchNameBytes)

	err := validatePrefix(tooLongPrefix)
	require.Error(t, err)
	require.Contains(t, err.Error(), bootstrap)
	require.Error(t, ValidateNamingFields(tooLongPrefix, "orders-write"))
}

func TestNaming_commaSeparatedRolloverAliasRejected(t *testing.T) {
	err := validateRolloverAlias("orders-write,other-alias")
	require.Error(t, err)
	require.Contains(t, err.Error(), "orders-write,other-alias")
	require.Contains(t, strings.ToLower(err.Error()), "comma")
	require.Error(t, ValidateNamingFields("orders", "orders-write,other-alias"))
}

func TestNaming_literalNamesRejectDiscoverySyntax(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
		alias  string
		needle string
	}{
		{name: "wildcard alias", prefix: "orders", alias: "orders-*", needle: "wildcard"},
		{name: "question mark prefix", prefix: "orders-?", alias: "orders-write", needle: "wildcard"},
		{name: "comma prefix", prefix: "orders,other", alias: "orders-write", needle: "comma"},
		{name: "remote prefix", prefix: "remote:orders", alias: "orders-write", needle: "remote-cluster"},
		{name: "remote alias", prefix: "orders", alias: "remote:orders-write", needle: "remote-cluster"},
		{name: "date math prefix", prefix: "<orders-{now/d}>", alias: "orders-write", needle: "date math"},
		{name: "date math alias", prefix: "orders", alias: "<orders-write-{now/d}>", needle: "date math"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateNamingFields(tc.prefix, tc.alias)
			require.Error(t, err)
			require.Contains(t, strings.ToLower(err.Error()), tc.needle)
		})
	}
}

func TestNaming_charsetAndLengthBoundaries(t *testing.T) {
	require.NoError(t, ValidateNamingFields("a", "a"))
	require.NoError(t, ValidateNamingFields("orders.write_1", "orders-write"))
	require.NoError(t, validatePrefix(".orders"))

	require.Error(t, validatePrefix(""))
	require.Error(t, validatePrefix("-orders"))
	require.Error(t, validatePrefix("_orders"))
	require.Error(t, validatePrefix("+orders"))
	require.Error(t, validatePrefix("Orders"))
	require.Error(t, validatePrefix("."))
	require.Error(t, validatePrefix(".."))

	maxAlias := strings.Repeat("b", maxElasticsearchNameBytes)
	require.NoError(t, validateRolloverAlias(maxAlias))
	require.Error(t, validateRolloverAlias(maxAlias+"c"))
}

func TestNaming_aliasRejectsLeadingDotAllowedOnIndexPrefix(t *testing.T) {
	require.NoError(t, validatePrefix(".orders"))
	err := validateRolloverAlias(".orders-write")
	require.Error(t, err)
	require.Contains(t, err.Error(), ".orders-write")
	require.Error(t, ValidateNamingFields("orders", ".orders-write"))
}
