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

package template_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// slowlogIncludeUserMinVersion is the lowest Elasticsearch version that accepts
// the index.search.slowlog.include.user setting (added in 8.14.0). Older
// versions reject the field with a 400, so the regression test is skipped
// against them.
var slowlogIncludeUserMinVersion = version.Must(version.NewVersion("8.14.0"))

// TestAccReproduceIssue3124 is the regression test for
// https://github.com/elastic/terraform-provider-elasticstack/issues/3124.
//
// An index_template whose template.settings JSON contained
// index.search.slowlog.include.user previously failed apply with
// "Provider produced inconsistent result after apply" because the read path
// decoded the response through the typed go-elasticsearch SlowlogSettings
// struct, which has no Include field and silently dropped it during
// deserialization. The same class of bug affected
// index.lifecycle.parse_origination_date, whose typed field is *bool and
// coerced the user-supplied "true" string into a bool literal on read.
//
// The provider now decodes index template responses through
// internal/models.IndexTemplate (settings as map[string]any) so unmodeled
// fields and string-encoded scalars survive the refresh.
func TestAccReproduceIssue3124(t *testing.T) {
	versionutils.SkipIfUnsupported(t, slowlogIncludeUserMinVersion, versionutils.FlavorAny)

	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	addr := "elasticstack_elasticsearch_index_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("repro"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrIndexSettingsSemantic(addr, `{
						"index": {
							"number_of_shards": "1",
							"number_of_replicas": "0",
							"search": {
								"slowlog": {
									"include": {
										"user": "true"
									},
									"threshold": {
										"query": {
											"warn": "10s"
										}
									}
								}
							},
							"lifecycle": {
								"parse_origination_date": "true"
							}
						}
					}`),
				),
			},
			{
				// Refresh must not introduce drift: a second plan must be empty.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("repro"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
				PlanOnly:                 true,
			},
		},
	})
}
