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

package checks

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestCheckResourceAttrJSONSubset checks that the JSON value of the specified
// resource attribute contains all keys and values from the expected JSON subset.
// Object keys are matched regardless of order; nested objects and arrays are
// compared recursively.
func TestCheckResourceAttrJSONSubset(name, key, expectedJSON string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("no primary instance: %s", name)
		}

		actualRaw, ok := is.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: attribute %s not found", name, key)
		}

		var actual, expected any
		if err := json.Unmarshal([]byte(actualRaw), &actual); err != nil {
			return fmt.Errorf("%s: failed to unmarshal actual JSON for %s: %w", name, key, err)
		}
		if err := json.Unmarshal([]byte(expectedJSON), &expected); err != nil {
			return fmt.Errorf("failed to unmarshal expected JSON: %w", err)
		}

		if err := jsonSubset(actual, expected, ""); err != nil {
			return fmt.Errorf("%s: JSON subset check failed for %s: %w\nactual: %s", name, key, err, actualRaw)
		}

		return nil
	}
}

func jsonSubset(actual, expected any, path string) error {
	if expected == nil {
		return nil
	}

	if reflect.DeepEqual(actual, expected) {
		return nil
	}

	switch exp := expected.(type) {
	case map[string]any:
		act, ok := actual.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object, got %T", path, actual)
		}
		for k, v := range exp {
			subPath := k
			if path != "" {
				subPath = path + "." + k
			}
			av, ok := act[k]
			if !ok {
				return fmt.Errorf("%s: missing key %q", path, k)
			}
			if err := jsonSubset(av, v, subPath); err != nil {
				return err
			}
		}
	case []any:
		act, ok := actual.([]any)
		if !ok {
			return fmt.Errorf("%s: expected array, got %T", path, actual)
		}
		if len(act) < len(exp) {
			return fmt.Errorf("%s: expected array of length >= %d, got %d", path, len(exp), len(act))
		}
		for i, v := range exp {
			subPath := fmt.Sprintf("%s[%d]", path, i)
			if err := jsonSubset(act[i], v, subPath); err != nil {
				return err
			}
		}
	default:
		if actual != expected {
			return fmt.Errorf("%s: expected %v, got %v", path, expected, actual)
		}
	}

	return nil
}
