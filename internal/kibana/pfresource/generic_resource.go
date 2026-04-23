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

package pfresource

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceAPI[CreateRequest any, UpdateRequest any, Remote any] interface {
	Create(ctx context.Context, client *kibanaoapi.Client, spaceID string, request CreateRequest) (string, diag.Diagnostics)
	Get(ctx context.Context, client *kibanaoapi.Client, spaceID string, resourceID string) (Remote, bool, diag.Diagnostics)
	Update(ctx context.Context, client *kibanaoapi.Client, spaceID string, resourceID string, request UpdateRequest) diag.Diagnostics
	Delete(ctx context.Context, client *kibanaoapi.Client, spaceID string, resourceID string) diag.Diagnostics
}

type KibanaConnectionModel interface {
	GetKibanaConnection() types.List
}

type IDModel interface {
	GetID() types.String
	SetID(id types.String)
}

type SpaceIDModel interface {
	GetSpaceID() types.String
	SetSpaceID(spaceID types.String)
}

type ModelContract[CreateRequest any, UpdateRequest any, Remote any] interface {
	KibanaConnectionModel
	IDModel
	VersionRequirement() VersionRequirement
	ToCreateRequest(ctx context.Context) (CreateRequest, diag.Diagnostics)
	ToUpdateRequest(ctx context.Context) (UpdateRequest, diag.Diagnostics)
	PopulateFromRemote(ctx context.Context, spaceID string, remote Remote) diag.Diagnostics
}

type Assembly[
	CreateRequest any,
	UpdateRequest any,
	Remote any,
	Model ModelContract[CreateRequest, UpdateRequest, Remote],
] interface {
	TypeNameSuffix() string
	API() ResourceAPI[CreateRequest, UpdateRequest, Remote]
	NewModel() Model
	ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse)
}

type Orchestrator[
	CreateRequest any,
	UpdateRequest any,
	Remote any,
	Model ModelContract[CreateRequest, UpdateRequest, Remote],
] struct {
	Factory  *clients.ProviderClientFactory
	Assembly Assembly[CreateRequest, UpdateRequest, Remote, Model]
}

type ResolvedRuntime struct {
	ScopedClient *clients.KibanaScopedClient
	APIClient    *kibanaoapi.Client
}

func ResolveRuntime[
	CreateRequest any,
	UpdateRequest any,
	Remote any,
	Model ModelContract[CreateRequest, UpdateRequest, Remote],
](ctx context.Context, factory *clients.ProviderClientFactory, model Model) (*ResolvedRuntime, diag.Diagnostics) {
	scopedClient, diags := ResolveKibanaClient(ctx, factory, model.GetKibanaConnection())
	if diags.HasError() {
		return nil, diags
	}

	if versionDiags := EnforceVersion(ctx, scopedClient, model.VersionRequirement()); versionDiags.HasError() {
		return nil, versionDiags
	}

	apiClient, err := scopedClient.GetKibanaOapiClient()
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to create Kibana client", err.Error())}
	}

	return &ResolvedRuntime{ScopedClient: scopedClient, APIClient: apiClient}, nil
}

func ReadAfterWrite[
	CreateRequest any,
	UpdateRequest any,
	Remote any,
](
	ctx context.Context,
	api ResourceAPI[CreateRequest, UpdateRequest, Remote],
	client *kibanaoapi.Client,
	spaceID string,
	resourceID string,
) (Remote, diag.Diagnostics) {
	remote, present, diags := api.Get(ctx, client, spaceID, resourceID)
	if diags.HasError() {
		var zero Remote
		return zero, diags
	}
	if !present {
		var zero Remote
		return zero, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Resource not found after write",
			"The resource could not be read back from Kibana after a successful write.",
		)}
	}
	return remote, nil
}

func ReadRemote[
	CreateRequest any,
	UpdateRequest any,
	Remote any,
](
	ctx context.Context,
	api ResourceAPI[CreateRequest, UpdateRequest, Remote],
	client *kibanaoapi.Client,
	spaceID string,
	resourceID string,
) (Remote, bool, diag.Diagnostics) {
	return api.Get(ctx, client, spaceID, resourceID)
}

func (o Orchestrator[CreateRequest, UpdateRequest, Remote, Model]) Create(
	ctx context.Context,
	model Model,
	spaceID string,
) (Model, diag.Diagnostics) {
	runtime, diags := ResolveRuntime[CreateRequest, UpdateRequest, Remote](ctx, o.Factory, model)
	if diags.HasError() {
		return model, diags
	}

	request, reqDiags := model.ToCreateRequest(ctx)
	if reqDiags.HasError() {
		return model, reqDiags
	}

	resourceID, createDiags := o.Assembly.API().Create(ctx, runtime.APIClient, spaceID, request)
	if createDiags.HasError() {
		return model, createDiags
	}

	remote, readDiags := ReadAfterWrite(ctx, o.Assembly.API(), runtime.APIClient, spaceID, resourceID)
	if readDiags.HasError() {
		return model, readDiags
	}

	model.SetID(types.StringValue(resourceID))
	if spaceAware, ok := any(model).(SpaceIDModel); ok {
		spaceAware.SetSpaceID(types.StringValue(spaceID))
	}
	populateDiags := model.PopulateFromRemote(ctx, spaceID, remote)
	return model, populateDiags
}

func (o Orchestrator[CreateRequest, UpdateRequest, Remote, Model]) Read(
	ctx context.Context,
	model Model,
	spaceID string,
) (Model, bool, diag.Diagnostics) {
	runtime, diags := ResolveRuntime[CreateRequest, UpdateRequest, Remote](ctx, o.Factory, model)
	if diags.HasError() {
		return model, false, diags
	}

	remote, present, readDiags := ReadRemote(
		ctx,
		o.Assembly.API(),
		runtime.APIClient,
		spaceID,
		model.GetID().ValueString(),
	)
	if readDiags.HasError() {
		return model, false, readDiags
	}
	if !present {
		return model, false, nil
	}

	if spaceAware, ok := any(model).(SpaceIDModel); ok {
		spaceAware.SetSpaceID(types.StringValue(spaceID))
	}
	populateDiags := model.PopulateFromRemote(ctx, spaceID, remote)
	return model, true, populateDiags
}

func (o Orchestrator[CreateRequest, UpdateRequest, Remote, Model]) Update(
	ctx context.Context,
	model Model,
	spaceID string,
) (Model, diag.Diagnostics) {
	runtime, diags := ResolveRuntime[CreateRequest, UpdateRequest, Remote](ctx, o.Factory, model)
	if diags.HasError() {
		return model, diags
	}

	request, reqDiags := model.ToUpdateRequest(ctx)
	if reqDiags.HasError() {
		return model, reqDiags
	}

	updateDiags := o.Assembly.API().Update(ctx, runtime.APIClient, spaceID, model.GetID().ValueString(), request)
	if updateDiags.HasError() {
		return model, updateDiags
	}

	remote, readDiags := ReadAfterWrite(
		ctx,
		o.Assembly.API(),
		runtime.APIClient,
		spaceID,
		model.GetID().ValueString(),
	)
	if readDiags.HasError() {
		return model, readDiags
	}

	if spaceAware, ok := any(model).(SpaceIDModel); ok {
		spaceAware.SetSpaceID(types.StringValue(spaceID))
	}
	populateDiags := model.PopulateFromRemote(ctx, spaceID, remote)
	return model, populateDiags
}

func (o Orchestrator[CreateRequest, UpdateRequest, Remote, Model]) Delete(
	ctx context.Context,
	model Model,
	spaceID string,
) diag.Diagnostics {
	runtime, diags := ResolveRuntime[CreateRequest, UpdateRequest, Remote](ctx, o.Factory, model)
	if diags.HasError() {
		return diags
	}
	return o.Assembly.API().Delete(ctx, runtime.APIClient, spaceID, model.GetID().ValueString())
}
