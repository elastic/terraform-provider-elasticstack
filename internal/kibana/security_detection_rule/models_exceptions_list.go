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

package securitydetectionrule

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Helper function to process exceptions list configuration for all rule types
func (d Data) exceptionsListToAPI(ctx context.Context) ([]kbapi.SecurityDetectionsAPIRuleExceptionList, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiExceptionsList := convertListFieldToAPI(ctx, d.ExceptionsList, path.Root("exceptions_list"), &diags,
		func(exception ExceptionsListModel, _ typeutils.ListMeta) kbapi.SecurityDetectionsAPIRuleExceptionList {
			return kbapi.SecurityDetectionsAPIRuleExceptionList{
				Id:            exception.ID.ValueString(),
				ListId:        exception.ListID.ValueString(),
				NamespaceType: kbapi.SecurityDetectionsAPIRuleExceptionListNamespaceType(exception.NamespaceType.ValueString()),
				Type:          kbapi.SecurityDetectionsAPIExceptionListType(exception.Type.ValueString()),
			}
		})

	// Filter out empty exceptions (where required fields were null)
	validExceptions := make([]kbapi.SecurityDetectionsAPIRuleExceptionList, 0)
	for _, exception := range apiExceptionsList {
		if exception.Id != "" && exception.ListId != "" {
			validExceptions = append(validExceptions, exception)
		}
	}

	return validExceptions, diags
}

// convertExceptionsListToModel converts kbapi.SecurityDetectionsAPIRuleExceptionList slice to Terraform model
func convertExceptionsListToModel(ctx context.Context, apiExceptionsList []kbapi.SecurityDetectionsAPIRuleExceptionList) (types.List, diag.Diagnostics) {
	return convertAPISliceToListField(ctx, apiExceptionsList, getExceptionsListElementType(),
		func(apiException kbapi.SecurityDetectionsAPIRuleExceptionList) ExceptionsListModel {
			return ExceptionsListModel{
				ID:            types.StringValue(apiException.Id),
				ListID:        types.StringValue(apiException.ListId),
				NamespaceType: types.StringValue(string(apiException.NamespaceType)),
				Type:          types.StringValue(string(apiException.Type)),
			}
		})
}

func (d *Data) updateExceptionsListFromAPI(ctx context.Context, exceptionsList []kbapi.SecurityDetectionsAPIRuleExceptionList) diag.Diagnostics {
	var diags diag.Diagnostics
	d.ExceptionsList, diags = updateListFieldFromAPI(ctx, exceptionsList,
		types.ListNull(getExceptionsListElementType()),
		convertExceptionsListToModel)
	return diags
}
