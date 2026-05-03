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
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, user *types.User, password, passwordHash *string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
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

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	res, err := typedClient.Security.GetUser().Username(username).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, diag.FromErr(err)
	}

	if user, ok := res[username]; ok {
		return &user, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find a user in the cluster",
		Detail:   fmt.Sprintf(`Unable to find "%s" user in the cluster`, username),
	})
	return nil, diags
}

func DeleteUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	_, err = typedClient.Security.DeleteUser(username).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return diags
		}
		diags.AddError("Unable to delete a user", err.Error())
		return diags
	}

	return diags
}

func EnableUser(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
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

	typedClient, err := apiClient.GetESTypedClient()
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

func ChangeUserPassword(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, username string, userPassword *models.UserPassword) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		diags.AddError(
			"Unable to get Elasticsearch client",
			err.Error(),
		)
		return diags
	}

	req := typedClient.Security.ChangePassword().Username(username)
	if userPassword.Password != nil {
		req.Password(*userPassword.Password)
	}
	if userPassword.PasswordHash != nil {
		req.PasswordHash(*userPassword.PasswordHash)
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

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diag.FromErr(err)
	}

	req := typedClient.Security.PutRole(name)

	if len(role.Applications) > 0 {
		req.Applications(role.Applications...)
	}
	if len(role.Cluster) > 0 {
		req.Cluster(role.Cluster...)
	}
	if role.Description != nil {
		req.Description(*role.Description)
	}
	if len(role.Global) > 0 {
		globalJSON, err := json.Marshal(role.Global)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to marshal global privileges",
				Detail:   err.Error(),
			})
			return diags
		}
		var global map[string]json.RawMessage
		if err := json.Unmarshal(globalJSON, &global); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to convert global privileges",
				Detail:   err.Error(),
			})
			return diags
		}
		req.Global(global)
	}
	if len(role.Indices) > 0 {
		req.Indices(role.Indices...)
	}
	if role.Metadata != nil && len(role.Metadata) > 0 {
		req.Metadata(role.Metadata)
	}
	if len(role.RemoteIndices) > 0 {
		req.RemoteIndices(role.RemoteIndices...)
	}
	if len(role.RunAs) > 0 {
		req.RunAs(role.RunAs...)
	}

	_, err = req.Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create or update a role",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func GetRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rolename string) (*types.Role, diag.Diagnostics) {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	res, err := typedClient.Security.GetRole().Name(rolename).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return nil, nil
		}
		return nil, diag.FromErr(err)
	}

	if role, ok := res[rolename]; ok {
		return &role, diags
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find a role in the cluster",
		Detail:   fmt.Sprintf(`Unable to find "%s" role in the cluster`, rolename),
	})
	return nil, diags
}

func DeleteRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rolename string) diag.Diagnostics {
	var diags diag.Diagnostics

	typedClient, err := apiClient.GetESTypedClient()
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = typedClient.Security.DeleteRole(rolename).Do(ctx)
	if err != nil {
		var esErr *types.ElasticsearchError
		if errors.As(err, &esErr) && esErr.Status == 404 {
			return diags
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete a role",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func PutRoleMapping(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, roleMapping *models.RoleMapping) fwdiag.Diagnostics {
	return doFWWrite(apiClient, roleMapping,
		"Unable to marshal role mapping",
		"Unable to put role mapping",
		"Unable to put role mapping",
		func(esClient *elasticsearch.Client, body io.Reader) (*esapi.Response, error) {
			return esClient.Security.PutRoleMapping(roleMapping.Name, body, esClient.Security.PutRoleMapping.WithContext(ctx))
		},
	)
}

func GetRoleMapping(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, roleMappingName string) (*models.RoleMapping, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return nil, diags
	}
	req := esClient.Security.GetRoleMapping.WithName(roleMappingName)
	res, err := esClient.Security.GetRoleMapping(req, esClient.Security.GetRoleMapping.WithContext(ctx))
	if err != nil {
		diags.AddError("Unable to get role mapping", err.Error())
		return nil, diags
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, diags
	}
	if diags := diagutil.CheckErrorFromFW(res, "Unable to get a role mapping."); diags.HasError() {
		return nil, diags
	}
	roleMappings := make(map[string]models.RoleMapping)
	if err := json.NewDecoder(res.Body).Decode(&roleMappings); err != nil {
		diags.AddError("Unable to decode role mapping response", err.Error())
		return nil, diags

	}
	if roleMapping, ok := roleMappings[roleMappingName]; ok {
		roleMapping.Name = roleMappingName
		return &roleMapping, diags
	}

	diags.AddError("Role mapping not found", fmt.Sprintf("unable to find role mapping '%s' in the cluster", roleMappingName))
	return nil, diags
}

func DeleteRoleMapping(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, roleMappingName string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}
	res, err := esClient.Security.DeleteRoleMapping(roleMappingName, esClient.Security.DeleteRoleMapping.WithContext(ctx))
	if err != nil {
		diags.AddError("Unable to delete role mapping", err.Error())
		return diags
	}
	defer res.Body.Close()
	return diagutil.CheckErrorFromFW(res, "Unable to delete role mapping")
}

func CreateAPIKey(apiClient *clients.ElasticsearchScopedClient, apikey *models.APIKey) (*models.APIKeyCreateResponse, fwdiag.Diagnostics) {
	apikeyBytes, err := json.Marshal(apikey)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := esClient.Security.CreateAPIKey(bytes.NewReader(apikeyBytes))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, "Unable to create apikey"); diags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(diags)
	}

	var apiKey models.APIKeyCreateResponse

	if err := json.NewDecoder(res.Body).Decode(&apiKey); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return &apiKey, nil
}

func UpdateAPIKey(apiClient *clients.ElasticsearchScopedClient, apikey models.APIKey) fwdiag.Diagnostics {
	id := apikey.ID

	apikey.Expiration = ""
	apikey.Name = ""
	apikey.ID = ""
	apikeyBytes, err := json.Marshal(apikey)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	res, err := esClient.Security.UpdateAPIKey(id, esClient.Security.UpdateAPIKey.WithBody(bytes.NewReader(apikeyBytes)))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, "Unable to update apikey"); diags.HasError() {
		return diagutil.FrameworkDiagsFromSDK(diags)
	}

	return nil
}

func GetAPIKey(apiClient *clients.ElasticsearchScopedClient, id string) (*models.APIKeyResponse, fwdiag.Diagnostics) {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	req := esClient.Security.GetAPIKey.WithID(id)
	res, err := esClient.Security.GetAPIKey(req)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := diagutil.CheckError(res, "Unable to get an apikey."); diags.HasError() {
		return nil, diagutil.FrameworkDiagsFromSDK(diags)
	}

	// unmarshal our response to proper type
	var apiKeys struct {
		APIKeys []models.APIKeyResponse `json:"api_keys"`
	}
	if err := json.NewDecoder(res.Body).Decode(&apiKeys); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	if len(apiKeys.APIKeys) != 1 {
		return nil, fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(
				"Unable to find an apikey in the cluster",
				fmt.Sprintf(`Unable to find "%s" apikey in the cluster`, id),
			),
		}
	}

	apiKey := apiKeys.APIKeys[0]
	return &apiKey, nil
}

func DeleteAPIKey(apiClient *clients.ElasticsearchScopedClient, id string) fwdiag.Diagnostics {
	apiKeys := struct {
		IDs []string `json:"ids"`
	}{
		[]string{id},
	}

	apikeyBytes, err := json.Marshal(apiKeys)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	res, err := esClient.Security.InvalidateAPIKey(bytes.NewReader(apikeyBytes))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, "Unable to delete an apikey"); diags.HasError() {
		return diagutil.FrameworkDiagsFromSDK(diags)
	}
	return nil
}

func CreateCrossClusterAPIKey(apiClient *clients.ElasticsearchScopedClient, apikey *models.CrossClusterAPIKey) (*models.CrossClusterAPIKeyCreateResponse, fwdiag.Diagnostics) {
	apikeyBytes, err := json.Marshal(apikey)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := esClient.Security.CreateCrossClusterAPIKey(bytes.NewReader(apikeyBytes))
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckErrorFromFW(res, "Unable to create cross cluster apikey"); diags.HasError() {
		return nil, diags
	}

	var apiKey models.CrossClusterAPIKeyCreateResponse

	if err := json.NewDecoder(res.Body).Decode(&apiKey); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	return &apiKey, nil
}

func UpdateCrossClusterAPIKey(apiClient *clients.ElasticsearchScopedClient, apikey models.CrossClusterAPIKey) fwdiag.Diagnostics {
	id := apikey.ID

	apikey.Expiration = ""
	apikey.Name = ""
	apikey.ID = ""
	apikeyBytes, err := json.Marshal(apikey)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	res, err := esClient.Security.UpdateCrossClusterAPIKey(id, bytes.NewReader(apikeyBytes))
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	return diagutil.CheckErrorFromFW(res, "Unable to update cross cluster apikey")
}
