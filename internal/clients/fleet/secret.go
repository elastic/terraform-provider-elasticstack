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

package fleet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const fleetSecretPath = "/_fleet/secret"

type fleetSecretResponse struct {
	ID string `json:"id"`
}

// CreateFleetSecret stores a secret value in Fleet secret storage and returns
// the generated secret reference id.
func CreateFleetSecret(ctx context.Context, es *elasticsearch.TypedClient, value string) (string, diag.Diagnostics) {
	if es == nil {
		return "", diag.Diagnostics{diag.NewErrorDiagnostic(
			"Elasticsearch client not configured",
			"Fleet secret creation requires an Elasticsearch client on the provider connection.",
		)}
	}

	body, err := json.Marshal(map[string]string{"value": value})
	if err != nil {
		return "", diag.Diagnostics{diag.NewErrorDiagnostic("Failed to encode Fleet secret request", err.Error())}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fleetSecretPath, bytes.NewReader(body))
	if err != nil {
		return "", diag.Diagnostics{diag.NewErrorDiagnostic("Failed to create Fleet secret request", err.Error())}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := es.Transport.Perform(req)
	if err != nil {
		return "", diag.Diagnostics{diag.NewErrorDiagnostic("Failed to create Fleet secret", err.Error())}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", diag.Diagnostics{diag.NewErrorDiagnostic("Failed to read Fleet secret response", err.Error())}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", diagutil.ReportUnknownHTTPError(resp.StatusCode, respBody)
	}

	var parsed fleetSecretResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", diag.Diagnostics{diag.NewErrorDiagnostic("Failed to decode Fleet secret response", err.Error())}
	}
	if parsed.ID == "" {
		return "", diag.Diagnostics{diag.NewErrorDiagnostic(
			"Failed to create Fleet secret",
			fmt.Sprintf("Fleet secret response did not include an id (HTTP %d).", resp.StatusCode),
		)}
	}

	return parsed.ID, nil
}
