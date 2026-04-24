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

package customintegration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *customIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customIntegrationModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, fwDiags := plan.Timeouts.Create(ctx, 20*time.Minute)
	resp.Diagnostics.Append(fwDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	apiClient, diags := r.Client().GetKibanaClient(ctx, plan.KibanaConnection)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fleetClient, err := apiClient.GetFleetClient()
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	meetsMinVersion, verDiags := apiClient.EnforceMinVersion(ctx, minVersionCustomPackageGet)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(verDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !meetsMinVersion {
		resp.Diagnostics.AddError(
			"Kibana version not supported",
			"elasticstack_fleet_custom_integration requires Kibana 8.2.0 or later.",
		)
		return
	}

	filePath := plan.PackagePath.ValueString()
	contentType := detectContentType(filePath)

	result, diags := fleet.UploadPackage(ctx, fleetClient, fleet.UploadPackageOptions{
		PackagePath:               filePath,
		ContentType:               contentType,
		IgnoreMappingUpdateErrors: plan.IgnoreMappingUpdateErrors.ValueBool(),
		SkipDataStreamRollover:    plan.SkipDataStreamRollover.ValueBool(),
		SpaceID:                   plan.SpaceID.ValueString(),
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if result.PackageName == "" || result.PackageVersion == "" {
		resp.Diagnostics.AddError(
			"Package name or version could not be determined",
			"Fleet returned an empty package name or version. Ensure the archive contains "+
				"a valid manifest.yml with non-empty name and version fields.",
		)
		return
	}

	checksum, err := computeSHA256(filePath)
	if err != nil {
		resp.Diagnostics.AddError("Failed to compute checksum", err.Error())
		return
	}

	plan.PackageName = types.StringValue(result.PackageName)
	plan.PackageVersion = types.StringValue(result.PackageVersion)
	plan.Checksum = types.StringValue(checksum)
	plan.ID = types.StringValue(getPackageID(result.PackageName, result.PackageVersion))

	if plan.SpaceID.IsUnknown() {
		plan.SpaceID = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// detectContentType returns the MIME content type for the given file path
// based on its extension.
func detectContentType(filePath string) string {
	lower := strings.ToLower(filePath)
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") || strings.HasSuffix(lower, ".gz") {
		return "application/gzip"
	}
	// Default to zip (covers .zip and unknown extensions).
	return "application/zip"
}

// computeSHA256 returns the hex-encoded SHA256 digest of the file at filePath.
func computeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
