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

package checks

import (
	"encoding/base64"
	"fmt"
	"io"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func CheckUserCanAuthenticate(username string, password string) func(*terraform.State) error {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}

		credentials := fmt.Sprintf("%s:%s", username, password)
		authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(credentials)))

		req := esClient.Security.Authenticate.WithHeader(map[string]string{"Authorization": authHeader})
		resp, err := esClient.Security.Authenticate(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.IsError() {
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return fmt.Errorf("failed to authenticate as test user [%s]: failed reading response body: %w", username, readErr)
			}

			return fmt.Errorf("failed to authenticate as test user [%s]: %s", username, body)
		}
		return nil
	}
}
