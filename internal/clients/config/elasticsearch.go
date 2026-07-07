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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/terraform-provider-elasticstack/internal/debugutils"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type elasticsearchConfig struct {
	config                 elasticsearch.Config
	bearerToken            string
	esClientAuthentication string
}

func newElasticsearchConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, base baseConfig) (*elasticsearchConfig, fwdiags.Diagnostics) {
	if len(cfg.Elasticsearch) == 0 {
		return nil, nil
	}

	config := base.toElasticsearchConfig()
	esConfig := cfg.Elasticsearch[0]

	var endpoints []string
	diags := esConfig.Endpoints.ElementsAs(ctx, &endpoints, true)
	if diags.HasError() {
		return nil, diags
	}

	if len(endpoints) > 0 {
		config.config.Addresses = endpoints
	}

	for header, value := range esConfig.Headers.Elements() {
		strValue := value.(basetypes.StringValue)
		// trim the strings to remove any leading/trailing whitespace
		config.config.Header.Add(strings.TrimSpace(header), strings.TrimSpace(strValue.ValueString()))
	}

	if esConfig.BearerToken.ValueString() != "" {
		config.bearerToken = esConfig.BearerToken.ValueString()
		if esConfig.ESClientAuthentication.ValueString() != "" {
			config.esClientAuthentication = esConfig.ESClientAuthentication.ValueString()
		}
	}

	if esConfig.Insecure.ValueBool() {
		tlsClientConfig := config.ensureTLSClientConfig()
		tlsClientConfig.InsecureSkipVerify = true
	}

	if caFile := esConfig.CAFile.ValueString(); caFile != "" {
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			diags.Append(fwdiags.NewErrorDiagnostic("Unable to read CA file", err.Error()))
			return nil, diags
		}
		config.config.CACert = caCert
	}
	if caData := esConfig.CAData.ValueString(); caData != "" {
		config.config.CACert = []byte(caData)
	}
	if fingerprint := esConfig.CAFingerprint.ValueString(); fingerprint != "" {
		config.config.CertificateFingerprint = fingerprint
	}

	if certFile := esConfig.CertFile.ValueString(); certFile != "" {
		if keyFile := esConfig.KeyFile.ValueString(); keyFile != "" {
			cert, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				diags.Append(fwdiags.NewErrorDiagnostic("Unable to read certificate or key file", err.Error()))
				return nil, diags
			}
			tlsClientConfig := config.ensureTLSClientConfig()
			tlsClientConfig.Certificates = []tls.Certificate{cert}
		} else {
			diags.Append(fwdiags.NewErrorDiagnostic("Unable to read key file", "Path to key file has not been configured or is empty"))
			return nil, diags
		}
	}
	if certData := esConfig.CertData.ValueString(); certData != "" {
		if keyData := esConfig.KeyData.ValueString(); keyData != "" {
			cert, err := tls.X509KeyPair([]byte(certData), []byte(keyData))
			if err != nil {
				diags.Append(fwdiags.NewErrorDiagnostic("Unable to parse certificate or key", err.Error()))
				return nil, diags
			}
			tlsClientConfig := config.ensureTLSClientConfig()
			tlsClientConfig.Certificates = []tls.Certificate{cert}
		} else {
			diags.Append(fwdiags.NewErrorDiagnostic("Unable to parse key", "Key data has not been configured or is empty"))
			return nil, diags
		}
	}

	if debugutils.IsDebugOrHigher() {
		config.config.EnableDebugLogger = true
		config.config.Logger = &debugLogger{Name: "elasticsearch"}
	}

	config = config.withEnvironmentOverrides()
	return &config, nil
}

func (c *elasticsearchConfig) ensureTLSClientConfig() *tls.Config {
	if c.config.Transport == nil {
		c.config.Transport = http.DefaultTransport.(*http.Transport)
	}
	if c.config.Transport.(*http.Transport).TLSClientConfig == nil {
		c.config.Transport.(*http.Transport).TLSClientConfig = &tls.Config{}
	}
	return c.config.Transport.(*http.Transport).TLSClientConfig
}

func (c elasticsearchConfig) withEnvironmentOverrides() elasticsearchConfig {
	if endpointsCSV, ok := os.LookupEnv("ELASTICSEARCH_ENDPOINTS"); ok {
		endpoints := make([]string, 0)
		for e := range strings.SplitSeq(endpointsCSV, ",") {
			endpoints = append(endpoints, strings.TrimSpace(e))
		}
		c.config.Addresses = endpoints
	}

	if insecure, ok := os.LookupEnv("ELASTICSEARCH_INSECURE"); ok {
		if insecureValue, err := strconv.ParseBool(insecure); err == nil {
			tlsClientConfig := c.ensureTLSClientConfig()
			tlsClientConfig.InsecureSkipVerify = insecureValue
		}
	}

	if bearerToken := os.Getenv("ELASTICSEARCH_BEARER_TOKEN"); bearerToken != "" {
		c.bearerToken = bearerToken
	}

	if esClientAuthentication := os.Getenv("ELASTICSEARCH_ES_CLIENT_AUTHENTICATION"); esClientAuthentication != "" {
		c.esClientAuthentication = esClientAuthentication
	}

	if caFingerprint := os.Getenv("ELASTICSEARCH_CA_FINGERPRINT"); caFingerprint != "" {
		c.config.CertificateFingerprint = caFingerprint
		c.config.CACert = nil
	}

	return c
}

func (c elasticsearchConfig) toElasticsearchConfiguration() elasticsearch.Config {
	if c.bearerToken != "" {
		c.config.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
	}

	if c.esClientAuthentication != "" {
		c.config.Header.Set("ES-Client-Authentication", fmt.Sprintf("SharedSecret %s", c.esClientAuthentication))
	}

	return c.config
}
