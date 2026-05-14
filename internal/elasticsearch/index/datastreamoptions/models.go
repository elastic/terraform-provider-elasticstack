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

package datastreamoptions

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model is the inner shape of template.data_stream_options.
type Model struct {
	FailureStore types.Object `tfsdk:"failure_store"`
}

// FailureStoreModel is the inner shape of template.data_stream_options.failure_store.
type FailureStoreModel struct {
	Enabled   types.Bool   `tfsdk:"enabled"`
	Lifecycle types.Object `tfsdk:"lifecycle"`
}

// FailureStoreLifecycleModel is the inner shape of failure_store.lifecycle.
type FailureStoreLifecycleModel struct {
	DataRetention types.String `tfsdk:"data_retention"`
}

// AttrTypes returns attribute types for template.data_stream_options.
func AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"failure_store": types.ObjectType{AttrTypes: FailureStoreAttrTypes()},
	}
}

// FailureStoreAttrTypes returns attribute types for failure_store.
func FailureStoreAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":   types.BoolType,
		"lifecycle": types.ObjectType{AttrTypes: FailureStoreLifecycleAttrTypes()},
	}
}

// FailureStoreLifecycleAttrTypes returns attribute types for failure_store.lifecycle.
func FailureStoreLifecycleAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"data_retention": types.StringType,
	}
}