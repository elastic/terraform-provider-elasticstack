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
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// KibanaConnectionNullList returns a properly-typed null list value for the
// kibana_connection block. Use this when building a state struct from scratch
// (e.g., in ImportState or state upgraders) so the framework can match the
// list element type against the schema instead of encountering a zero-value.
func KibanaConnectionNullList() types.List {
	return types.ListNull(types.ObjectType{
		AttrTypes: kibanaConnectionBlockObjectAttrTypes(),
	})
}

// ElasticsearchConnectionNullList returns a properly-typed null list value for the
// elasticsearch_connection block. Use when building state in ImportState so the
// framework list element type matches the resource schema.
func ElasticsearchConnectionNullList() types.List {
	return types.ListNull(types.ObjectType{
		AttrTypes: elasticsearchConnectionBlockObjectAttrTypes(),
	})
}

func GetEsFWConnectionBlock() fwschema.Block {
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

	return fwschema.ListNestedBlock{
		MarkdownDescription: "Elasticsearch connection configuration block.",
		Description:         "Elasticsearch connection configuration block.",
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"username": fwschema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Elasticsearch.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				"password": fwschema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Elasticsearch.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				"api_key": fwschema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Elasticsearch",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				"bearer_token": fwschema.StringAttribute{
					MarkdownDescription: "Bearer Token to use for authentication to Elasticsearch",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				"es_client_authentication": fwschema.StringAttribute{
					MarkdownDescription: "ES Client Authentication field to be used with the JWT token",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
						stringvalidator.AlsoRequires(bearerTokenPath),
					},
				},
				"endpoints": fwschema.ListAttribute{
					MarkdownDescription: "A list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				"headers": fwschema.MapAttribute{
					MarkdownDescription: "A list of headers to be sent with each request to Elasticsearch.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				"insecure": fwschema.BoolAttribute{
					MarkdownDescription: "Disable TLS certificate validation",
					Optional:            true,
				},
				"ca_file": fwschema.StringAttribute{
					MarkdownDescription: "Path to a custom Certificate Authority certificate",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(caDataPath),
					},
				},
				"ca_data": fwschema.StringAttribute{
					MarkdownDescription: "PEM-encoded custom Certificate Authority certificate",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(caFilePath),
					},
				},
				"cert_file": fwschema.StringAttribute{
					MarkdownDescription: "Path to a file containing the PEM encoded certificate for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(keyFilePath),
						stringvalidator.ConflictsWith(caDataPath, keyDataPath),
					},
				},
				"key_file": fwschema.StringAttribute{
					MarkdownDescription: "Path to a file containing the PEM encoded private key for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(certFilePath),
						stringvalidator.ConflictsWith(certDataPath, keyDataPath),
					},
				},
				"cert_data": fwschema.StringAttribute{
					MarkdownDescription: "PEM encoded certificate for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(keyDataPath),
						stringvalidator.ConflictsWith(certFilePath, keyFilePath),
					},
				},
				"key_data": fwschema.StringAttribute{
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

func GetKbFWConnectionBlock() fwschema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName("username")
	passwordPath := path.MatchRelative().AtParent().AtName("password")
	apiKeyPath := path.MatchRelative().AtParent().AtName("api_key")
	bearerTokenPath := path.MatchRelative().AtParent().AtName("bearer_token")

	return fwschema.ListNestedBlock{
		MarkdownDescription: "Kibana connection configuration block.",
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"api_key": fwschema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Kibana",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				"bearer_token": fwschema.StringAttribute{
					MarkdownDescription: "Bearer Token to use for authentication to Kibana",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				"username": fwschema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Kibana.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				"password": fwschema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Kibana.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				"endpoints": fwschema.ListAttribute{
					MarkdownDescription: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				"ca_certs": fwschema.ListAttribute{
					MarkdownDescription: "A list of paths to CA certificates to validate the certificate presented by the Kibana server.",
					Optional:            true,
					ElementType:         types.StringType,
				},
				"insecure": fwschema.BoolAttribute{
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

func GetFleetFWConnectionBlock() fwschema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName("username")
	passwordPath := path.MatchRelative().AtParent().AtName("password")
	apiKeyPath := path.MatchRelative().AtParent().AtName("api_key")
	bearerTokenPath := path.MatchRelative().AtParent().AtName("bearer_token")

	return fwschema.ListNestedBlock{
		MarkdownDescription: "Fleet connection configuration block.",
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"username": fwschema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Fleet.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				"password": fwschema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Fleet.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				"api_key": fwschema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Fleet.",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				"bearer_token": fwschema.StringAttribute{
					MarkdownDescription: "Bearer Token to use for authentication to Fleet.",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				"endpoint": fwschema.StringAttribute{
					MarkdownDescription: "The Fleet server where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
				},
				"ca_certs": fwschema.ListAttribute{
					MarkdownDescription: "A list of paths to CA certificates to validate the certificate presented by the Fleet server.",
					Optional:            true,
					ElementType:         types.StringType,
				},
				"insecure": fwschema.BoolAttribute{
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

var (
	elasticsearchConnectionBlockObjectAttrTypesOnce sync.Once
	elasticsearchConnectionBlockObjectAttrTypesVal  map[string]attr.Type

	kibanaConnectionBlockObjectAttrTypesOnce sync.Once
	kibanaConnectionBlockObjectAttrTypesVal  map[string]attr.Type
)

func elasticsearchConnectionBlockObjectAttrTypes() map[string]attr.Type {
	elasticsearchConnectionBlockObjectAttrTypesOnce.Do(func() {
		b := GetEsFWConnectionBlock()
		lb, ok := b.(fwschema.ListNestedBlock)
		if !ok {
			panic("internal/schema: GetEsFWConnectionBlock must be a ListNestedBlock")
		}
		m, err := fwNestedBlockAttributesToAttrTypes(lb.NestedObject.Attributes)
		if err != nil {
			panic(fmt.Sprintf("internal/schema: elasticsearch_connection attr types: %v", err))
		}
		elasticsearchConnectionBlockObjectAttrTypesVal = m
	})
	return elasticsearchConnectionBlockObjectAttrTypesVal
}

func kibanaConnectionBlockObjectAttrTypes() map[string]attr.Type {
	kibanaConnectionBlockObjectAttrTypesOnce.Do(func() {
		b := GetKbFWConnectionBlock()
		lb, ok := b.(fwschema.ListNestedBlock)
		if !ok {
			panic("internal/schema: GetKbFWConnectionBlock must be a ListNestedBlock")
		}
		m, err := fwNestedBlockAttributesToAttrTypes(lb.NestedObject.Attributes)
		if err != nil {
			panic(fmt.Sprintf("internal/schema: kibana_connection attr types: %v", err))
		}
		kibanaConnectionBlockObjectAttrTypesVal = m
	})
	return kibanaConnectionBlockObjectAttrTypesVal
}

func fwNestedBlockAttributesToAttrTypes(attrs map[string]fwschema.Attribute) (map[string]attr.Type, error) {
	out := make(map[string]attr.Type, len(attrs))
	for name, a := range attrs {
		t, err := fwAttributeToAttrType(name, a)
		if err != nil {
			return nil, err
		}
		out[name] = t
	}
	return out, nil
}

func fwAttributeToAttrType(name string, a fwschema.Attribute) (attr.Type, error) {
	switch a := a.(type) {
	case fwschema.StringAttribute:
		if a.CustomType != nil {
			return a.CustomType, nil
		}
		return types.StringType, nil
	case fwschema.BoolAttribute:
		if a.CustomType != nil {
			return a.CustomType, nil
		}
		return types.BoolType, nil
	case fwschema.ListAttribute:
		if a.CustomType != nil {
			return a.CustomType, nil
		}
		if a.ElementType == nil {
			return nil, fmt.Errorf("attribute %q: ListAttribute missing ElementType", name)
		}
		return types.ListType{ElemType: a.ElementType}, nil
	case fwschema.MapAttribute:
		if a.CustomType != nil {
			return a.CustomType, nil
		}
		if a.ElementType == nil {
			return nil, fmt.Errorf("attribute %q: MapAttribute missing ElementType", name)
		}
		return types.MapType{ElemType: a.ElementType}, nil
	default:
		return nil, fmt.Errorf("attribute %q: unsupported framework attribute type %T (extend fwAttributeToAttrType)", name, a)
	}
}
