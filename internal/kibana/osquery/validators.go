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

package osquery

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// ECSMappingExactlyOneOfValidator returns an object validator that enforces exactly
// one of field/value/values per ecs_mapping element.
func ECSMappingExactlyOneOfValidator() validator.Object {
	return validators.ExactlyOneOfNestedAttrsValidator(validators.ExactlyOneOfNestedAttrsOpts{
		AttrNames:     []string{AttrECSMappingField, AttrECSMappingValue, AttrECSMappingValues},
		Summary:       "Invalid ecs_mapping element",
		MissingDetail: "Exactly one of `field`, `value`, or `values` must be set per `ecs_mapping` element.",
		TooManyDetail: "Exactly one of `field`, `value`, or `values` must be set per `ecs_mapping` element, not more than one.",
		Description:   "Ensures exactly one of `field`, `value`, or `values` is set on each `ecs_mapping` map value.",
	})
}
