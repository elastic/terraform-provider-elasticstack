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

package output

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/stateutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	MinVersionOutputKafka               = version.Must(version.NewVersion("8.13.0"))
	MinVersionOutputSSLVerificationMode = version.Must(version.NewVersion("8.10.0"))
)

var (
	_ resource.Resource                 = newOutputResource()
	_ resource.ResourceWithConfigure    = newOutputResource()
	_ resource.ResourceWithImportState  = newOutputResource()
	_ resource.ResourceWithUpgradeState = newOutputResource()
)

type outputResource struct {
	*entitycore.KibanaResource[outputModel]
	*fleet.SpaceImporter
}

func newOutputResource() *outputResource {
	return &outputResource{
		KibanaResource: entitycore.NewKibanaResource[outputModel](
			entitycore.ComponentFleet,
			"output",
			entitycore.KibanaResourceOptions[outputModel]{
				Schema: getSchema,
				Read:   readOutput,
				Delete: deleteOutput,
				Create: createOutput,
				Update: updateOutput,
			},
		),
		SpaceImporter: fleet.NewSpaceImporter(path.Root("output_id")),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newOutputResource()
}

func (r *outputResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			// Legacy provider versions used a block for the `ssl` attribute which means it was stored as a list.
			// This upgrader migrates the list into a single object if available within the raw state
			StateUpgrader: func(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				// Default to returning the original state if no changes are needed
				if req.RawState != nil && req.RawState.JSON != nil {
					resp.DynamicValue = &tfprotov6.DynamicValue{JSON: req.RawState.JSON}
				}

				stateMap := stateutil.UnmarshalStateMap(req, resp)
				if resp.Diagnostics.HasError() {
					return
				}

				sslInterface, ok := stateMap["ssl"]
				if !ok {
					return
				}

				sslList, ok := sslInterface.([]any)
				if !ok {
					resp.Diagnostics.AddAttributeError(path.Root("ssl"),
						"Unexpected type for legacy ssl attribute",
						fmt.Sprintf("Expected []any, got %T", sslInterface),
					)
					return
				}

				if len(sslList) > 0 {
					stateMap["ssl"] = sslList[0]
				} else {
					delete(stateMap, "ssl")
				}

				stateutil.MarshalStateMap(stateMap, resp)
			},
		},
	}
}
