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
	"os"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
)

// envVarActive reports whether an environment variable is both set and non-empty.
func envVarActive(key string) bool {
	v, ok := os.LookupEnv(key)
	return ok && v != ""
}

// applyAuthOverride applies auth fields from a later-layer source (e.g. a
// provider Kibana/Fleet block) onto c, clearing conflicting prior-layer auth
// when the source specifies a primary auth method.
//
// Priority for "primary method": bearer > api_key > username. A username is
// required to claim basic auth; a password alone is not a valid intent signal
// (basic auth without a username cannot authenticate), so it does not trigger
// clearing of inherited api_key/bearer_token.
//
// Non-empty source fields overwrite c; empty source fields leave c unchanged,
// which preserves partial overrides (e.g. setting only username while
// inheriting password from a prior layer).
func applyAuthOverride(c *kibanaoapi.Config, user, pass, key, bearer string) {
	switch {
	case bearer != "":
		c.Username, c.Password, c.APIKey = "", "", ""
	case key != "":
		c.Username, c.Password, c.BearerToken = "", "", ""
	case user != "":
		c.APIKey, c.BearerToken = "", ""
	}

	if user != "" {
		c.Username = user
	}
	if pass != "" {
		c.Password = pass
	}
	if key != "" {
		c.APIKey = key
	}
	if bearer != "" {
		c.BearerToken = bearer
	}
}

// applyAuthEnvOverrides applies <prefix>_USERNAME/PASSWORD/API_KEY/BEARER_TOKEN
// environment variables onto c. The clearing rule matches applyAuthOverride
// (bearer > api_key > username), but only non-empty env vars signal intent —
// an env var explicitly set to "" still overrides the field on c (the
// established override contract), without triggering clearing.
func applyAuthEnvOverrides(c *kibanaoapi.Config, prefix string) {
	userKey := prefix + "_USERNAME"
	passKey := prefix + "_PASSWORD"
	apiKeyKey := prefix + "_API_KEY"
	bearerKey := prefix + "_BEARER_TOKEN"

	switch {
	case envVarActive(bearerKey):
		c.Username, c.Password, c.APIKey = "", "", ""
	case envVarActive(apiKeyKey):
		c.Username, c.Password, c.BearerToken = "", "", ""
	case envVarActive(userKey):
		c.APIKey, c.BearerToken = "", ""
	}

	c.Username = withEnvironmentOverride(c.Username, userKey)
	c.Password = withEnvironmentOverride(c.Password, passKey)
	c.APIKey = withEnvironmentOverride(c.APIKey, apiKeyKey)
	c.BearerToken = withEnvironmentOverride(c.BearerToken, bearerKey)
}

// authMethodCount returns the number of distinct primary auth methods set on
// c. Basic auth is only counted when Username is set, mirroring
// applyAuthOverride's intent rule and the transport's selection (which
// requires a username to issue a Basic header).
func authMethodCount(c kibanaoapi.Config) int {
	count := 0
	if c.Username != "" {
		count++
	}
	if c.APIKey != "" {
		count++
	}
	if c.BearerToken != "" {
		count++
	}
	return count
}
