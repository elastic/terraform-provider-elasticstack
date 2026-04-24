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

package clients

import (
	"testing"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/stretchr/testify/require"
)

// newTestAPIClientCore is shared by in-package _test.go helpers and by
// [NewTestProviderClientFactoryForResourceUnitTests] for use from other
// internal package tests.
func newTestAPIClientCore(t *testing.T) *apiClient {
	t.Helper()

	kibOapi, err := kibanaoapi.NewClient(kibanaoapi.Config{
		URL:      "http://localhost:5601",
		Username: "elastic",
		Password: "changeme",
	})
	require.NoError(t, err)

	return &apiClient{
		kibanaOapi:     kibOapi,
		version:        "unit-testing",
		kibanaEndpoint: "http://localhost:5601",
		fleetEndpoint:  "", // fleet client is nil; empty endpoint represents unconfigured Fleet
	}
}

// NewTestProviderClientFactoryForResourceUnitTests returns a [ProviderClientFactory]
// for use from unit tests in other internal packages (e.g. [resourcecore]). It
// is not for production provider code.
func NewTestProviderClientFactoryForResourceUnitTests(t *testing.T) *ProviderClientFactory {
	t.Helper()
	return NewProviderClientFactory(newTestAPIClientCore(t))
}
