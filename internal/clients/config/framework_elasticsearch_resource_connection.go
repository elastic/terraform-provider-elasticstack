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

package config

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// NewFromFrameworkElasticsearchResourceConnection builds a minimal [Client] whose Elasticsearch
// field is set from a Plugin Framework entity's decoded elasticsearch_connection block (exactly
// one [ElasticsearchConnection] element).
//
// Unlike [NewFromFramework], credentials from the connection block are not replaced by
// ELASTICSEARCH_USERNAME / ELASTICSEARCH_PASSWORD / ELASTICSEARCH_API_KEY / ELASTICSEARCH_BEARER_TOKEN
// when those environment variables are set and the corresponding field is non-empty in the block.
// Empty fields still pick up environment defaults so optional credentials behave like the provider.
func NewFromFrameworkElasticsearchResourceConnection(ctx context.Context, esConns []ElasticsearchConnection, version string) (Client, diag.Diagnostics) {
	if len(esConns) == 0 {
		return Client{}, nil
	}

	cfg := ProviderConfiguration{Elasticsearch: esConns}
	base := newBaseConfigForElasticsearchFrameworkConnection(cfg, version)

	client := Client{
		UserAgent: base.UserAgent,
	}

	esCfg, diags := newElasticsearchConfigFromFramework(ctx, cfg, base)
	if diags.HasError() {
		return Client{}, diags
	}

	if esCfg != nil {
		client.Elasticsearch = new(esCfg.toElasticsearchConfiguration())
	}

	return client, nil
}

func newBaseConfigForElasticsearchFrameworkConnection(config ProviderConfiguration, version string) baseConfig {
	userAgent := buildUserAgent(version)
	out := baseConfig{
		UserAgent: userAgent,
		Header:    http.Header{"User-Agent": []string{userAgent}},
	}

	if len(config.Elasticsearch) == 0 {
		return out
	}

	es := config.Elasticsearch[0]
	out.Username = es.Username.ValueString()
	out.Password = es.Password.ValueString()
	out.APIKey = es.APIKey.ValueString()
	out.BearerToken = es.BearerToken.ValueString()

	return mergeElasticsearchCredentialEnvDefaultsWhenEmpty(out)
}

func mergeElasticsearchCredentialEnvDefaultsWhenEmpty(b baseConfig) baseConfig {
	if b.Username == "" {
		if v, ok := os.LookupEnv("ELASTICSEARCH_USERNAME"); ok {
			b.Username = v
		}
	}
	if b.Password == "" {
		if v, ok := os.LookupEnv("ELASTICSEARCH_PASSWORD"); ok {
			b.Password = v
		}
	}
	if b.APIKey == "" {
		if v, ok := os.LookupEnv("ELASTICSEARCH_API_KEY"); ok {
			b.APIKey = v
		}
	}
	if b.BearerToken == "" {
		if v, ok := os.LookupEnv("ELASTICSEARCH_BEARER_TOKEN"); ok {
			b.BearerToken = v
		}
	}
	return b
}
