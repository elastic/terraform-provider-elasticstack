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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/action"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

const defaultInvokeTimeout = 20 * time.Minute

// NewRestoreAction returns a constructor for the snapshot restore action.
// The Configure, Metadata, Schema, and Invoke prelude are owned by the
// [entitycore] action envelope; this package supplies only the schema body
// and the invoke callback.
func NewRestoreAction() action.Action {
	return entitycore.NewElasticsearchAction[Model]("snapshot_restore", entitycore.ElasticsearchActionOptions[Model]{
		Schema:               GetSchema,
		Invoke:               invokeRestore,
		DefaultInvokeTimeout: defaultInvokeTimeout,
	})
}

// invokeRestore is the entity-specific work for elasticstack_elasticsearch_snapshot_restore.
// The envelope has already decoded req.Config, resolved the client, and
// applied the invoke timeout to ctx.
func invokeRestore(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.ActionRequest[Model]) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	model := req.Config

	waitForCompletion := true
	if typeutils.IsKnown(model.WaitForCompletion) {
		waitForCompletion = model.WaitForCompletion.ValueBool()
	}

	body, bodyDiags := restoreRequestFromModel(ctx, model)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return diags
	}

	diags.Append(esclient.RestoreSnapshot(
		ctx,
		client,
		model.Repository.ValueString(),
		model.Snapshot.ValueString(),
		body,
		waitForCompletion,
	)...)
	return diags
}

func restoreRequestFromModel(ctx context.Context, model Model) (*esclient.RestoreSnapshotRequest, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics
	body := &esclient.RestoreSnapshotRequest{}

	if typeutils.IsKnown(model.Indices) {
		var indices []string
		diags.Append(model.Indices.ElementsAs(ctx, &indices, false)...)
		if diags.HasError() {
			return nil, diags
		}
		body.Indices = indices
	}

	if typeutils.IsKnown(model.FeatureStates) {
		var featureStates []string
		diags.Append(model.FeatureStates.ElementsAs(ctx, &featureStates, false)...)
		if diags.HasError() {
			return nil, diags
		}
		body.FeatureStates = featureStates
	}

	if typeutils.IsKnown(model.IgnoreIndexSettings) {
		var ignoreSettings []string
		diags.Append(model.IgnoreIndexSettings.ElementsAs(ctx, &ignoreSettings, false)...)
		if diags.HasError() {
			return nil, diags
		}
		body.IgnoreIndexSettings = ignoreSettings
	}

	body.IgnoreUnavailable = typeutils.OptionalBool(model.IgnoreUnavailable)
	body.IncludeGlobalState = typeutils.OptionalBool(model.IncludeGlobalState)
	body.IncludeAliases = typeutils.OptionalBool(model.IncludeAliases)
	body.Partial = typeutils.OptionalBool(model.Partial)
	body.RenamePattern = typeutils.OptionalString(model.RenamePattern)
	body.RenameReplacement = typeutils.OptionalString(model.RenameReplacement)

	if typeutils.IsKnown(model.IndexSettings) {
		settingsStr := model.IndexSettings.ValueString()
		if settingsStr != "" {
			body.IndexSettings = json.RawMessage(settingsStr)
		}
	}

	return body, diags
}
