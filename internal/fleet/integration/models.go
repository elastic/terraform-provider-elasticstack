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

package integration

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type integrationModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Version                   types.String `tfsdk:"version"`
	Force                     types.Bool   `tfsdk:"force"`
	Prerelease                types.Bool   `tfsdk:"prerelease"`
	IgnoreMappingUpdateErrors types.Bool   `tfsdk:"ignore_mapping_update_errors"`
	SkipDataStreamRollover    types.Bool   `tfsdk:"skip_data_stream_rollover"`
	IgnoreConstraints         types.Bool   `tfsdk:"ignore_constraints"`
	SkipDestroy               types.Bool   `tfsdk:"skip_destroy"`
	SpaceID                   types.String `tfsdk:"space_id"`
}

func getPackageID(name string, version string) string {
	hash, _ := schemautil.StringToHash(name + version)
	return *hash
}
