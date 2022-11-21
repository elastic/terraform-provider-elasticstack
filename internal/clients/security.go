package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func (a *ApiClient) PutElasticsearchUser(ctx context.Context, user *models.User) diag.Diagnostics {
	var diags diag.Diagnostics
	userBytes, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := a.es.Security.PutUser(user.Username, bytes.NewReader(userBytes), a.es.Security.PutUser.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update a user"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) GetElasticsearchUser(ctx context.Context, username string) (*models.User, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Security.GetUser.WithUsername(username)
	res, err := a.es.Security.GetUser(req, a.es.Security.GetUser.WithContext(ctx))
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

func (a *ApiClient) DeleteElasticsearchUser(ctx context.Context, username string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Security.DeleteUser(username, a.es.Security.DeleteUser.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete a user"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) EnableElasticsearchUser(ctx context.Context, username string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Security.EnableUser(username, a.es.Security.EnableUser.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to enable system user"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) DisableElasticsearchUser(ctx context.Context, username string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Security.DisableUser(username, a.es.Security.DisableUser.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to disable system user"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) ChangeElasticsearchUserPassword(ctx context.Context, username string, userPassword *models.UserPassword) diag.Diagnostics {
	var diags diag.Diagnostics
	userPasswordBytes, err := json.Marshal(userPassword)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := a.es.Security.ChangePassword(
		bytes.NewReader(userPasswordBytes),
		a.es.Security.ChangePassword.WithUsername(username),
		a.es.Security.ChangePassword.WithContext(ctx),
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

func (a *ApiClient) PutElasticsearchRole(ctx context.Context, role *models.Role) diag.Diagnostics {
	var diags diag.Diagnostics

	roleBytes, err := json.Marshal(role)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := a.es.Security.PutRole(role.Name, bytes.NewReader(roleBytes), a.es.Security.PutRole.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create role"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchRole(ctx context.Context, rolename string) (*models.Role, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := a.es.Security.GetRole.WithName(rolename)
	res, err := a.es.Security.GetRole(req, a.es.Security.GetRole.WithContext(ctx))
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

func (a *ApiClient) DeleteElasticsearchRole(ctx context.Context, rolename string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Security.DeleteRole(rolename, a.es.Security.DeleteRole.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete role"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) PutElasticsearchRoleMapping(ctx context.Context, roleMapping *models.RoleMapping) diag.Diagnostics {
	roleMappingBytes, err := json.Marshal(roleMapping)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := a.es.Security.PutRoleMapping(roleMapping.Name, bytes.NewReader(roleMappingBytes), a.es.Security.PutRoleMapping.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to put role mapping"); diags.HasError() {
		return diags
	}

	return nil
}

func (a *ApiClient) GetElasticsearchRoleMapping(ctx context.Context, roleMappingName string) (*models.RoleMapping, diag.Diagnostics) {
	req := a.es.Security.GetRoleMapping.WithName(roleMappingName)
	res, err := a.es.Security.GetRoleMapping(req, a.es.Security.GetRoleMapping.WithContext(ctx))
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

func (a *ApiClient) DeleteElasticsearchRoleMapping(ctx context.Context, roleMappingName string) diag.Diagnostics {
	res, err := a.es.Security.DeleteRoleMapping(roleMappingName, a.es.Security.DeleteRoleMapping.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete role mapping"); diags.HasError() {
		return diags
	}

	return nil
}

func (a *ApiClient) PutElasticsearchApiKey(apikey *models.ApiKey) (*models.ApiKeyResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	apikeyBytes, err := json.Marshal(apikey)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	res, err := a.es.Security.CreateAPIKey(bytes.NewReader(apikeyBytes))
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

func (a *ApiClient) GetElasticsearchApiKey(id string) (*models.ApiKeyResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Security.GetAPIKey.WithID(id)
	res, err := a.es.Security.GetAPIKey(req)
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

func (a *ApiClient) DeleteElasticsearchApiKey(id string) diag.Diagnostics {
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
	res, err := a.es.Security.InvalidateAPIKey(bytes.NewReader(apikeyBytes))
	if err != nil && res.IsError() {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete an apikey"); diags.HasError() {
		return diags
	}
	return diags
}
