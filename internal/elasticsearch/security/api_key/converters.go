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

import (
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
)

// toTypedRoleDescriptors converts resource model role descriptors to typed API role descriptors.
func toTypedRoleDescriptors(descriptors map[string]models.APIKeyRoleDescriptor) (map[string]types.RoleDescriptor, error) {
	if descriptors == nil {
		return nil, nil
	}

	result := make(map[string]types.RoleDescriptor, len(descriptors))
	for key, descriptor := range descriptors {
		typedDescriptor, err := toTypedRoleDescriptor(descriptor)
		if err != nil {
			return nil, err
		}
		result[key] = typedDescriptor
	}
	return result, nil
}

// toTypedRoleDescriptor converts a single resource model role descriptor to a typed API role descriptor.
func toTypedRoleDescriptor(descriptor models.APIKeyRoleDescriptor) (types.RoleDescriptor, error) {
	jsonBytes, err := json.Marshal(descriptor)
	if err != nil {
		return types.RoleDescriptor{}, err
	}

	var typed types.RoleDescriptor
	if err := json.Unmarshal(jsonBytes, &typed); err != nil {
		return types.RoleDescriptor{}, err
	}

	return typed, nil
}

// toModelRoleDescriptors converts typed API role descriptors to resource model role descriptors.
func toModelRoleDescriptors(descriptors map[string]types.RoleDescriptor) (map[string]models.APIKeyRoleDescriptor, error) {
	if descriptors == nil {
		return nil, nil
	}

	result := make(map[string]models.APIKeyRoleDescriptor, len(descriptors))
	for key, descriptor := range descriptors {
		modelDescriptor, err := toModelRoleDescriptor(descriptor)
		if err != nil {
			return nil, err
		}
		result[key] = modelDescriptor
	}
	return result, nil
}

// toModelRoleDescriptor converts a single typed API role descriptor to a resource model role descriptor.
func toModelRoleDescriptor(descriptor types.RoleDescriptor) (models.APIKeyRoleDescriptor, error) {
	jsonBytes, err := json.Marshal(descriptor)
	if err != nil {
		return models.APIKeyRoleDescriptor{}, err
	}

	// The typed client represents "global" as []GlobalPrivilege (array),
	// but our model expects map[string]any (object). Unwrap the first
	// element of the array so the model can unmarshal it correctly.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		return models.APIKeyRoleDescriptor{}, err
	}
	if globalArray, ok := raw["global"]; ok && len(globalArray) > 0 && globalArray[0] == '[' {
		var globals []json.RawMessage
		if err := json.Unmarshal(globalArray, &globals); err == nil && len(globals) > 0 {
			raw["global"] = globals[0]
		}
	}
	fixedJSON, err := json.Marshal(raw)
	if err != nil {
		return models.APIKeyRoleDescriptor{}, err
	}

	var model models.APIKeyRoleDescriptor
	if err := json.Unmarshal(fixedJSON, &model); err != nil {
		return models.APIKeyRoleDescriptor{}, err
	}

	return model, nil
}

// toTypedMetadata converts a map of metadata to typed API metadata.
func toTypedMetadata(metadata map[string]any) (types.Metadata, error) {
	if metadata == nil {
		return nil, nil
	}

	result := make(types.Metadata, len(metadata))
	for key, value := range metadata {
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		result[key] = raw
	}
	return result, nil
}

// toModelMetadata converts typed API metadata to a map of metadata.
func toModelMetadata(metadata types.Metadata) (map[string]any, error) {
	if metadata == nil {
		return nil, nil
	}

	result := make(map[string]any, len(metadata))
	for key, raw := range metadata {
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

// toTypedAccess converts a resource model cross-cluster access to a typed API access.
func toTypedAccess(access models.CrossClusterAPIKeyAccess) (types.Access, error) {
	jsonBytes, err := json.Marshal(access)
	if err != nil {
		return types.Access{}, err
	}

	var typed types.Access
	if err := json.Unmarshal(jsonBytes, &typed); err != nil {
		return types.Access{}, err
	}

	return typed, nil
}
