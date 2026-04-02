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

package acctestconfigdirlint

const (
	// msgInlineConfigWithoutExternalProviders is reported when a step uses Config without ExternalProviders.
	// Ordinary steps must use directory-backed fixtures via ConfigDirectory: acctest.NamedTestCaseDirectory("case-name").
	msgInlineConfigWithoutExternalProviders = "resource.TestStep sets Config without ExternalProviders; " +
		"ordinary steps must use ConfigDirectory: acctest.NamedTestCaseDirectory(\"case-name\"), " +
		"or pair Config with ExternalProviders for compatibility steps"

	// msgConfigDirectoryNotNamedHelper is reported when ConfigDirectory is set to something other than
	// a direct call to acctest.NamedTestCaseDirectory(...).
	msgConfigDirectoryNotNamedHelper = "resource.TestStep sets ConfigDirectory to a value other than acctest.NamedTestCaseDirectory(...); " +
		"use ConfigDirectory: acctest.NamedTestCaseDirectory(\"case-name\") for ordinary directory-backed steps"

	// msgExternalProvidersWithoutConfig is reported when ExternalProviders is set but Config is not.
	// Compatibility steps that declare ExternalProviders must also set inline Config (no ConfigDirectory).
	msgExternalProvidersWithoutConfig = "resource.TestStep sets ExternalProviders without inline Config; " +
		"compatibility steps must pair ExternalProviders with Config: \"...\", not with ConfigDirectory"

	// msgExternalProvidersWithConfigDirectory is reported when both ExternalProviders and ConfigDirectory are set.
	// This is an invalid mix: compatibility steps use ExternalProviders + Config, not ConfigDirectory.
	msgExternalProvidersWithConfigDirectory = "resource.TestStep sets both ExternalProviders and ConfigDirectory; " +
		"compatibility steps must pair ExternalProviders with Config: \"...\", not with ConfigDirectory"

	// msgTestCaseProtoV6ProviderFactories is reported when resource.TestCase sets ProtoV6ProviderFactories.
	msgTestCaseProtoV6ProviderFactories = "resource.TestCase sets ProtoV6ProviderFactories; " +
		"declare ProtoV6ProviderFactories on each ordinary resource.TestStep instead, " +
		"or use ExternalProviders on compatibility steps"

	// msgMissingStepProviderWiring is reported when a step sets neither ProtoV6ProviderFactories nor ExternalProviders.
	msgMissingStepProviderWiring = "resource.TestStep sets neither ProtoV6ProviderFactories nor ExternalProviders; " +
		"ordinary steps must set ProtoV6ProviderFactories on the step, " +
		"and backwards-compatibility steps must set ExternalProviders (with inline Config)"

	// msgMixedStepProviderWiring is reported when a step sets both wiring modes.
	msgMixedStepProviderWiring = "resource.TestStep sets both ProtoV6ProviderFactories and ExternalProviders; " +
		"choose exactly one: ProtoV6ProviderFactories for ordinary coverage, or ExternalProviders for compatibility steps"
)
