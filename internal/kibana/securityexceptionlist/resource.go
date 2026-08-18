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

package securityexceptionlist

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
	_ resource.Resource                = newExceptionListResource()
	_ resource.ResourceWithConfigure   = newExceptionListResource()
	_ resource.ResourceWithImportState = newExceptionListResource()
)

type ExceptionListResource struct {
	*entitycore.KibanaResource[ExceptionListModel]
}

func newExceptionListResource() *ExceptionListResource {
	return &ExceptionListResource{
		KibanaResource: entitycore.NewKibanaResource[ExceptionListModel](
			entitycore.ComponentKibana,
			"security_exception_list",
			entitycore.KibanaResourceOptions[ExceptionListModel]{
				Schema: getSchema,
				Read: entitycore.SimpleKibanaReadWithParams[
					ExceptionListModel,
					kbapi.ReadExceptionListParams,
					kbapi.SecurityExceptionsAPIExceptionList,
				](buildExceptionListReadParams, getExceptionListWithNamespaceRetry, (*ExceptionListModel).fromAPI),
				Delete: entitycore.SimpleKibanaDeleteWithParams[ExceptionListModel, kbapi.DeleteExceptionListParams](buildExceptionListDeleteParams, kibanaoapi.DeleteExceptionList),
				Create: createExceptionList,
				Update: updateExceptionList,
			},
		),
	}
}

func buildExceptionListReadParams(resourceID string, prior ExceptionListModel) *kbapi.ReadExceptionListParams {
	id := resourceID
	params := &kbapi.ReadExceptionListParams{Id: &id}
	if typeutils.IsKnown(prior.NamespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(prior.NamespaceType.ValueString())
		params.NamespaceType = &nsType
	}
	return params
}

// getExceptionListWithNamespaceRetry retries with namespace_type=agnostic when
// namespace_type was not known up front (e.g. during import) and the initial
// lookup found nothing.
func getExceptionListWithNamespaceRetry(
	ctx context.Context,
	client *kibanaoapi.Client,
	spaceID string,
	params *kbapi.ReadExceptionListParams,
) (*kbapi.SecurityExceptionsAPIExceptionList, diag.Diagnostics) {
	return kibanaoapi.GetWithNamespaceRetry(ctx, client, spaceID, params, params.NamespaceType != nil, setExceptionListAgnosticNamespaceType, kibanaoapi.GetExceptionList)
}

func setExceptionListAgnosticNamespaceType(params *kbapi.ReadExceptionListParams) {
	agnostic := kbapi.SecurityExceptionsAPIExceptionNamespaceType("agnostic")
	params.NamespaceType = &agnostic
}

func buildExceptionListDeleteParams(resourceID string, prior ExceptionListModel) *kbapi.DeleteExceptionListParams {
	id := resourceID
	params := &kbapi.DeleteExceptionListParams{Id: &id}
	if prior.NamespaceType.ValueString() != "" {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(prior.NamespaceType.ValueString())
		params.NamespaceType = &nsType
	}
	return params
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newExceptionListResource()
}

func (r *ExceptionListResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
