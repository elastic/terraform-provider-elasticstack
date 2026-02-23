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

package apikey

import "github.com/elastic/terraform-provider-elasticstack/internal/models"

// populateRoleDescriptorsDefaults ensures that all role descriptors have proper defaults
func populateRoleDescriptorsDefaults(model map[string]models.APIKeyRoleDescriptor) map[string]models.APIKeyRoleDescriptor {
	for role, descriptor := range model {
		resultDescriptor := descriptor

		// Ensure AllowRestrictedIndices is set to false for all indices that don't have it set
		for i, index := range resultDescriptor.Indices {
			if index.AllowRestrictedIndices == nil {
				resultDescriptor.Indices[i].AllowRestrictedIndices = new(bool)
				*resultDescriptor.Indices[i].AllowRestrictedIndices = false
			}
		}

		model[role] = resultDescriptor
	}

	return model
}
