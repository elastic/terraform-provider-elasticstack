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

package snapshot_restore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/action"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ action.Action               = (*snapshotRestoreAction)(nil)
	_ action.ActionWithConfigure  = (*snapshotRestoreAction)(nil)
)

const defaultInvokeTimeout = 20 * time.Minute

type snapshotRestoreAction struct {
	factory *clients.ProviderClientFactory
}

// NewRestoreAction returns a constructor for the snapshot restore action.
func NewRestoreAction() action.Action {
	return &snapshotRestoreAction{}
}

func (a *snapshotRestoreAction) Metadata(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = "elasticstack_elasticsearch_snapshot_restore"
}

func (a *snapshotRestoreAction) Schema(ctx context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = GetSchema(ctx)
}

func (a *snapshotRestoreAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

func (a *snapshotRestoreAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var model Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
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

	body, bodyDiags := restoreRequestFromModel(ctx, model)
	resp.Diagnostics.Append(bodyDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(esclient.RestoreSnapshot(
		ctx,
		client,
		model.Repository.ValueString(),
		model.Snapshot.ValueString(),
		body,
		waitForCompletion,
	)...)
}

func restoreRequestFromModel(ctx context.Context, model Model) (*esclient.RestoreSnapshotRequest, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics
	body := &esclient.RestoreSnapshotRequest{}

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

	if !model.IgnoreIndexSettings.IsNull() && !model.IgnoreIndexSettings.IsUnknown() {
		var ignoreSettings []string
		diags.Append(model.IgnoreIndexSettings.ElementsAs(ctx, &ignoreSettings, false)...)
		if diags.HasError() {
			return nil, diags
		}
		body.IgnoreIndexSettings = ignoreSettings
	}

	body.IgnoreUnavailable = optionalBool(model.IgnoreUnavailable)
	body.IncludeGlobalState = optionalBool(model.IncludeGlobalState)
	body.IncludeAliases = optionalBool(model.IncludeAliases)
	body.Partial = optionalBool(model.Partial)
	body.RenamePattern = optionalString(model.RenamePattern)
	body.RenameReplacement = optionalString(model.RenameReplacement)

	if !model.IndexSettings.IsNull() && !model.IndexSettings.IsUnknown() {
		settingsStr := model.IndexSettings.ValueString()
		if settingsStr != "" {
			body.IndexSettings = json.RawMessage(settingsStr)
		}
	}

	return body, diags
}

func optionalBool(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueBool()
	return &v
}

func optionalString(value types.String) *string {
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return nil
	}
	v := value.ValueString()
	return &v
}
