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
	"maps"
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
	return types.ListNull(KibanaConnectionObjectType())
}

// KibanaConnectionObjectType returns the object type for kibana_connection list
// elements. Managed and ephemeral resources use the same connection block shape.
func KibanaConnectionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: kibanaConnectionBlockObjectAttrTypes(),
	}
}

// ElasticsearchConnectionNullList returns a properly-typed null list value for the
// elasticsearch_connection block. Use when building state in ImportState so the
// framework list element type matches the resource schema.
func ElasticsearchConnectionNullList() types.List {
	return types.ListNull(ElasticsearchConnectionObjectType())
}

// ElasticsearchConnectionObjectType returns the object type for elasticsearch_connection
// list elements. Managed and ephemeral resources use the same connection block shape.
func ElasticsearchConnectionObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: elasticsearchConnectionBlockObjectAttrTypes(),
	}
}

func GetEsFWConnectionBlock() fwschema.Block {
	return fwschema.ListNestedBlock{
		MarkdownDescription: descESConnectionBlock,
		Description:         descESConnectionBlock,
		NestedObject: fwschema.NestedBlockObject{
			Attributes: buildFWESConnectionAttributes(),
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

func GetKbFWConnectionBlock() fwschema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName(attrUsername)
	passwordPath := path.MatchRelative().AtParent().AtName(attrPassword)
	apiKeyPath := path.MatchRelative().AtParent().AtName(attrAPIKey)
	bearerTokenPath := path.MatchRelative().AtParent().AtName(attrBearerToken)

	return fwschema.ListNestedBlock{
		MarkdownDescription: descKbConnectionBlock,
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				attrAPIKey: fwschema.StringAttribute{
					MarkdownDescription: descKbAPIKey,
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				attrBearerToken: fwschema.StringAttribute{
					MarkdownDescription: descKbBearerToken,
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				attrUsername: fwschema.StringAttribute{
					MarkdownDescription: descKbUsername,
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				attrPassword: fwschema.StringAttribute{
					MarkdownDescription: descKbPassword,
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				attrEndpoints: fwschema.ListAttribute{
					MarkdownDescription: descKbEndpoints,
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				attrCACerts: fwschema.ListAttribute{
					MarkdownDescription: descKbCACerts,
					Optional:            true,
					ElementType:         types.StringType,
				},
				attrInsecure: fwschema.BoolAttribute{
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

func GetFleetFWConnectionBlock() fwschema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName(attrUsername)
	passwordPath := path.MatchRelative().AtParent().AtName(attrPassword)
	apiKeyPath := path.MatchRelative().AtParent().AtName(attrAPIKey)
	bearerTokenPath := path.MatchRelative().AtParent().AtName(attrBearerToken)

	return fwschema.ListNestedBlock{
		MarkdownDescription: "Fleet connection configuration block.",
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				attrUsername: fwschema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Fleet.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				attrPassword: fwschema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Fleet.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				attrAPIKey: fwschema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Fleet.",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				attrBearerToken: fwschema.StringAttribute{
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
				attrCACerts: fwschema.ListAttribute{
					MarkdownDescription: "A list of paths to CA certificates to validate the certificate presented by the Fleet server.",
					Optional:            true,
					ElementType:         types.StringType,
				},
				attrInsecure: fwschema.BoolAttribute{
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

var (
	elasticsearchConnectionBlockObjectAttrTypesOnce sync.Once
	elasticsearchConnectionBlockObjectAttrTypesVal  map[string]attr.Type

	kibanaConnectionBlockObjectAttrTypesOnce sync.Once
	kibanaConnectionBlockObjectAttrTypesVal  map[string]attr.Type
)

func copyAttrTypes(src map[string]attr.Type) map[string]attr.Type {
	if src == nil {
		return nil
	}
	out := make(map[string]attr.Type, len(src))
	maps.Copy(out, src)
	return out
}

func connectionBlockObjectAttrTypes(block fwschema.Block) (map[string]attr.Type, error) {
	lb, ok := block.(fwschema.ListNestedBlock)
	if !ok {
		return nil, fmt.Errorf("connection block is %T, want ListNestedBlock", block)
	}
	return fwNestedBlockAttributesToAttrTypes(lb.NestedObject.Attributes)
}

func elasticsearchConnectionBlockObjectAttrTypesFallback() map[string]attr.Type {
	return map[string]attr.Type{
		attrUsername:               types.StringType,
		attrPassword:               types.StringType,
		attrAPIKey:                 types.StringType,
		attrBearerToken:            types.StringType,
		attrESClientAuthentication: types.StringType,
		attrEndpoints:              types.ListType{ElemType: types.StringType},
		attrHeaders:                types.MapType{ElemType: types.StringType},
		attrInsecure:               types.BoolType,
		attrCAFile:                 types.StringType,
		attrCAData:                 types.StringType,
		attrCAFingerprint:          types.StringType,
		attrCertFile:               types.StringType,
		attrKeyFile:                types.StringType,
		attrCertData:               types.StringType,
		attrKeyData:                types.StringType,
	}
}

func kibanaConnectionBlockObjectAttrTypesFallback() map[string]attr.Type {
	return map[string]attr.Type{
		attrAPIKey:      types.StringType,
		attrBearerToken: types.StringType,
		attrUsername:    types.StringType,
		attrPassword:    types.StringType,
		attrEndpoints:   types.ListType{ElemType: types.StringType},
		attrCACerts:     types.ListType{ElemType: types.StringType},
		attrInsecure:    types.BoolType,
	}
}

func elasticsearchConnectionBlockObjectAttrTypes() map[string]attr.Type {
	elasticsearchConnectionBlockObjectAttrTypesOnce.Do(func() {
		m, err := connectionBlockObjectAttrTypes(GetEsFWConnectionBlock())
		if err != nil {
			elasticsearchConnectionBlockObjectAttrTypesVal = elasticsearchConnectionBlockObjectAttrTypesFallback()
			return
		}
		elasticsearchConnectionBlockObjectAttrTypesVal = m
	})
	return copyAttrTypes(elasticsearchConnectionBlockObjectAttrTypesVal)
}

func kibanaConnectionBlockObjectAttrTypes() map[string]attr.Type {
	kibanaConnectionBlockObjectAttrTypesOnce.Do(func() {
		m, err := connectionBlockObjectAttrTypes(GetKbFWConnectionBlock())
		if err != nil {
			kibanaConnectionBlockObjectAttrTypesVal = kibanaConnectionBlockObjectAttrTypesFallback()
			return
		}
		kibanaConnectionBlockObjectAttrTypesVal = m
	})
	return copyAttrTypes(kibanaConnectionBlockObjectAttrTypesVal)
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
