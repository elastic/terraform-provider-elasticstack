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

// TestNestedInTRun verifies resource.Test inside a t.Run closure is evaluated.
func TestNestedInTRun(t *testing.T) {
	t.Run("subtest", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("nested_create"),
				},
			},
		})
	})
}

// TestNestedInIf verifies resource.Test inside a conditional block is evaluated.
func TestNestedInIf(t *testing.T) {
	if true {
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("nested_if"),
				},
			},
		})
	}
}

// TestNestedInLoop verifies resource.Test inside a for loop is evaluated.
func TestNestedInLoop(t *testing.T) {
	for range 1 {
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					ProtoV6ProviderFactories: acctest.Providers,
					ConfigDirectory:          acctest.NamedTestCaseDirectory("nested_loop"),
				},
			},
		})
	}
}
