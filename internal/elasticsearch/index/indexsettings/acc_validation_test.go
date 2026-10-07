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

package indexsettings_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// REQ-001 plan-time schema-validation scenarios. Each test pins the
// diagnostic text asserted in the requirement's scenario.
func TestAccResourceIndexSettings_validationIndexRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)index`),
			},
		},
	})
}

func TestAccResourceIndexSettings_validationOverlapTypedAttributeAndSettingsJSON(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)number_of_replicas`),
			},
		},
	})
}

func TestAccResourceIndexSettings_validationSettingsJSONNestedObject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)flat dotted setting keys`),
			},
		},
	})
}

func TestAccResourceIndexSettings_validationSettingsJSONExplicitNull(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)null`),
			},
		},
	})
}

func TestAccResourceIndexSettings_validationSettingsJSONEmptyObject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)at least one setting`),
			},
		},
	})
}

func TestAccResourceIndexSettings_validationSettingsJSONStaticKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)can only be set at index creation time`),
			},
		},
	})
}

// Unmodeled dynamic keys are permitted: `terraform validate`/`plan` SHALL NOT
// emit a validation error. A plan-only step exercises plan-time validation
// without applying, so the not-yet-implemented Create callback is never
// invoked and the config must validate cleanly.
func TestAccResourceIndexSettings_validationSettingsJSONUnmodeledKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				PlanOnly:                 true,
				ExpectNonEmptyPlan:       true,
			},
		},
	})
}

func TestAccResourceIndexSettings_validationAtLeastOneSettingRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ExpectError:              regexp.MustCompile(`(?i)at least one`),
			},
		},
	})
}
