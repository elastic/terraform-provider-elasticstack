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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func updateExceptionItem(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[ExceptionItemModel],
) (entitycore.KibanaWriteResult[ExceptionItemModel], diag.Diagnostics) {
	return entitycore.SimpleKibanaUpdate[ExceptionItemModel, kbapi.UpdateExceptionListItemJSONRequestBody, kbapi.SecurityExceptionsAPIExceptionListItem](
		func(plan ExceptionItemModel, ctx context.Context, writeID string) (kbapi.UpdateExceptionListItemJSONRequestBody, diag.Diagnostics) {
			commentsBody, diags := reconcileRequiredComments(ctx, client, req)
			if diags.HasError() {
				return kbapi.UpdateExceptionListItemJSONRequestBody{}, diags
			}

			body, d := plan.toUpdateRequest(ctx, writeID, commentsBody)
			diags.Append(d...)
			if diags.HasError() {
				return kbapi.UpdateExceptionListItemJSONRequestBody{}, diags
			}
			return *body, diags
		},
		// The update endpoint takes the item ID in the request body, not the URL.
		func(ctx context.Context, client *kibanaoapi.Client, spaceID, _ string, body kbapi.UpdateExceptionListItemJSONRequestBody) (*kbapi.SecurityExceptionsAPIExceptionListItem, diag.Diagnostics) {
			return kibanaoapi.UpdateExceptionListItem(ctx, client, spaceID, body)
		},
		(*ExceptionItemModel).populateUpdated,
	)(ctx, client, req)
}

// Even with no required texts, a failed fresh read must abort before PUT.
func reconcileRequiredComments(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[ExceptionItemModel],
) (*updateCommentsArray, diag.Diagnostics) {
	required, _ := requiredCommentTexts(req.Plan.RequiredComments)

	reconciler := newCommentReconciler(client, req.SpaceID, req.WriteID, req.Plan.NamespaceType)
	return reconciler.reconcile(ctx, required)
}

func newCommentReconciler(
	client *clients.KibanaScopedClient,
	spaceID, writeID string,
	namespaceType types.String,
) commentReconciler {
	return commentReconciler{
		readStoredComments: func(ctx context.Context) ([]kbapi.SecurityExceptionsAPIExceptionListItemComment, bool, diag.Diagnostics) {
			readResp, d := kibanaoapi.GetExceptionListItem(ctx, client.GetKibanaOapiClient(), spaceID, exceptionItemReadParams(writeID, namespaceType))
			if d.HasError() {
				return nil, false, d
			}
			if readResp == nil {
				return nil, false, nil
			}
			return readResp.Comments, true, nil
		},
		supportsMissingOnlyBody: func(ctx context.Context) (bool, diag.Diagnostics) {
			return client.EnforceVersionCheck(ctx, supportsMissingOnlyCommentBody)
		},
	}
}

func (m *ExceptionItemModel) populateUpdated(_ context.Context, _ string, updated *kbapi.SecurityExceptionsAPIExceptionListItem) diag.Diagnostics {
	var diags diag.Diagnostics
	entitycore.RequireNonNilKibanaWriteResponse(&diags, updated, "update", "exception item")
	return diags
}
