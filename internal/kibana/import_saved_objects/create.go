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

package importsavedobjects

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	r.importObjects(ctx, request.Plan, &response.State, &response.Diagnostics)
}

func (r *Resource) importObjects(ctx context.Context, plan tfsdk.Plan, state *tfsdk.State, diags *diag.Diagnostics) {
	var model modelV0

	diags.Append(plan.Get(ctx, &model)...)
	if diags.HasError() {
		return
	}

	client, clientDiags := r.Client().GetKibanaClient(ctx, model.KibanaConnection)
	diags.Append(clientDiags...)
	if diags.HasError() {
		return
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("unable to get Kibana OpenAPI client", err.Error())
		return
	}

	params := kbapi.PostSavedObjectsImportParams{}
	if typeutils.IsKnown(model.Overwrite) && model.Overwrite.ValueBool() {
		v := true
		params.Overwrite = &v
	}
	if typeutils.IsKnown(model.CreateNewCopies) && model.CreateNewCopies.ValueBool() {
		v := true
		params.CreateNewCopies = &v
	}
	if typeutils.IsKnown(model.CompatibilityMode) && model.CompatibilityMode.ValueBool() {
		v := true
		params.CompatibilityMode = &v
	}

	result, importDiags := kibanaoapi.ImportSavedObjects(ctx, oapiClient, model.SpaceID.ValueString(), []byte(model.FileContents.ValueString()), params)
	diags.Append(importDiags...)
	if diags.HasError() {
		return
	}

	if model.ID.IsUnknown() {
		model.ID = types.StringValue(uuid.NewString())
	}

	diags.Append(state.Set(ctx, model)...)
	diags.Append(state.SetAttribute(ctx, path.Root("success"), result.Success)...)
	diags.Append(state.SetAttribute(ctx, path.Root("success_count"), result.SuccessCount)...)

	errors := mapImportErrors(result.Errors)
	diags.Append(state.SetAttribute(ctx, path.Root("errors"), errors)...)

	successResults := mapSuccessResults(result.SuccessResults)
	diags.Append(state.SetAttribute(ctx, path.Root("success_results"), successResults)...)

	if diags.HasError() {
		return
	}

	if !result.Success && (!typeutils.IsKnown(model.IgnoreImportErrors) || !model.IgnoreImportErrors.ValueBool()) {
		var detail strings.Builder
		for i, e := range errors {
			fmt.Fprintf(&detail, "import error [%d]: %s\n", i, e)
		}
		detail.WriteString("see the `errors` attribute for the full response")

		if result.SuccessCount > 0 {
			diags.AddWarning(
				"not all objects were imported successfully",
				detail.String(),
			)
		} else {
			diags.AddError(
				"no objects imported successfully",
				detail.String(),
			)
		}
	}
}

// mapImportErrors converts the raw map slice from the API response into typed importError structs.
func mapImportErrors(raw []map[string]any) []importError {
	result := make([]importError, 0, len(raw))
	for _, m := range raw {
		ie := importError{}
		if v, ok := m["id"].(string); ok {
			ie.ID = v
		}
		if v, ok := m["type"].(string); ok {
			ie.Type = v
		}
		if v, ok := m["title"].(string); ok {
			ie.Title = v
		}
		if errMap, ok := m["error"].(map[string]any); ok {
			if v, ok := errMap["type"].(string); ok {
				ie.Error = importErrorType{Type: v}
			}
		}
		if metaMap, ok := m["meta"].(map[string]any); ok {
			meta := importMeta{}
			if v, ok := metaMap["icon"].(string); ok {
				meta.Icon = v
			}
			if v, ok := metaMap["title"].(string); ok {
				meta.Title = v
			}
			ie.Meta = meta
		}
		result = append(result, ie)
	}
	return result
}

// mapSuccessResults converts the raw map slice from the API response into typed importSuccess structs.
func mapSuccessResults(raw []map[string]any) []importSuccess {
	result := make([]importSuccess, 0, len(raw))
	for _, m := range raw {
		is := importSuccess{}
		if v, ok := m["id"].(string); ok {
			is.ID = v
		}
		if v, ok := m["type"].(string); ok {
			is.Type = v
		}
		if v, ok := m["destinationId"].(string); ok {
			is.DestinationID = v
		}
		if metaMap, ok := m["meta"].(map[string]any); ok {
			meta := importMeta{}
			if v, ok := metaMap["icon"].(string); ok {
				meta.Icon = v
			}
			if v, ok := metaMap["title"].(string); ok {
				meta.Title = v
			}
			is.Meta = meta
		}
		result = append(result, is)
	}
	return result
}

type importSuccess struct {
	ID            string     `tfsdk:"id" json:"id"`
	Type          string     `tfsdk:"type" json:"type"`
	DestinationID string     `tfsdk:"destination_id" json:"destinationId"`
	Meta          importMeta `tfsdk:"meta" json:"meta"`
}

type importError struct {
	ID    string          `tfsdk:"id" json:"id"`
	Type  string          `tfsdk:"type" json:"type"`
	Title string          `tfsdk:"title" json:"title"`
	Error importErrorType `tfsdk:"error" json:"error"`
	Meta  importMeta      `tfsdk:"meta" json:"meta"`
}

func (ie importError) String() string {
	title := ie.Title
	if title == "" {
		title = ie.Meta.Title
	}

	return fmt.Sprintf("[%s] error on [%s] with ID [%s] and title [%s]", ie.Error.Type, ie.Type, ie.ID, title)
}

type importErrorType struct {
	Type string `tfsdk:"type" json:"type"`
}

type importMeta struct {
	Icon  string `tfsdk:"icon" json:"icon"`
	Title string `tfsdk:"title" json:"title"`
}
