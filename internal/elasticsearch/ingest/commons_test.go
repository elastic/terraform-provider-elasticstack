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
	"fmt"

	schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// check if the provided json string equal to the generated one
func CheckResourceJSON(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", name, ms.Path)
		}
		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
		}

		v, ok := is.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: Attribute '%s' not found", name, key)
		}
		if eq, jsonErr := schemautil.JSONBytesEqual([]byte(value), []byte(v)); !eq {
			if jsonErr != nil {
				return fmt.Errorf(
					"%s: Attribute '%s' expected %#v, got %#v: %w",
					name,
					key,
					value,
					v,
					jsonErr,
				)
			}

			return fmt.Errorf(
				"%s: Attribute '%s' expected %#v, got %#v",
				name,
				key,
				value,
				v,
			)
		}
		return nil
	}
}
