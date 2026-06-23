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
	"fmt"
	"os"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// clearConflictingAuth removes auth fields on c that are incompatible with the
// selected primary method. The method values mirror the priority used by the
// transport layer and the auth resolution helpers: bearer > api_key > basic.
func clearConflictingAuth(c *kibanaoapi.Config, method string) {
	switch method {
	case "bearer":
		c.Username, c.Password, c.APIKey = "", "", ""
	case "api_key":
		c.Username, c.Password, c.BearerToken = "", "", ""
	case "basic":
		c.APIKey, c.BearerToken = "", ""
	}
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
		clearConflictingAuth(c, "bearer")
	case key != "":
		clearConflictingAuth(c, "api_key")
	case user != "":
		clearConflictingAuth(c, "basic")
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

	userValue, userOK := os.LookupEnv(userKey)
	passValue, passOK := os.LookupEnv(passKey)
	apiKeyValue, apiKeyOK := os.LookupEnv(apiKeyKey)
	bearerValue, bearerOK := os.LookupEnv(bearerKey)

	switch {
	case bearerOK && bearerValue != "":
		clearConflictingAuth(c, "bearer")
	case apiKeyOK && apiKeyValue != "":
		clearConflictingAuth(c, "api_key")
	case userOK && userValue != "":
		clearConflictingAuth(c, "basic")
	}

	if userOK {
		c.Username = userValue
	}
	if passOK {
		c.Password = passValue
	}
	if apiKeyOK {
		c.APIKey = apiKeyValue
	}
	if bearerOK {
		c.BearerToken = bearerValue
	}
}

// addMultipleAuthWarning appends a uniform warning when more than one primary
// auth method is present in a resolved component configuration.
func addMultipleAuthWarning(diags *fwdiags.Diagnostics, component, envDesc string) {
	diags.AddWarning(
		fmt.Sprintf("Multiple %s authentication methods configured", component),
		fmt.Sprintf("More than one of username/password (username must be set), api_key, or bearer_token is set in "+
			"the resolved %s configuration. Only one will be used. Check your "+
			"provider configuration and %s for conflicting auth settings.", component, envDesc),
	)
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
