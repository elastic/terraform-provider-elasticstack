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

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ResourceAPI[CreateRequest any, UpdateRequest any, Remote any] interface {
	Create(context.Context, *kibanaoapi.Client, string, CreateRequest) (string, diag.Diagnostics)
	Get(context.Context, *kibanaoapi.Client, string, string) (Remote, diag.Diagnostics)
	Update(context.Context, *kibanaoapi.Client, string, string, UpdateRequest) diag.Diagnostics
	Delete(context.Context, *kibanaoapi.Client, string, string) diag.Diagnostics
}

type ModelContract[CreateRequest any, UpdateRequest any, Remote any] interface {
	VersionRequirement() VersionRequirement
	ToCreateRequest(context.Context) (CreateRequest, diag.Diagnostics)
	ToUpdateRequest(context.Context) (UpdateRequest, diag.Diagnostics)
	PopulateFromRemote(context.Context, string, Remote) diag.Diagnostics
}

type Assembly interface {
	TypeNameSuffix() string
}

func ReadAfterWrite[Remote any](ctx context.Context, api ResourceAPI[any, any, Remote], client *kibanaoapi.Client, spaceID string, resourceID string) (Remote, diag.Diagnostics) {
	return api.Get(ctx, client, spaceID, resourceID)
}

func ReadRemote[Remote any](ctx context.Context, api ResourceAPI[any, any, Remote], client *kibanaoapi.Client, spaceID string, resourceID string) (Remote, bool, diag.Diagnostics) {
	remote, diags := api.Get(ctx, client, spaceID, resourceID)
	if diags.HasError() {
		var zero Remote
		return zero, false, diags
	}
	if isNil(remote) {
		var zero Remote
		return zero, false, nil
	}
	return remote, true, nil
}

func isNil[T any](v T) bool {
	var zero T
	return any(v) == any(zero)
}
