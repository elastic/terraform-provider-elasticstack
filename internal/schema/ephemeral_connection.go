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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GetEsEphemeralConnectionBlock returns the elasticsearch_connection block for
// ephemeral resources, mirroring GetEsFWConnectionBlock for managed resources.
func GetEsEphemeralConnectionBlock() schema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName("username")
	passwordPath := path.MatchRelative().AtParent().AtName("password")
	apiKeyPath := path.MatchRelative().AtParent().AtName("api_key")
	bearerTokenPath := path.MatchRelative().AtParent().AtName("bearer_token")
	caFilePath := path.MatchRelative().AtParent().AtName("ca_file")
	caDataPath := path.MatchRelative().AtParent().AtName("ca_data")
	certFilePath := path.MatchRelative().AtParent().AtName("cert_file")
	certDataPath := path.MatchRelative().AtParent().AtName("cert_data")
	keyFilePath := path.MatchRelative().AtParent().AtName("key_file")
	keyDataPath := path.MatchRelative().AtParent().AtName("key_data")

	return schema.ListNestedBlock{
		MarkdownDescription: "Elasticsearch connection configuration block.",
		Description:         "Elasticsearch connection configuration block.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"username": schema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Elasticsearch.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				"password": schema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Elasticsearch.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				"api_key": schema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Elasticsearch",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				"bearer_token": schema.StringAttribute{
					MarkdownDescription: "Bearer Token to use for authentication to Elasticsearch",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				"es_client_authentication": schema.StringAttribute{
					MarkdownDescription: "ES Client Authentication field to be used with the JWT token",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
						stringvalidator.AlsoRequires(bearerTokenPath),
					},
				},
				"endpoints": schema.ListAttribute{
					MarkdownDescription: "A list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				"headers": schema.MapAttribute{
					MarkdownDescription: "A list of headers to be sent with each request to Elasticsearch.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				"insecure": schema.BoolAttribute{
					MarkdownDescription: "Disable TLS certificate validation",
					Optional:            true,
				},
				"ca_file": schema.StringAttribute{
					MarkdownDescription: "Path to a custom Certificate Authority certificate",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(caDataPath),
					},
				},
				"ca_data": schema.StringAttribute{
					MarkdownDescription: "PEM-encoded custom Certificate Authority certificate",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(caFilePath),
					},
				},
				"cert_file": schema.StringAttribute{
					MarkdownDescription: "Path to a file containing the PEM encoded certificate for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(keyFilePath),
						stringvalidator.ConflictsWith(caDataPath, keyDataPath),
					},
				},
				"key_file": schema.StringAttribute{
					MarkdownDescription: "Path to a file containing the PEM encoded private key for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(certFilePath),
						stringvalidator.ConflictsWith(certDataPath, keyDataPath),
					},
				},
				"cert_data": schema.StringAttribute{
					MarkdownDescription: "PEM encoded certificate for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(keyDataPath),
						stringvalidator.ConflictsWith(certFilePath, keyFilePath),
					},
				},
				"key_data": schema.StringAttribute{
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

// ElasticsearchConnectionNullList returns a properly-typed null list value for the
// elasticsearch_connection block on ephemeral resources.
func ElasticsearchConnectionNullList() types.List {
	return types.ListNull(ElasticsearchConnectionObjectType())
}

// ElasticsearchConnectionObjectType returns the object type for elasticsearch_connection entries.
func ElasticsearchConnectionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"api_key":                  types.StringType,
			"bearer_token":             types.StringType,
			"ca_data":                  types.StringType,
			"ca_file":                  types.StringType,
			"cert_data":                types.StringType,
			"cert_file":                types.StringType,
			"endpoints":                types.ListType{ElemType: types.StringType},
			"es_client_authentication": types.StringType,
			"headers":                  types.MapType{ElemType: types.StringType},
			"insecure":                 types.BoolType,
			"key_data":                 types.StringType,
			"key_file":                 types.StringType,
			"password":                 types.StringType,
			"username":                 types.StringType,
		},
	}
}

// GetKbEphemeralConnectionBlock returns the kibana_connection block for
// ephemeral resources, mirroring GetKbFWConnectionBlock for managed resources.
func GetKbEphemeralConnectionBlock() schema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName("username")
	passwordPath := path.MatchRelative().AtParent().AtName("password")
	apiKeyPath := path.MatchRelative().AtParent().AtName("api_key")
	bearerTokenPath := path.MatchRelative().AtParent().AtName("bearer_token")

	return schema.ListNestedBlock{
		MarkdownDescription: "Kibana connection configuration block.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"api_key": schema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Kibana",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				"bearer_token": schema.StringAttribute{
					MarkdownDescription: "Bearer Token to use for authentication to Kibana",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				"username": schema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Kibana.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				"password": schema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Kibana.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				"endpoints": schema.ListAttribute{
					MarkdownDescription: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				"ca_certs": schema.ListAttribute{
					MarkdownDescription: "A list of paths to CA certificates to validate the certificate presented by the Kibana server.",
					Optional:            true,
					ElementType:         types.StringType,
				},
				"insecure": schema.BoolAttribute{
					MarkdownDescription: "Disable TLS certificate validation",
					Optional:            true,
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// KibanaConnectionObjectType returns the object type for kibana_connection entries
// on ephemeral resources.
func KibanaConnectionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"api_key":      types.StringType,
			"bearer_token": types.StringType,
			"ca_certs":     types.ListType{ElemType: types.StringType},
			"endpoints":    types.ListType{ElemType: types.StringType},
			"insecure":     types.BoolType,
			"password":     types.StringType,
			"username":     types.StringType,
		},
	}
}
