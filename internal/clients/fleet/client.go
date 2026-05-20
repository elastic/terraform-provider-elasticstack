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
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
)

// Config is the configuration for the fleet client.
// It is identical in structure to kibanaoapi.Config and is kept as a type alias
// so that both packages share a single implementation.
type Config = kibanaoapi.Config

// Client provides an API client for Elastic Fleet.
// It is identical in structure to kibanaoapi.Client and is kept as a type alias
// so that both packages share a single implementation.
type Client = kibanaoapi.Client

// NewClient creates a new Elastic Fleet API client.
func NewClient(cfg Config) (*Client, error) {
	return kibanaoapi.NewClientWithLabel(cfg, "Fleet")
}
