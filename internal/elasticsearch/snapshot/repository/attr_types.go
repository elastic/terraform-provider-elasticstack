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

package repository

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"maps"
)

func commonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		settingChunkSize:              types.StringType,
		settingCompress:               types.BoolType,
		settingMaxSnapshotBytesPerSec: types.StringType,
		settingMaxRestoreBytesPerSec:  types.StringType,
		settingReadonly:               types.BoolType,
	}
}

func commonStdAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		settingMaxNumberOfSnapshots: types.Int64Type,
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
		settingLocation: types.StringType,
	})
}

func urlAttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), commonStdAttrTypes(), map[string]attr.Type{
		settingURL:               types.StringType,
		settingHTTPMaxRetries:    types.Int64Type,
		settingHTTPSocketTimeout: types.StringType,
	})
}

func gcsAttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), map[string]attr.Type{
		settingBucket:   types.StringType,
		settingClient:   types.StringType,
		settingBasePath: types.StringType,
	})
}

func azureAttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), map[string]attr.Type{
		settingContainer:    types.StringType,
		settingClient:       types.StringType,
		settingBasePath:     types.StringType,
		settingLocationMode: types.StringType,
	})
}

func s3AttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), map[string]attr.Type{
		settingBucket:               types.StringType,
		settingEndpoint:             types.StringType,
		settingClient:               types.StringType,
		settingBasePath:             types.StringType,
		settingServerSideEncryption: types.BoolType,
		settingBufferSize:           types.StringType,
		settingCannedACL:            types.StringType,
		settingStorageClass:         types.StringType,
		settingPathStyleAccess:      types.BoolType,
	})
}

func hdfsAttrTypes() map[string]attr.Type {
	return mergeAttrTypes(commonAttrTypes(), map[string]attr.Type{
		settingURI:          types.StringType,
		settingPath:         types.StringType,
		settingLoadDefaults: types.BoolType,
	})
}
