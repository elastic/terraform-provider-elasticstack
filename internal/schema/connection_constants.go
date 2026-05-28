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

package schema

// Connection block attribute keys and descriptions shared by managed and
// ephemeral Elasticsearch/Kibana/Fleet connection schemas.
const (
	attrUsername               = "username"
	attrPassword               = "password"
	attrAPIKey                 = "api_key"
	attrBearerToken            = "bearer_token"
	attrESClientAuthentication = "es_client_authentication"
	attrEndpoints              = "endpoints"
	attrHeaders                = "headers"
	attrInsecure               = "insecure"
	attrCAFile                 = "ca_file"
	attrCAData                 = "ca_data"
	attrCertFile               = "cert_file"
	attrKeyFile                = "key_file"
	attrCertData               = "cert_data"
	attrKeyData                = "key_data"
	attrCACerts                = "ca_certs"

	descESConnectionBlock = "Elasticsearch connection configuration block."
	descInsecureTLS       = "Disable TLS certificate validation"

	descUsername               = "Username to use for API authentication to Elasticsearch."
	descPassword               = "Password to use for API authentication to Elasticsearch."
	descAPIKey                 = "API Key to use for authentication to Elasticsearch"
	descBearerToken            = "Bearer Token to use for authentication to Elasticsearch"
	descESClientAuthentication = "ES Client Authentication field to be used with the JWT token"
	descEndpoints              = "A list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number."
	descHeaders                = "A list of headers to be sent with each request to Elasticsearch."
	descCAFile                 = "Path to a custom Certificate Authority certificate"
	descCAData                 = "PEM-encoded custom Certificate Authority certificate"
	descCertFile               = "Path to a file containing the PEM encoded certificate for client auth"
	descKeyFile                = "Path to a file containing the PEM encoded private key for client auth"
	descCertData               = "PEM encoded certificate for client auth"
	descKeyData                = "PEM encoded private key for client auth"

	descKbConnectionBlock = "Kibana connection configuration block."
	descKbAPIKey          = "API Key to use for authentication to Kibana"
	descKbBearerToken     = "Bearer Token to use for authentication to Kibana"
	descKbUsername        = "Username to use for API authentication to Kibana."
	descKbPassword        = "Password to use for API authentication to Kibana."
	descKbEndpoints       = "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number."
	descKbCACerts         = "A list of paths to CA certificates to validate the certificate presented by the Kibana server."
)
