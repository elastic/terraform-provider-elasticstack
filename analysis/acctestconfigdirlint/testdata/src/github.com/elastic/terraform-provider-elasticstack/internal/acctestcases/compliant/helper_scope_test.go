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

package compliant_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// helperWithNonCompliantResourceTest demonstrates the analyzer scope boundary:
// only Test-prefixed functions are entry points; helper functions that call
// resource.Test are intentionally not inspected.
func helperWithNonCompliantResourceTest(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: `resource "null_resource" "example" {}`,
			},
		},
	})
}

var _ func(*testing.T) = helperWithNonCompliantResourceTest

// TestHelperScopeBoundary is a minimal test that does not call resource.Test.
func TestHelperScopeBoundary(t *testing.T) {
	// This test exists to document that helperWithNonCompliantResourceTest is
	// not analyzed even though it contains violations.
}
