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

package kibanaoapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// kibanaStatusDTO is a minimal DTO for parsing the Kibana status response body.
type kibanaStatusDTO struct {
	Version struct {
		Number      string  `json:"number"`
		BuildFlavor *string `json:"build_flavor"`
	} `json:"version"`
}

// GetKibanaStatus calls the Kibana status API and returns the parsed version
// number and build flavor. BuildFlavor is an empty string when absent (i.e. on
// older stateful deployments that predate the serverless distinction).
func GetKibanaStatus(ctx context.Context, client *kbapi.ClientWithResponses) (versionNumber string, buildFlavor string, diags sdkdiag.Diagnostics) {
	resp, err := client.GetStatusWithResponse(ctx, &kbapi.GetStatusParams{})
	if err != nil {
		diags = append(diags, sdkdiag.Diagnostic{
			Severity: sdkdiag.Error,
			Summary:  "Failed to get Kibana status",
			Detail:   err.Error(),
		})
		return "", "", diags
	}

	if resp.StatusCode() != http.StatusOK {
		diags = reportUnknownErrorSDK(resp.StatusCode(), resp.Body)
		return "", "", diags
	}

	var dto kibanaStatusDTO
	if err := json.Unmarshal(resp.Body, &dto); err != nil {
		diags = append(diags, sdkdiag.Diagnostic{
			Severity: sdkdiag.Error,
			Summary:  "Failed to parse Kibana status response",
			Detail:   err.Error(),
		})
		return "", "", diags
	}

	if dto.Version.Number == "" {
		diags = append(diags, sdkdiag.Diagnostic{
			Severity: sdkdiag.Error,
			Summary:  "Failed to get version from Kibana status",
			Detail:   "The 'version.number' field was absent or empty in the Kibana status response.",
		})
		return "", "", diags
	}

	flavor := ""
	if dto.Version.BuildFlavor != nil {
		flavor = *dto.Version.BuildFlavor
	}

	return dto.Version.Number, flavor, nil
}
