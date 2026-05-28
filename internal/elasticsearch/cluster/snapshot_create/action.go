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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/action"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

const defaultInvokeTimeout = 20 * time.Minute

// NewCreateAction returns a constructor for the snapshot create action. The
// Configure, Metadata, Schema, and Invoke prelude are owned by the
// [entitycore] action envelope; this package supplies only the schema body
// and the invoke callback.
func NewCreateAction() action.Action {
	return entitycore.NewElasticsearchAction[Model]("snapshot_create", entitycore.ElasticsearchActionOptions[Model]{
		Schema:               GetSchema,
		Invoke:               invokeCreate,
		DefaultInvokeTimeout: defaultInvokeTimeout,
	})
}

// invokeCreate is the entity-specific work for elasticstack_elasticsearch_snapshot_create.
// The envelope has already decoded req.Config, resolved client, and applied
// the invoke timeout to ctx.
func invokeCreate(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.ActionRequest[Model]) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	model := req.Config

	waitForCompletion := true
	if !model.WaitForCompletion.IsNull() && !model.WaitForCompletion.IsUnknown() {
		waitForCompletion = model.WaitForCompletion.ValueBool()
	}

	body, bodyDiags := createRequestFromModel(ctx, model)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return diags
	}

	diags.Append(esclient.CreateSnapshot(
		ctx,
		client,
		model.Repository.ValueString(),
		model.Snapshot.ValueString(),
		body,
		waitForCompletion,
	)...)
	return diags
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
