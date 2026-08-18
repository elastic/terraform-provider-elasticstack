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

package securityexceptionitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                   = newExceptionItemResource()
	_ resource.ResourceWithConfigure      = newExceptionItemResource()
	_ resource.ResourceWithImportState    = newExceptionItemResource()
	_ resource.ResourceWithValidateConfig = newExceptionItemResource()
)

type ExceptionItemResource struct {
	*entitycore.KibanaResource[ExceptionItemModel]
}

func newExceptionItemResource() *ExceptionItemResource {
	return &ExceptionItemResource{
		KibanaResource: entitycore.NewKibanaResource[ExceptionItemModel](
			entitycore.ComponentKibana,
			"security_exception_item",
			entitycore.KibanaResourceOptions[ExceptionItemModel]{
				Schema: getSchema,
				Read: entitycore.SimpleKibanaReadWithParams[
					ExceptionItemModel,
					kbapi.ReadExceptionListItemParams,
					kbapi.SecurityExceptionsAPIExceptionListItem,
				](buildExceptionItemReadParams, getExceptionItemWithNamespaceRetry, (*ExceptionItemModel).fromAPI),
				Delete: entitycore.SimpleKibanaDeleteWithParams[ExceptionItemModel, kbapi.DeleteExceptionListItemParams](buildExceptionItemDeleteParams, kibanaoapi.DeleteExceptionListItem),
				Create: createExceptionItem,
				Update: updateExceptionItem,
			},
		),
	}
}

func buildExceptionItemReadParams(resourceID string, prior ExceptionItemModel) *kbapi.ReadExceptionListItemParams {
	id := resourceID
	params := &kbapi.ReadExceptionListItemParams{Id: &id}
	if typeutils.IsKnown(prior.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(prior.NamespaceType.ValueString())
		params.NamespaceType = &nsType
	}
	return params
}

// getExceptionItemWithNamespaceRetry retries with namespace_type=agnostic when
// namespace_type was not known up front (e.g. during import) and the initial
// lookup found nothing.
func getExceptionItemWithNamespaceRetry(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	params *kbapi.ReadExceptionListItemParams,
) (*kbapi.SecurityExceptionsAPIExceptionListItem, diag.Diagnostics) {
	return kibanaoapi.GetWithNamespaceRetry(ctx, client, spaceID, params, params.NamespaceType != nil, setExceptionItemAgnosticNamespaceType, kibanaoapi.GetExceptionListItem)
}

func setExceptionItemAgnosticNamespaceType(params *kbapi.ReadExceptionListItemParams) {
	agnostic := kbapi.SecurityExceptionsAPIExceptionNamespaceType("agnostic")
	params.NamespaceType = &agnostic
}

func buildExceptionItemDeleteParams(resourceID string, _ ExceptionItemModel) *kbapi.DeleteExceptionListItemParams {
	id := resourceID
	return &kbapi.DeleteExceptionListItemParams{Id: &id}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newExceptionItemResource()
}

func (r *ExceptionItemResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
