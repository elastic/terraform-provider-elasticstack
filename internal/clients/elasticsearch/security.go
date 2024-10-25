package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutUser(ctx context.Context, apiClient *clients.ApiClient, user *models.User) diag.Diagnostics {
	var diags diag.Diagnostics
	userBytes, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Security.PutUser(user.Username, bytes.NewReader(userBytes), esClient.Security.PutUser.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update a user"); diags.HasError() {
		return diags
	}
	return diags
}

func GetUser(ctx context.Context, apiClient *clients.ApiClient, username string) (*models.User, diag.Diagnostics) {
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
	if diags := utils.CheckError(res, "Unable to get a user."); diags.HasError() {
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

func DeleteUser(ctx context.Context, apiClient *clients.ApiClient, username string) diag.Diagnostics {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Security.DeleteUser(username, esClient.Security.DeleteUser.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete a user"); diags.HasError() {
		return diags
	}
	return diags
}

func EnableUser(ctx context.Context, apiClient *clients.ApiClient, username string) diag.Diagnostics {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Security.EnableUser(username, esClient.Security.EnableUser.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to enable system user"); diags.HasError() {
		return diags
	}
	return diags
}

func DisableUser(ctx context.Context, apiClient *clients.ApiClient, username string) diag.Diagnostics {
	var diags diag.Diagnostics
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Security.DisableUser(username, esClient.Security.DisableUser.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to disable system user"); diags.HasError() {
		return diags
	}
	return diags
}

func ChangeUserPassword(ctx context.Context, apiClient *clients.ApiClient, username string, userPassword *models.UserPassword) diag.Diagnostics {
	var diags diag.Diagnostics
	userPasswordBytes, err := json.Marshal(userPassword)
	if err != nil {
		return diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Security.ChangePassword(
		bytes.NewReader(userPasswordBytes),
		esClient.Security.ChangePassword.WithUsername(username),
		esClient.Security.ChangePassword.WithContext(ctx),
	)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to change user's password"); diags.HasError() {
		return diags
	}
	return diags
}

func PutRole(ctx context.Context, apiClient *clients.ApiClient, role *models.Role) diag.Diagnostics {
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
	if diags := utils.CheckError(res, "Unable to create role"); diags.HasError() {
		return diags
	}

	return diags
}

func GetRole(ctx context.Context, apiClient *clients.ApiClient, rolename string) (*models.Role, diag.Diagnostics) {
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
	if diags := utils.CheckError(res, "Unable to get a role."); diags.HasError() {
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

func DeleteRole(ctx context.Context, apiClient *clients.ApiClient, rolename string) diag.Diagnostics {
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
	if diags := utils.CheckError(res, "Unable to delete role"); diags.HasError() {
		return diags
	}

	return diags
}

func PutRoleMapping(ctx context.Context, apiClient *clients.ApiClient, roleMapping *models.RoleMapping) diag.Diagnostics {
	roleMappingBytes, err := json.Marshal(roleMapping)
	if err != nil {
		return diag.FromErr(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Security.PutRoleMapping(roleMapping.Name, bytes.NewReader(roleMappingBytes), esClient.Security.PutRoleMapping.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to put role mapping"); diags.HasError() {
		return diags
	}

	return nil
}

func GetRoleMapping(ctx context.Context, apiClient *clients.ApiClient, roleMappingName string) (*models.RoleMapping, diag.Diagnostics) {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	req := esClient.Security.GetRoleMapping.WithName(roleMappingName)
	res, err := esClient.Security.GetRoleMapping(req, esClient.Security.GetRoleMapping.WithContext(ctx))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, "Unable to get a role mapping."); diags.HasError() {
		return nil, diags
	}
	roleMappings := make(map[string]models.RoleMapping)
	if err := json.NewDecoder(res.Body).Decode(&roleMappings); err != nil {
		return nil, diag.FromErr(err)

	}
	if roleMapping, ok := roleMappings[roleMappingName]; ok {
		roleMapping.Name = roleMappingName
		return &roleMapping, nil
	}

	return nil, diag.Errorf("unable to find role mapping '%s' in the cluster", roleMappingName)
}

func DeleteRoleMapping(ctx context.Context, apiClient *clients.ApiClient, roleMappingName string) diag.Diagnostics {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := esClient.Security.DeleteRoleMapping(roleMappingName, esClient.Security.DeleteRoleMapping.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete role mapping"); diags.HasError() {
		return diags
	}

	return nil
}

func CreateApiKey(apiClient *clients.ApiClient, apikey *models.ApiKey) (*models.ApiKeyCreateResponse, fwdiag.Diagnostics) {
	apikeyBytes, err := json.Marshal(apikey)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}
	res, err := esClient.Security.CreateAPIKey(bytes.NewReader(apikeyBytes))
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create apikey"); diags.HasError() {
		return nil, utils.FrameworkDiagsFromSDK(diags)
	}

	var apiKey models.ApiKeyCreateResponse

	if err := json.NewDecoder(res.Body).Decode(&apiKey); err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	return &apiKey, nil
}

func UpdateApiKey(apiClient *clients.ApiClient, apikey models.ApiKey) fwdiag.Diagnostics {
	id := apikey.ID

	apikey.Expiration = ""
	apikey.Name = ""
	apikey.ID = ""
	apikeyBytes, err := json.Marshal(apikey)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}
	res, err := esClient.Security.UpdateAPIKey(id, esClient.Security.UpdateAPIKey.WithBody(bytes.NewReader(apikeyBytes)))
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update apikey"); diags.HasError() {
		return utils.FrameworkDiagsFromSDK(diags)
	}

	return nil
}

func GetApiKey(apiClient *clients.ApiClient, id string) (*models.ApiKeyResponse, fwdiag.Diagnostics) {
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}
	req := esClient.Security.GetAPIKey.WithID(id)
	res, err := esClient.Security.GetAPIKey(req)
	if err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, "Unable to get an apikey."); diags.HasError() {
		return nil, utils.FrameworkDiagsFromSDK(diags)
	}

	// unmarshal our response to proper type
	var apiKeys struct {
		ApiKeys []models.ApiKeyResponse `json:"api_keys"`
	}
	if err := json.NewDecoder(res.Body).Decode(&apiKeys); err != nil {
		return nil, utils.FrameworkDiagFromError(err)
	}

	if len(apiKeys.ApiKeys) != 1 {
		return nil, fwdiag.Diagnostics{
			fwdiag.NewErrorDiagnostic(
				"Unable to find an apikey in the cluster",
				fmt.Sprintf(`Unable to find "%s" apikey in the cluster`, id),
			),
		}
	}

	apiKey := apiKeys.ApiKeys[0]
	return &apiKey, nil
}

func DeleteApiKey(apiClient *clients.ApiClient, id string) fwdiag.Diagnostics {
	apiKeys := struct {
		Ids []string `json:"ids"`
	}{
		[]string{id},
	}

	apikeyBytes, err := json.Marshal(apiKeys)
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}
	esClient, err := apiClient.GetESClient()
	if err != nil {
		return utils.FrameworkDiagFromError(err)
	}
	res, err := esClient.Security.InvalidateAPIKey(bytes.NewReader(apikeyBytes))
	if err != nil && res.IsError() {
		return utils.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete an apikey"); diags.HasError() {
		return utils.FrameworkDiagsFromSDK(diags)
	}
	return nil
}
