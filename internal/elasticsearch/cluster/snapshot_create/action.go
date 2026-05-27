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

package snapshot_create

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/action"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	_ action.Action              = (*snapshotCreateAction)(nil)
	_ action.ActionWithConfigure = (*snapshotCreateAction)(nil)
)

const defaultInvokeTimeout = 20 * time.Minute

type snapshotCreateAction struct {
	factory *clients.ProviderClientFactory
}

// NewCreateAction returns a constructor for the snapshot create action.
func NewCreateAction() action.Action {
	return &snapshotCreateAction{}
}

func (a *snapshotCreateAction) Metadata(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = "elasticstack_elasticsearch_snapshot_create"
}

func (a *snapshotCreateAction) Schema(ctx context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = GetSchema(ctx)
}

func (a *snapshotCreateAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	factory, diags := clients.ConvertProviderDataToFactory(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	a.factory = factory
}

func (a *snapshotCreateAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var model Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if a.factory == nil {
		resp.Diagnostics.AddError("Provider not configured", "Expected configured provider client factory")
		return
	}

	client, diags := a.factory.GetElasticsearchClient(ctx, model.ElasticsearchConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	invokeTimeout, timeoutDiags := model.Timeouts.Invoke(ctx, defaultInvokeTimeout)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, invokeTimeout)
	defer cancel()

	waitForCompletion := true
	if !model.WaitForCompletion.IsNull() && !model.WaitForCompletion.IsUnknown() {
		waitForCompletion = model.WaitForCompletion.ValueBool()
	}

	body, bodyDiags := createRequestFromModel(ctx, model)
	resp.Diagnostics.Append(bodyDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(esclient.CreateSnapshot(
		ctx,
		client,
		model.Repository.ValueString(),
		model.Snapshot.ValueString(),
		body,
		waitForCompletion,
	)...)
}

func createRequestFromModel(ctx context.Context, model Model) (*esclient.CreateSnapshotRequest, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics
	body := &esclient.CreateSnapshotRequest{}

	if !model.Indices.IsNull() && !model.Indices.IsUnknown() {
		var indices []string
		diags.Append(model.Indices.ElementsAs(ctx, &indices, false)...)
		if diags.HasError() {
			return nil, diags
		}
		body.Indices = indices
	}

	if !model.FeatureStates.IsNull() && !model.FeatureStates.IsUnknown() {
		var featureStates []string
		diags.Append(model.FeatureStates.ElementsAs(ctx, &featureStates, false)...)
		if diags.HasError() {
			return nil, diags
		}
		body.FeatureStates = featureStates
	}

	body.IgnoreUnavailable = typeutils.OptionalBool(model.IgnoreUnavailable)
	body.IncludeGlobalState = typeutils.OptionalBool(model.IncludeGlobalState)
	body.Partial = typeutils.OptionalBool(model.Partial)
	body.ExpandWildcards = typeutils.OptionalString(model.ExpandWildcards)

	if !model.Metadata.IsNull() && !model.Metadata.IsUnknown() {
		metaStr := model.Metadata.ValueString()
		if metaStr != "" {
			body.Metadata = json.RawMessage(metaStr)
		}
	}

	return body, diags
}
