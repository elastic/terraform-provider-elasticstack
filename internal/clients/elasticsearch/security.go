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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutUser(ctx context.Context, apiClient *clients.ApiClient, user *models.User) diag.Diagnostics {
	var diags diag.Diagnostics
	userBytes, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := apiClient.GetESClient().Security.PutUser(user.Username, bytes.NewReader(userBytes), apiClient.GetESClient().Security.PutUser.WithContext(ctx))
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
	req := apiClient.GetESClient().Security.GetUser.WithUsername(username)
	res, err := apiClient.GetESClient().Security.GetUser(req, apiClient.GetESClient().Security.GetUser.WithContext(ctx))
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
	res, err := apiClient.GetESClient().Security.DeleteUser(username, apiClient.GetESClient().Security.DeleteUser.WithContext(ctx))
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
	res, err := apiClient.GetESClient().Security.EnableUser(username, apiClient.GetESClient().Security.EnableUser.WithContext(ctx))
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
	res, err := apiClient.GetESClient().Security.DisableUser(username, apiClient.GetESClient().Security.DisableUser.WithContext(ctx))
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
	res, err := apiClient.GetESClient().Security.ChangePassword(
		bytes.NewReader(userPasswordBytes),
		apiClient.GetESClient().Security.ChangePassword.WithUsername(username),
		apiClient.GetESClient().Security.ChangePassword.WithContext(ctx),
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
	res, err := apiClient.GetESClient().Security.PutRole(role.Name, bytes.NewReader(roleBytes), apiClient.GetESClient().Security.PutRole.WithContext(ctx))
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

	req := apiClient.GetESClient().Security.GetRole.WithName(rolename)
	res, err := apiClient.GetESClient().Security.GetRole(req, apiClient.GetESClient().Security.GetRole.WithContext(ctx))
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
	res, err := apiClient.GetESClient().Security.DeleteRole(rolename, apiClient.GetESClient().Security.DeleteRole.WithContext(ctx))
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
	res, err := apiClient.GetESClient().Security.PutRoleMapping(roleMapping.Name, bytes.NewReader(roleMappingBytes), apiClient.GetESClient().Security.PutRoleMapping.WithContext(ctx))
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
	req := apiClient.GetESClient().Security.GetRoleMapping.WithName(roleMappingName)
	res, err := apiClient.GetESClient().Security.GetRoleMapping(req, apiClient.GetESClient().Security.GetRoleMapping.WithContext(ctx))
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
	res, err := apiClient.GetESClient().Security.DeleteRoleMapping(roleMappingName, apiClient.GetESClient().Security.DeleteRoleMapping.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete role mapping"); diags.HasError() {
		return diags
	}

	return nil
}

func PutApiKey(apiClient *clients.ApiClient, apikey *models.ApiKey) (*models.ApiKeyResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	apikeyBytes, err := json.Marshal(apikey)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	res, err := apiClient.GetESClient().Security.CreateAPIKey(bytes.NewReader(apikeyBytes))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create apikey"); diags.HasError() {
		return nil, diags
	}

	var apiKey models.ApiKeyResponse

	if err := json.NewDecoder(res.Body).Decode(&apiKey); err != nil {
		return nil, diag.FromErr(err)
	}

	return &apiKey, diags
}

func GetApiKey(apiClient *clients.ApiClient, id string) (*models.ApiKeyResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := apiClient.GetESClient().Security.GetAPIKey.WithID(id)
	res, err := apiClient.GetESClient().Security.GetAPIKey(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		diags := append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find an apikey in the cluster.",
			Detail:   fmt.Sprintf("Unable to get apikey: '%s' from the cluster.", id),
		})
		return nil, diags
	}
	if diags := utils.CheckError(res, "Unable to get an apikey."); diags.HasError() {
		return nil, diags
	}

	// unmarshal our response to proper type
	var apiKeys struct {
		ApiKeys []models.ApiKeyResponse `json:"api_keys"`
	}
	if err := json.NewDecoder(res.Body).Decode(&apiKeys); err != nil {
		return nil, diag.FromErr(err)
	}

	if len(apiKeys.ApiKeys) != 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to find an apikey in the cluster",
			Detail:   fmt.Sprintf(`Unable to find "%s" apikey in the cluster`, id),
		})
		return nil, diags
	}

	apiKey := apiKeys.ApiKeys[0]
	return &apiKey, diags
}

func DeleteApiKey(apiClient *clients.ApiClient, id string) diag.Diagnostics {
	var diags diag.Diagnostics

	apiKeys := struct {
		Ids []string `json:"ids"`
	}{
		[]string{id},
	}

	apikeyBytes, err := json.Marshal(apiKeys)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := apiClient.GetESClient().Security.InvalidateAPIKey(bytes.NewReader(apikeyBytes))
	if err != nil && res.IsError() {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete an apikey"); diags.HasError() {
		return diags
	}
	return diags
}
