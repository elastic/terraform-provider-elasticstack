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

package snapshot_repository

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"maps"
)

func commonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"chunk_size":                 types.StringType,
		"compress":                   types.BoolType,
		"max_snapshot_bytes_per_sec": types.StringType,
		"max_restore_bytes_per_sec":  types.StringType,
		"readonly":                   types.BoolType,
	}
}

func commonStdAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"max_number_of_snapshots": types.Int64Type,
	}
}

func mergeAttrTypes(mps ...map[string]attr.Type) map[string]attr.Type {
	result := make(map[string]attr.Type)
	for _, m := range mps {
		maps.Copy(result, m)
	}
	return result
}

func fsAttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), commonStdAttrTypes(), map[string]attr.Type{
		"location": types.StringType,
	})
}

func urlAttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), commonStdAttrTypes(), map[string]attr.Type{
		"url":                 types.StringType,
		"http_max_retries":    types.Int64Type,
		"http_socket_timeout": types.StringType,
	})
}

func gcsAttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), map[string]attr.Type{
		"bucket":    types.StringType,
		"client":    types.StringType,
		"base_path": types.StringType,
	})
}

func azureAttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), map[string]attr.Type{
		"container":     types.StringType,
		"client":        types.StringType,
		"base_path":     types.StringType,
		"location_mode": types.StringType,
	})
}

func s3AttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), map[string]attr.Type{
		"bucket":                 types.StringType,
		"endpoint":               types.StringType,
		"client":                 types.StringType,
		"base_path":              types.StringType,
		"server_side_encryption": types.BoolType,
		"buffer_size":            types.StringType,
		"canned_acl":             types.StringType,
		"storage_class":          types.StringType,
		"path_style_access":      types.BoolType,
	})
}

func hdfsAttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), map[string]attr.Type{
		"uri":           types.StringType,
		"path":          types.StringType,
		"load_defaults": types.BoolType,
	})
}
