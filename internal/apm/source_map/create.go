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

package sourcemap

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/textproto"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *resourceSourceMap) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SourceMap
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scoped, fwDiags := r.Client().GetKibanaClient(ctx, plan.KibanaConnection)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kibana, err := scoped.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to get Kibana client", err.Error())
		return
	}

	// Resolve the source map bytes from plan attributes.
	var sourcemapBytes []byte
	switch {
	case !plan.SourcemapJSON.IsNull() && !plan.SourcemapJSON.IsUnknown():
		sourcemapBytes = []byte(plan.SourcemapJSON.ValueString())
	case !plan.SourcemapBinary.IsNull() && !plan.SourcemapBinary.IsUnknown():
		decoded, decErr := base64.StdEncoding.DecodeString(plan.SourcemapBinary.ValueString())
		if decErr != nil {
			resp.Diagnostics.AddError("Failed to decode sourcemap_binary", fmt.Sprintf("base64 decoding failed: %s", decErr.Error()))
			return
		}
		sourcemapBytes = decoded
	default:
		// This should never be reached if ExactlyOneOf validation is working correctly.
		resp.Diagnostics.AddError("No source map content provided", "Exactly one of sourcemap_json or sourcemap_binary must be set.")
		return
	}

	// Build the multipart/form-data body.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for _, field := range []struct{ name, value string }{
		{"bundle_filepath", plan.BundleFilepath.ValueString()},
		{"service_name", plan.ServiceName.ValueString()},
		{"service_version", plan.ServiceVersion.ValueString()},
	} {
		if writeErr := mw.WriteField(field.name, field.value); writeErr != nil {
			resp.Diagnostics.AddError("Failed to build multipart form", fmt.Sprintf("writing field %q: %s", field.name, writeErr.Error()))
			return
		}
	}

	// Write the sourcemap as a file field.
	fileContentType := "application/json"
	if !plan.SourcemapBinary.IsNull() && !plan.SourcemapBinary.IsUnknown() {
		fileContentType = "application/octet-stream"
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="sourcemap"; filename="sourcemap.js.map"`)
	h.Set("Content-Type", fileContentType)
	filePart, partErr := mw.CreatePart(h)
	if partErr != nil {
		resp.Diagnostics.AddError("Failed to build multipart form", fmt.Sprintf("creating file part: %s", partErr.Error()))
		return
	}
	if _, writeErr := filePart.Write(sourcemapBytes); writeErr != nil {
		resp.Diagnostics.AddError("Failed to build multipart form", fmt.Sprintf("writing sourcemap bytes: %s", writeErr.Error()))
		return
	}
	if closeErr := mw.Close(); closeErr != nil {
		resp.Diagnostics.AddError("Failed to build multipart form", fmt.Sprintf("closing writer: %s", closeErr.Error()))
		return
	}

	spaceID := plan.SpaceID.ValueString()

	apiResp, err := kibana.API.UploadSourceMapWithBodyWithResponse(
		ctx,
		&kbapi.UploadSourceMapParams{
			ElasticApiVersion: kbapi.UploadSourceMapParamsElasticApiVersionN20231031,
		},
		mw.FormDataContentType(),
		&buf,
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to upload APM source map", err.Error())
		return
	}

	if diags := diagutil.CheckHTTPErrorFromFW(apiResp.HTTPResponse, "Failed to upload APM source map"); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if apiResp.JSON200 == nil || apiResp.JSON200.Id == nil {
		resp.Diagnostics.AddError("Unexpected response from APM source map upload", "Expected a non-nil id in the upload response.")
		return
	}

	plan.ID = types.StringValue(*apiResp.JSON200.Id)

	updatedState, diags := r.read(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if updatedState == nil {
		tflog.Warn(ctx, "apm source map was uploaded but could not be read back immediately; storing plan state")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("Created APM source map with ID: %s", updatedState.ID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, updatedState)...)
}
