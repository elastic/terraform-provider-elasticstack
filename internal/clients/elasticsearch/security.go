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

package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/security/createapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/createcrossclusterapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/invalidateapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/updateapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/security/updatecrossclusterapikey"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, user *types.User, password, passwordHash *string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	req := typedClient.Security.PutUser(user.Username).
		Enabled(user.Enabled).
		Roles(user.Roles...)

	if user.Email != nil {
		req.Email(*user.Email)
	}
	if user.FullName != nil {
		req.FullName(*user.FullName)
	}
	if user.Metadata != nil {
		req.Metadata(user.Metadata)
	}

	if password != nil {
		req.Password(*password)
	}
	if passwordHash != nil {
		req.PasswordHash(*passwordHash)
	}

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError("Unable to create or update a user", err.Error())
		return diags
	}

	return diags
}

func GetUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string) (*types.User, diag.Diagnostics) {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	res, err := typedClient.Security.GetUser().Username(username).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diag.FromErr(err)
	}

	if user, ok := res[username]; ok {
		return &user, diags
	}

	return nil, diagutil.SDKErrorDiag(
		"Unable to find a user in the cluster",
		fmt.Sprintf(`Unable to find "%s" user in the cluster`, username),
	)
}

func DeleteUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Security.DeleteUser(username).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Unable to delete a user", err.Error())
		return diags
	}

	return diags
}

func EnableUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError(
			"Unable to get Elasticsearch client",
			err.Error(),
		)
		return diags
	}

	_, err = typedClient.Security.EnableUser(username).Do(ctx)
	if err != nil {
		diags.AddError(
			"Unable to enable system user",
			err.Error(),
		)
		return diags
	}

	return diags
}

func DisableUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError(
			"Unable to get Elasticsearch client",
			err.Error(),
		)
		return diags
	}

	_, err = typedClient.Security.DisableUser(username).Do(ctx)
	if err != nil {
		diags.AddError(
			"Unable to disable system user",
			err.Error(),
		)
		return diags
	}

	return diags
}

func ChangeUserPassword(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string, password, passwordHash *string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError(
			"Unable to get Elasticsearch client",
			err.Error(),
		)
		return diags
	}

	req := typedClient.Security.ChangePassword().Username(username)
	if password != nil {
		req.Password(*password)
	}
	if passwordHash != nil {
		req.PasswordHash(*passwordHash)
	}

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError(
			"Unable to change user's password",
			err.Error(),
		)
		return diags
	}

	return diags
}

func PutRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, role *types.Role) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	req := typedClient.Security.PutRole(name)

	req.Applications(role.Applications...)
	req.Cluster(role.Cluster...)
	if role.Description != nil {
		req.Description(*role.Description)
	}
	if role.Global != nil {
		globalJSON, err := json.Marshal(role.Global)
		if err != nil {
			return diagutil.SDKErrorDiag("Unable to marshal global privileges", err.Error())
		}
		var global map[string]json.RawMessage
		if err := json.Unmarshal(globalJSON, &global); err != nil {
			return diagutil.SDKErrorDiag("Unable to convert global privileges", err.Error())
		}
		req.Global(global)
	}
	req.Indices(role.Indices...)
	req.Metadata(role.Metadata)
	req.RemoteIndices(role.RemoteIndices...)
	req.RunAs(role.RunAs...)

	_, err = req.Do(ctx)
	if err != nil {
		return diagutil.SDKErrorDiag("Unable to create or update a role", err.Error())
	}

	return diags
}

func GetRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rolename string) (*types.Role, diag.Diagnostics) {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	res, err := typedClient.Security.GetRole().Name(rolename).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diag.FromErr(err)
	}

	if role, ok := res[rolename]; ok {
		return &role, diags
	}
	return nil, diagutil.SDKErrorDiag(
		"Unable to find a role in the cluster",
		fmt.Sprintf(`Unable to find "%s" role in the cluster`, rolename),
	)
}

func DeleteRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rolename string) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = typedClient.Security.DeleteRole(rolename).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		return diagutil.SDKErrorDiag("Unable to delete a role", err.Error())
	}

	return diags
}

func PutRoleMapping(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, roleMapping *types.SecurityRoleMapping) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	req := typedClient.Security.PutRoleMapping(name).
		Enabled(roleMapping.Enabled).
		Rules(&roleMapping.Rules)

	if len(roleMapping.Roles) > 0 {
		req.Roles(roleMapping.Roles...)
	}
	if len(roleMapping.RoleTemplates) > 0 {
		req.RoleTemplates(roleMapping.RoleTemplates...)
	}
	if roleMapping.Metadata != nil {
		req.Metadata(roleMapping.Metadata)
	}

	// The typed client's Script type marshals template strings as objects
	// {"source":"..."}. The ES role mapping API accepts and returns
	// templates as plain strings by default, and sending an object may
	// cause ES to wrap the value. Override the request body to preserve
	// the template as a plain string, matching the old provider behaviour.
	if len(roleMapping.RoleTemplates) > 0 {
		body := map[string]any{
			"enabled": roleMapping.Enabled,
			"rules":   &roleMapping.Rules,
		}
		if len(roleMapping.Roles) > 0 {
			body["roles"] = roleMapping.Roles
		}
		if roleMapping.Metadata != nil {
			body["metadata"] = roleMapping.Metadata
		}
		templates := make([]map[string]any, len(roleMapping.RoleTemplates))
		for i, rt := range roleMapping.RoleTemplates {
			t := map[string]any{}
			if rt.Format != nil {
				t["format"] = rt.Format.String()
			}
			if rt.Template.Source != nil {
				t["template"] = *rt.Template.Source
			}
			templates[i] = t
		}
		body["role_templates"] = templates
		data, err := json.Marshal(body)
		if err != nil {
			diags.AddError("Unable to marshal role mapping request", err.Error())
			return diags
		}
		req.Raw(bytes.NewReader(data))
	}

	_, err = req.Do(ctx)
	if err != nil {
		diags.AddError("Unable to create or update a role mapping", err.Error())
		return diags
	}

	return diags
}

func GetRoleMapping(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, roleMappingName string) (*types.SecurityRoleMapping, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return nil, diags
	}

	res, err := typedClient.Security.GetRoleMapping().Name(roleMappingName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, diags
		}
		diags.AddError("Unable to get role mapping", err.Error())
		return nil, diags
	}

	if roleMapping, ok := res[roleMappingName]; ok {
		return &roleMapping, diags
	}

	diags.AddError("Role mapping not found", fmt.Sprintf("unable to find role mapping '%s' in the cluster", roleMappingName))
	return nil, diags
}

func DeleteRoleMapping(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, roleMappingName string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Security.DeleteRoleMapping(roleMappingName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Unable to delete role mapping", err.Error())
		return diags
	}

	return diags
}

func CreateAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, req *createapikey.Request) (*createapikey.Response, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return nil, diags
	}

	res, err := typedClient.Security.CreateApiKey().Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to create apikey", err.Error())
		return nil, diags
	}

	return res, diags
}

func UpdateAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, req *updateapikey.Request) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Security.UpdateApiKey(id).Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to update apikey", err.Error())
		return diags
	}

	return diags
}

func GetAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string) (*types.ApiKey, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return nil, diags
	}

	res, err := typedClient.Security.GetApiKey().Id(id).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, diags
		}
		diags.AddError("Unable to get an apikey", err.Error())
		return nil, diags
	}

	if len(res.ApiKeys) != 1 {
		diags.AddError(
			"Unable to find an apikey in the cluster",
			fmt.Sprintf(`Unable to find "%s" apikey in the cluster`, id),
		)
		return nil, diags
	}

	apiKey := res.ApiKeys[0]
	return &apiKey, diags
}

func DeleteAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Security.InvalidateApiKey().Request(&invalidateapikey.Request{
		Ids: []string{id},
	}).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return diags
		}
		diags.AddError("Unable to delete an apikey", err.Error())
		return diags
	}

	return diags
}

func CreateCrossClusterAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, req *createcrossclusterapikey.Request) (*createcrossclusterapikey.Response, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return nil, diags
	}

	res, err := typedClient.Security.CreateCrossClusterApiKey().Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to create cross cluster apikey", err.Error())
		return nil, diags
	}

	return res, diags
}

func UpdateCrossClusterAPIKey(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, id string, req *updatecrossclusterapikey.Request) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Security.UpdateCrossClusterApiKey(id).Request(req).Do(ctx)
	if err != nil {
		diags.AddError("Unable to update cross cluster apikey", err.Error())
		return diags
	}

	return diags
}
