package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func (a *ApiClient) PutElasticsearchUser(user *models.User) diag.Diagnostics {
	var diags diag.Diagnostics
	userBytes, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending request to ES: %s", userBytes)
	res, err := a.es.Security.PutUser(user.Username, bytes.NewReader(userBytes))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update a user"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) GetElasticsearchUser(username string) (*models.User, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Security.GetUser.WithUsername(username)
	res, err := a.es.Security.GetUser(req)
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

func (a *ApiClient) DeleteElasticsearchUser(username string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Security.DeleteUser(username)
	if err != nil && res.IsError() {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete a user"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchRole(role *models.Role) diag.Diagnostics {
	var diags diag.Diagnostics

	roleBytes, err := json.Marshal(role)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending request to ES: %s", roleBytes)
	res, err := a.es.Security.PutRole(role.Name, bytes.NewReader(roleBytes))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create role"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchRole(rolename string) (*models.Role, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := a.es.Security.GetRole.WithName(rolename)
	res, err := a.es.Security.GetRole(req)
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

func (a *ApiClient) DeleteElasticsearchRole(rolename string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Security.DeleteRole(rolename)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete role"); diags.HasError() {
		return diags
	}

	return diags
}

func (a ApiClient) PutElasticsearchRoleMapping(roleMapping *models.RoleMapping) diag.Diagnostics {
	var diags diag.Diagnostics
	roleMappingBytes, err := json.Marshal(roleMapping)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] creating role mapping %s: %s", roleMapping.Name, roleMappingBytes)

	res, err := a.es.Security.PutRoleMapping(roleMapping.Name, bytes.NewReader(roleMappingBytes))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags = utils.CheckError(res, fmt.Sprintf("Unable to create or update role mapping: %s", roleMapping.Name)); diags.HasError() {
		return diags
	}

	return diags
}

func (a ApiClient) GetElasticsearchRoleMapping(roleMappingName string) (*models.RoleMapping, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Security.GetRoleMapping.WithName(roleMappingName)
	res, err := a.es.Security.GetRoleMapping(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get requested role mapping: %s", roleMappingName)); diags.HasError() {
		return nil, diags
	}

	roleMappings := make(map[string]models.RoleMapping)
	if err := json.NewDecoder(res.Body).Decode(&roleMappings); err != nil {
		return nil, diag.FromErr(err)
	}

	if role, ok := roleMappings[roleMappingName]; ok {
		return &role, diags
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find the role mapping in the cluster",
		Detail:   fmt.Sprintf(`Unable to find "%s" role in the cluster`, roleMappingName),
	})
	return nil, diags
}

func (a *ApiClient) DeleteElasticsearchRoleMapping(roleMappingName string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Security.DeleteRoleMapping(roleMappingName)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete role mapping: %s", roleMappingName)); diags.HasError() {
		return diags
	}

	return diags
}
