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

type authMethod int

const (
	authMethodNone authMethod = iota
	authMethodBasicAuth
	authMethodAPIKey
	authMethodBearerToken
)

func clearConflictingAuth(c *kibanaoapi.Config, method authMethod) {
	switch method {
	case authMethodBasicAuth:
		c.APIKey = ""
		c.BearerToken = ""
	case authMethodAPIKey:
		c.Username = ""
		c.Password = ""
		c.BearerToken = ""
	case authMethodBearerToken:
		c.Username = ""
		c.Password = ""
		c.APIKey = ""
	}
}

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
