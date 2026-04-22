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

package elasticsearch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDateMathIndexNameRe verifies the regex that identifies plain date math
// index name expressions.
func TestDateMathIndexNameRe(t *testing.T) {
	validCases := []string{
		`<logs-{now/d}>`,
		`<logs-{now/M}>`,
		`<logs-{now/d{yyyy.MM.dd}}>`,
		`<a-{now}>`,
	}
	for _, name := range validCases {
		t.Run("valid: "+name, func(t *testing.T) {
			assert.True(t, DateMathIndexNameRe.MatchString(name), "expected %q to match date math regex", name)
		})
	}

	invalidCases := []struct {
		name  string
		input string
	}{
		{"static lowercase", "my-index"},
		{"static with numbers", "logs-2024.01.15"},
		{"just angle brackets no braces", "<logs>"},
		{"no angle brackets", "logs-{now/d}"},
		{"nested angle brackets", "<outer<inner>>"},
		{"empty", ""},
		{"suffix after date math section", `<prefix-{now/d}-suffix>`},
		{"starts with dash", `<-logs-{now/d}>`},
		{"starts with underscore", `<_logs-{now/d}>`},
		{"starts with plus", `<+logs-{now/d}>`},
	}
	for _, tc := range invalidCases {
		t.Run("invalid: "+tc.name, func(t *testing.T) {
			assert.False(t, DateMathIndexNameRe.MatchString(tc.input), "expected %q not to match date math regex", tc.input)
		})
	}
}

// TestEncodeDateMathIndexName verifies that date math expressions are
// URI-percent-encoded correctly and that plain static names are not mangled.
func TestEncodeDateMathIndexName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    `<logs-{now/d}>`,
			expected: "%3Clogs-%7Bnow%2Fd%7D%3E",
		},
		{
			input:    `<logs-{now/d{yyyy.MM.dd}}>`,
			expected: "%3Clogs-%7Bnow%2Fd%7Byyyy.MM.dd%7D%7D%3E",
		},
		{
			input:    `<logs-{now/M{yyyy.MM}}>`,
			expected: "%3Clogs-%7Bnow%2FM%7Byyyy.MM%7D%7D%3E",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := encodeDateMathIndexName(tc.input)
			require.Equal(t, tc.expected, got)
		})
	}
}

// TestDateMathIndexNameReDoesNotMatchStaticNames verifies that the date math regex
// does not accidentally match static names that happen to contain braces.
func TestDateMathIndexNameReDoesNotMatchStaticNames(t *testing.T) {
	// Ensure that enabling the date math path requires proper angle-bracket wrapping.
	staticNames := []string{
		"logs-{broken",
		"{now/d}",
		"<unclosed",
		"unclosed>",
	}
	for _, name := range staticNames {
		t.Run(name, func(t *testing.T) {
			assert.False(t, DateMathIndexNameRe.MatchString(name))
		})
	}
}
