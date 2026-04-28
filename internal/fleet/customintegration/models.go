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

package customintegration

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type customIntegrationModel struct {
	ID                        types.String   `tfsdk:"id"`
	KibanaConnection          types.List     `tfsdk:"kibana_connection"`
	PackagePath               types.String   `tfsdk:"package_path"`
	PackageName               types.String   `tfsdk:"package_name"`
	PackageVersion            types.String   `tfsdk:"package_version"`
	Checksum                  types.String   `tfsdk:"checksum"`
	IgnoreMappingUpdateErrors types.Bool     `tfsdk:"ignore_mapping_update_errors"`
	SkipDataStreamRollover    types.Bool     `tfsdk:"skip_data_stream_rollover"`
	SkipDestroy               types.Bool     `tfsdk:"skip_destroy"`
	SpaceID                   types.String   `tfsdk:"space_id"`
	Timeouts                  timeouts.Value `tfsdk:"timeouts"`
}

func getPackageID(name string, version string) string {
	return fmt.Sprintf("%s/%s", name, version)
}
