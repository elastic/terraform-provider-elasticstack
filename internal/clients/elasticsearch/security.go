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
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutUser(ctx context.Context, apiClient *clients.APIClient, user *models.User) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	userBytes, err := json.Marshal(user)
	if err != nil {
		diags.AddError("Unable to marshal user", err.Error())
		return diags
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}
	res, err := esClient.Security.PutUser(user.Username, bytes.NewReader(userBytes), esClient.Security.PutUser.WithContext(ctx))
	if err != nil {
		diags.AddError("Unable to create or update user", err.Error())
		return diags
	}
	defer res.Body.Close()
	if fwDiags := diagutil.CheckErrorFromFW(res, "Unable to create or update a user"); fwDiags.HasError() {
		return fwDiags
	}
	return diags
}

func GetUser(ctx context.Context, apiClient *clients.APIClient, username string) (*models.User, diag.Diagnostics) {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.Security.GetUser.WithUsername(username)
	res, err := esClient.Security.GetUser(req, esClient.Security.GetUser.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := diagutil.CheckError(res, "Unable to get a user."); diags.HasError() {
		return nil, diags
	}

	// unmarshal our response to proper type
	users := make(map[string]models.User)
	if err := json.NewDecoder(res.Body).Decode(&users); err != nil {
		return nil, diag.FromErr(err)
	}

	if user, ok := users[username]; ok {
		return &user, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find a user in the cluster",
		Detail:   fmt.Sprintf(`Unable to find "%s" user in the cluster`, username),
	})
	return nil, diags
}

func DeleteUser(ctx context.Context, apiClient *clients.APIClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}
	res, err := esClient.Security.DeleteUser(username, esClient.Security.DeleteUser.WithContext(ctx))
	if err != nil {
		diags.AddError("Unable to delete user", err.Error())
		return diags
	}
	defer res.Body.Close()
	if fwDiags := diagutil.CheckErrorFromFW(res, "Unable to delete a user"); fwDiags.HasError() {
		return fwDiags
	}
	return diags
}

func EnableUser(ctx context.Context, apiClient *clients.APIClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError(
			"Unable to get Elasticsearch client",
			err.Error(),
		)
		return diags
	}
	res, err := esClient.Security.EnableUser(username, esClient.Security.EnableUser.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Unable to enable system user",
			err.Error(),
		)
		return diags
	}
	defer res.Body.Close()
	if diags := diagutil.CheckErrorFromFW(res, "Unable to enable system user"); diags.HasError() {
		return diags
	}
	return diags
}

func DisableUser(ctx context.Context, apiClient *clients.APIClient, username string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError(
			"Unable to get Elasticsearch client",
			err.Error(),
		)
		return diags
	}
	res, err := esClient.Security.DisableUser(username, esClient.Security.DisableUser.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Unable to disable system user",
			err.Error(),
		)
		return diags
	}
	defer res.Body.Close()
	if diags := diagutil.CheckErrorFromFW(res, "Unable to disable system user"); diags.HasError() {
		return diags
	}
	return diags
}

func ChangeUserPassword(ctx context.Context, apiClient *clients.APIClient, username string, userPassword *models.UserPassword) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	userPasswordBytes, err := json.Marshal(userPassword)
	if err != nil {
		diags.AddError(
			"Unable to marshal user password",
			err.Error(),
		)
		return diags
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError(
			"Unable to get Elasticsearch client",
			err.Error(),
		)
		return diags
	}
	res, err := esClient.Security.ChangePassword(
		bytes.NewReader(userPasswordBytes),
		esClient.Security.ChangePassword.WithUsername(username),
		esClient.Security.ChangePassword.WithContext(ctx),
	)
	if err != nil {
		diags.AddError(
			"Unable to change user password",
			err.Error(),
		)
		return diags
	}
	defer res.Body.Close()
	if diags := diagutil.CheckErrorFromFW(res, "Unable to change user's password"); diags.HasError() {
		return diags
	}
	return diags
}

func PutRole(ctx context.Context, apiClient *clients.APIClient, role *models.Role) diag.Diagnostics {
	var diags diag.Diagnostics

	roleBytes, err := json.Marshal(role)
	if err != nil {
		return diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Security.PutRole(role.Name, bytes.NewReader(roleBytes), esClient.Security.PutRole.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, "Unable to create role"); diags.HasError() {
		return diags
	}

	return diags
}

func GetRole(ctx context.Context, apiClient *clients.APIClient, rolename string) (*models.Role, diag.Diagnostics) {
	var diags diag.Diagnostics

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.Security.GetRole.WithName(rolename)
	res, err := esClient.Security.GetRole(req, esClient.Security.GetRole.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := diagutil.CheckError(res, "Unable to get a role."); diags.HasError() {
		return nil, diags
	}
	roles := make(map[string]models.Role)
	if err := json.NewDecoder(res.Body).Decode(&roles); err != nil {
		return nil, diag.FromErr(err)
	}

	if role, ok := roles[rolename]; ok {
		return &role, diags
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find a role in the cluster",
		Detail:   fmt.Sprintf(`Unable to find "%s" role in the cluster`, rolename),
	})
	return nil, diags
}

func DeleteRole(ctx context.Context, apiClient *clients.APIClient, rolename string) diag.Diagnostics {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Security.DeleteRole(rolename, esClient.Security.DeleteRole.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, "Unable to delete role"); diags.HasError() {
		return diags
	}

	return diags
}

func PutRoleMapping(ctx context.Context, apiClient *clients.APIClient, roleMapping *models.RoleMapping) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	roleMappingBytes, err := json.Marshal(roleMapping)
	if err != nil {
		diags.AddError("Unable to marshal role mapping", err.Error())
		return diags
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}
	res, err := esClient.Security.PutRoleMapping(roleMapping.Name, bytes.NewReader(roleMappingBytes), esClient.Security.PutRoleMapping.WithContext(ctx))
	if err != nil {
		diags.AddError("Unable to put role mapping", err.Error())
		return diags
	}
	defer res.Body.Close()
	return diagutil.CheckErrorFromFW(res, "Unable to put role mapping")
}

func GetRoleMapping(ctx context.Context, apiClient *clients.APIClient, roleMappingName string) (*models.RoleMapping, fwdiag.Diagnostics) {
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

func DeleteRoleMapping(ctx context.Context, apiClient *clients.APIClient, roleMappingName string) fwdiag.Diagnostics {
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

func CreateAPIKey(apiClient *clients.APIClient, apikey *models.APIKey) (*models.APIKeyCreateResponse, fwdiag.Diagnostics) {
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

func UpdateAPIKey(apiClient *clients.APIClient, apikey models.APIKey) fwdiag.Diagnostics {
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

func GetAPIKey(apiClient *clients.APIClient, id string) (*models.APIKeyResponse, fwdiag.Diagnostics) {
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

func DeleteAPIKey(apiClient *clients.APIClient, id string) fwdiag.Diagnostics {
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
	if err != nil && res.IsError() {
		return diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if diags := diagutil.CheckError(res, "Unable to delete an apikey"); diags.HasError() {
		return diagutil.FrameworkDiagsFromSDK(diags)
	}
	return nil
}

func CreateCrossClusterAPIKey(apiClient *clients.APIClient, apikey *models.CrossClusterAPIKey) (*models.CrossClusterAPIKeyCreateResponse, fwdiag.Diagnostics) {
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

func UpdateCrossClusterAPIKey(apiClient *clients.APIClient, apikey models.CrossClusterAPIKey) fwdiag.Diagnostics {
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
