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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func createOutput(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[outputModel]) (entitycore.KibanaWriteResult[outputModel], diag.Diagnostics) {
	planModel := req.Plan
	fleetClient := client.GetFleetClient()

	return entitycore.WriteEntity(
		func() (kbapi.NewOutputUnion, diag.Diagnostics) {
			return planModel.toAPICreateModel(ctx)
		},
		func(body kbapi.NewOutputUnion) (*kbapi.OutputUnion, diag.Diagnostics) {
			return fleet.CreateOutput(ctx, fleetClient, req.SpaceID, body)
		},
		func(output *kbapi.OutputUnion) (outputModel, diag.Diagnostics) {
			diags := planModel.populateFromAPI(ctx, output)
			return planModel, diags
		},
	)
}
