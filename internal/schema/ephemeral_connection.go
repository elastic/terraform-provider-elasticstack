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

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetEsEphemeralConnectionBlock returns the elasticsearch_connection block for
// ephemeral resources, mirroring GetEsFWConnectionBlock for managed resources.
func GetEsEphemeralConnectionBlock() schema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName(attrUsername)
	passwordPath := path.MatchRelative().AtParent().AtName(attrPassword)
	apiKeyPath := path.MatchRelative().AtParent().AtName(attrAPIKey)
	bearerTokenPath := path.MatchRelative().AtParent().AtName(attrBearerToken)
	caFilePath := path.MatchRelative().AtParent().AtName(attrCAFile)
	caDataPath := path.MatchRelative().AtParent().AtName(attrCAData)
	certFilePath := path.MatchRelative().AtParent().AtName(attrCertFile)
	certDataPath := path.MatchRelative().AtParent().AtName(attrCertData)
	keyFilePath := path.MatchRelative().AtParent().AtName(attrKeyFile)
	keyDataPath := path.MatchRelative().AtParent().AtName(attrKeyData)

	return schema.ListNestedBlock{
		MarkdownDescription: descESConnectionBlock,
		Description:         descESConnectionBlock,
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attrUsername: schema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Elasticsearch.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				attrPassword: schema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Elasticsearch.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				attrAPIKey: schema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Elasticsearch",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				attrBearerToken: schema.StringAttribute{
					MarkdownDescription: "Bearer Token to use for authentication to Elasticsearch",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				attrESClientAuthentication: schema.StringAttribute{
					MarkdownDescription: "ES Client Authentication field to be used with the JWT token",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
						stringvalidator.AlsoRequires(bearerTokenPath),
					},
				},
				attrEndpoints: schema.ListAttribute{
					MarkdownDescription: "A list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				attrHeaders: schema.MapAttribute{
					MarkdownDescription: "A list of headers to be sent with each request to Elasticsearch.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				attrInsecure: schema.BoolAttribute{
					MarkdownDescription: descInsecureTLS,
					Optional:            true,
				},
				attrCAFile: schema.StringAttribute{
					MarkdownDescription: "Path to a custom Certificate Authority certificate",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(caDataPath),
					},
				},
				attrCAData: schema.StringAttribute{
					MarkdownDescription: "PEM-encoded custom Certificate Authority certificate",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(caFilePath),
					},
				},
				attrCertFile: schema.StringAttribute{
					MarkdownDescription: "Path to a file containing the PEM encoded certificate for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(keyFilePath),
						stringvalidator.ConflictsWith(caDataPath, keyDataPath),
					},
				},
				attrKeyFile: schema.StringAttribute{
					MarkdownDescription: "Path to a file containing the PEM encoded private key for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(certFilePath),
						stringvalidator.ConflictsWith(certDataPath, keyDataPath),
					},
				},
				attrCertData: schema.StringAttribute{
					MarkdownDescription: "PEM encoded certificate for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(keyDataPath),
						stringvalidator.ConflictsWith(certFilePath, keyFilePath),
					},
				},
				attrKeyData: schema.StringAttribute{
					MarkdownDescription: "PEM encoded private key for client auth",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(certDataPath),
						stringvalidator.ConflictsWith(certFilePath, keyFilePath),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// GetKbEphemeralConnectionBlock returns the kibana_connection block for
// ephemeral resources, mirroring GetKbFWConnectionBlock for managed resources.
func GetKbEphemeralConnectionBlock() schema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName(attrUsername)
	passwordPath := path.MatchRelative().AtParent().AtName(attrPassword)
	apiKeyPath := path.MatchRelative().AtParent().AtName(attrAPIKey)
	bearerTokenPath := path.MatchRelative().AtParent().AtName(attrBearerToken)

	return schema.ListNestedBlock{
		MarkdownDescription: "Kibana connection configuration block.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				attrAPIKey: schema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Kibana",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				attrBearerToken: schema.StringAttribute{
					MarkdownDescription: "Bearer Token to use for authentication to Kibana",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				attrUsername: schema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Kibana.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				attrPassword: schema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Kibana.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				attrEndpoints: schema.ListAttribute{
					MarkdownDescription: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				attrCACerts: schema.ListAttribute{
					MarkdownDescription: "A list of paths to CA certificates to validate the certificate presented by the Kibana server.",
					Optional:            true,
					ElementType:         types.StringType,
				},
				attrInsecure: schema.BoolAttribute{
					MarkdownDescription: descInsecureTLS,
					Optional:            true,
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
