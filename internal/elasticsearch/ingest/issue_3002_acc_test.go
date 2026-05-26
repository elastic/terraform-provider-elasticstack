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

package ingest_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// renameProcessorOverrideMinVersion is the lowest Elasticsearch version that
// accepts the `override` option on the rename ingest processor. Older versions
// reject the field with a 400, so the regression test is skipped against them.
var renameProcessorOverrideMinVersion = version.Must(version.NewVersion("8.13.0"))

// TestAccReproduceIssue3002 is the regression test for
// https://github.com/elastic/terraform-provider-elasticstack/issues/3002.
//
// A rename processor with `override = true` previously produced drift after
// apply because the typed go-elasticsearch client (types.RenameProcessor)
// has no Override field and silently dropped it during deserialization.
// The provider now decodes ingest pipeline responses opaquely so unmodeled
// processor fields survive the refresh.
func TestAccReproduceIssue3002(t *testing.T) {
	versionutils.SkipIfUnsupported(t, renameProcessorOverrideMinVersion, versionutils.FlavorAny)

	pipelineName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resourceName := "elasticstack_elasticsearch_ingest_pipeline.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("repro"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "processors.#", "1"),
					// regression for issue #3002: unmodeled processor fields must survive refresh
					CheckResourceJSON(resourceName, "processors.0", `{"rename":{"field":"tmp_source_field","target_field":"destination_field","override":true,"ignore_missing":true}}`),
				),
			},
			{
				// Second plan must be empty: a refresh must not introduce drift.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("repro"),
				ConfigVariables:          config.Variables{"name": config.StringVariable(pipelineName)},
				PlanOnly:                 true,
			},
		},
	})
}
