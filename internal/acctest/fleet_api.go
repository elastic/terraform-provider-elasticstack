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

package acctest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
)

// FleetAPIRequest sends a raw request to the Kibana Fleet API, path relative to
// the Kibana endpoint (for example "/api/fleet/package_policies/<id>"). It uses
// the acceptance-test Fleet client, so it authenticates exactly like the
// provider. Tests use it to change Fleet objects outside Terraform, as a user
// would in the Kibana UI. It returns the response body and an error for a
// transport failure or a non-2xx status.
func FleetAPIRequest(method, path, body string) ([]byte, error) {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return nil, err
	}
	fc := client.GetFleetClient()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, fc.URL+path, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("Content-Type", "application/json")

	resp, err := fc.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, fmt.Errorf("%s %s returned %d: %s", method, path, resp.StatusCode, respBody)
	}
	return respBody, nil
}
