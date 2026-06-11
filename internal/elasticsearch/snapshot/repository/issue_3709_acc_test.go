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

package repository_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue3709 reproduces the "Provider produced inconsistent result
// after apply" error that occurs when optional string settings (chunk_size,
// max_restore_bytes_per_sec, max_snapshot_bytes_per_sec) are explicitly set to ""
// in the gcs (or any) repository block.
//
// Root cause: write-side uses setIfNotEmpty, which omits "" from the API request,
// but read-side uses StrSettingNull, which returns null when the key is absent from
// the API response. This produces a "" → null drift that Terraform flags as an
// inconsistent post-apply result.
//
// Related to: https://github.com/elastic/terraform-provider-elasticstack/issues/3709
func TestAccReproduceIssue3709(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	// Use fs repository because GCS is unavailable in the local acceptance-test
	// stack, but the same setIfNotEmpty / StrSettingNull mismatch affects every
	// repository type including gcs (reported in the issue).
	cfg := fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_snapshot_repository" "test" {
  name = %q

  fs {
    location                   = "/tmp"
    compress                   = true
    # Setting optional string fields to "" triggers the bug: setIfNotEmpty omits
    # them from the API PUT, but StrSettingNull returns null on the subsequent GET,
    # causing a "" vs null inconsistency that Terraform reports as an error.
    chunk_size                 = ""
    max_restore_bytes_per_sec  = ""
    max_snapshot_bytes_per_sec = ""
  }
}
`, name)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkRepoDestroy(name),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   cfg,
				// The apply succeeds on the API side but the provider's post-apply
				// refresh reads null for the omitted keys, differing from the ""
				// values held in the plan. Terraform surfaces this as the
				// "inconsistent result after apply" error.
				ExpectError: regexp.MustCompile(`(?i)inconsistent result after apply`),
			},
		},
	})
}
