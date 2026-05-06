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
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
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

	sm := plan.Sourcemap
	if sm == nil {
		resp.Diagnostics.AddError("No source map content provided", "The sourcemap block is required.")
		return
	}

	var sourcemapBytes []byte
	var checksum string

	switch {
	case !sm.JSON.IsNull() && !sm.JSON.IsUnknown():
		sourcemapBytes = []byte(sm.JSON.ValueString())

	case !sm.Binary.IsNull() && !sm.Binary.IsUnknown():
		decoded, decErr := base64.StdEncoding.DecodeString(sm.Binary.ValueString())
		if decErr != nil {
			resp.Diagnostics.AddError("Failed to decode sourcemap binary", fmt.Sprintf("base64 decoding failed: %s", decErr.Error()))
			return
		}
		sourcemapBytes = decoded

	case sm.File != nil && !sm.File.Path.IsNull() && !sm.File.Path.IsUnknown():
		f, openErr := os.Open(sm.File.Path.ValueString())
		if openErr != nil {
			resp.Diagnostics.AddError("Failed to open source map file", openErr.Error())
			return
		}
		defer f.Close()
		h := sha256.New()
		var buf []byte
		raw, readErr := io.ReadAll(io.TeeReader(f, h))
		if readErr != nil {
			resp.Diagnostics.AddError("Failed to read source map file", readErr.Error())
			return
		}
		buf = raw
		sourcemapBytes = buf
		checksum = hex.EncodeToString(h.Sum(nil))

	default:
		resp.Diagnostics.AddError("No source map content provided", "Exactly one of sourcemap.json, sourcemap.binary, or sourcemap.file.path must be set.")
		return
	}

	artifactID, diags := kibanaoapi.UploadSourceMap(ctx, kibana, kibanaoapi.UploadSourceMapOptions{
		SpaceID:        plan.SpaceID.ValueString(),
		BundleFilepath: plan.BundleFilepath.ValueString(),
		ServiceName:    plan.ServiceName.ValueString(),
		ServiceVersion: plan.ServiceVersion.ValueString(),
		SourcemapBytes: sourcemapBytes,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(artifactID)

	if checksum != "" && sm.File != nil {
		sm.File.Checksum = types.StringValue(checksum)
	}

	updatedState, readDiags := r.read(ctx, &plan)
	resp.Diagnostics.Append(readDiags...)
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
