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
// ELASTICSEARCH_USERNAME / ELASTICSEARCH_PASSWORD / ELASTICSEARCH_API_KEY / ELASTICSEARCH_BEARER_TOKEN /
// ELASTICSEARCH_ES_CLIENT_AUTHENTICATION when those environment variables are set and the corresponding
// field is non-empty in the block.
//
// When the block selects an auth mode (bearer token, API key, or username/password), environment
// variables for other auth mechanisms are not applied, so the client does not silently pick a
// different credential type than the one configured in Terraform.
//
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
		// newElasticsearchConfigFromFramework ends with withEnvironmentOverrides(), which can still
		// overwrite bearer / ES-Client-Authentication from process env even when the connection block
		// chose basic auth or API keys. Re-align typed auth with the connection block.
		scopedRestoreElasticsearchAuthFromElasticsearchConnection(esCfg, esConns[0])
		client.Elasticsearch = new(esCfg.toElasticsearchConfiguration())
	}

	return client, diags
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
	usingBearer := b.BearerToken != ""
	usingAPIKey := b.APIKey != ""
	usingBasic := b.Username != "" || b.Password != ""

	switch {
	case usingBearer:
		return b
	case usingAPIKey:
		if b.APIKey == "" {
			if v, ok := os.LookupEnv("ELASTICSEARCH_API_KEY"); ok {
				b.APIKey = v
			}
		}
		return b
	case usingBasic:
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
		return b
	default:
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
}

func scopedRestoreElasticsearchAuthFromElasticsearchConnection(esCfg *elasticsearchConfig, es ElasticsearchConnection) {
	blockBearer := es.BearerToken.ValueString()
	blockAPIKey := es.APIKey.ValueString()
	blockUser := es.Username.ValueString()
	blockPass := es.Password.ValueString()

	usingBearer := blockBearer != ""
	usingAPIKey := blockAPIKey != ""
	usingBasic := blockUser != "" || blockPass != ""

	switch {
	case usingBearer:
		esCfg.bearerToken = blockBearer
		if es.ESClientAuthentication.ValueString() != "" {
			esCfg.esClientAuthentication = es.ESClientAuthentication.ValueString()
		} else {
			esCfg.esClientAuthentication = ""
		}
	case usingAPIKey:
		esCfg.config.APIKey = blockAPIKey
		esCfg.bearerToken = ""
		esCfg.esClientAuthentication = ""
	case usingBasic:
		if blockUser != "" {
			esCfg.config.Username = blockUser
		}
		if blockPass != "" {
			esCfg.config.Password = blockPass
		}
		esCfg.bearerToken = ""
		esCfg.esClientAuthentication = ""
	default:
		// No explicit auth in the connection block: leave bearer / ES-Client-Authentication as set by
		// withEnvironmentOverrides (env-driven defaults).
	}
}
