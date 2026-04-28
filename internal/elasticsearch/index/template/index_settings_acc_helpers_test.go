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
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccCheckResourceAttrIndexSettingsSemantic asserts template.settings matches the expected
// effective index settings JSON using the same rules as DiffIndexSettingSuppress /
// IndexSettingsValue.SemanticallyEqual.
func testAccCheckResourceAttrIndexSettingsSemantic(addr, want string) resource.TestCheckFunc {
	const attr = "template.settings"
	return func(s *terraform.State) error {
		ctx := context.Background()
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("resource not found: %s", addr)
		}
		got, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("%s: attribute %q not found in state", addr, attr)
		}
		a := customtypes.NewIndexSettingsValue(want)
		b := customtypes.NewIndexSettingsValue(got)
		eq, diags := a.SemanticallyEqual(ctx, b)
		if diags.HasError() {
			return fmt.Errorf("%s: %v", addr, diags)
		}
		if !eq {
			return fmt.Errorf("%s: %s = %q, expected semantically equivalent to %q", addr, attr, got, want)
		}
		return nil
	}
}
