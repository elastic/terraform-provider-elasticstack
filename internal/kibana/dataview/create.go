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

package dataview

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createDataView(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[dataViewModel],
) (entitycore.KibanaWriteResult[dataViewModel], diag.Diagnostics) {
	planModel := req.Plan
	var diags diag.Diagnostics

	body, bodyDiags := planModel.toAPICreateModel(ctx)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[dataViewModel]{}, diags
	}

	oapiClient, getDiags := client.GetKibanaOapiClient()
	diags.Append(getDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[dataViewModel]{}, diags
	}

	spaceID := req.SpaceID
	dataView, createDiags := createOrReconcileManagedDataView(ctx, oapiClient, spaceID, body)
	diags.Append(createDiags...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[dataViewModel]{}, diags
	}

	compositeID := clients.CompositeID{
		ClusterID:  spaceID,
		ResourceID: *dataView.DataView.Id,
	}
	planModel.ID = types.StringValue(compositeID.String())

	return entitycore.KibanaWriteResult[dataViewModel]{Model: planModel}, diags
}

func createOrReconcileManagedDataView(
	ctx context.Context,
	oapiClient *kibanaoapi.Client,
	spaceID string,
	body kbapi.DataViewsCreateDataViewRequestObject,
) (*kbapi.DataViewsDataViewResponseObject, diag.Diagnostics) {
	dataView, createDiags := kibanaoapi.CreateDataView(ctx, oapiClient, spaceID, body)
	if !createDiags.HasError() {
		return dataView, nil
	}

	if body.DataView.Id == nil || *body.DataView.Id == "" {
		return nil, createDiags
	}

	recoveredDataView, readDiags := kibanaoapi.GetDataView(ctx, oapiClient, spaceID, *body.DataView.Id)
	if readDiags.HasError() || recoveredDataView == nil {
		return nil, createDiags
	}

	return recoveredDataView, nil
}
