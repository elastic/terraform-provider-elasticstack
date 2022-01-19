package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type CompositeId struct {
	ClusterId  string
	ResourceId string
}

func CompositeIdFromStr(id string) (*CompositeId, diag.Diagnostics) {
	var diags diag.Diagnostics
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong resource ID.",
			Detail:   "Resource ID must have following format: <cluster_uuid>/<resource identifier>",
		})
		return nil, diags
	}
	return &CompositeId{
			ClusterId:  idParts[0],
			ResourceId: idParts[1],
		},
		diags
}

func (c *CompositeId) String() string {
	return fmt.Sprintf("%s/%s", c.ClusterId, c.ResourceId)
}

type ApiClient struct {
	es      *elasticsearch.Client
	version string
}

func NewApiClientFunc(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics
		config := elasticsearch.Config{}
		config.Header = http.Header{"User-Agent": []string{fmt.Sprintf("elasticstack-terraform-provider/%s", version)}}

		if v, ok := d.GetOk("elasticsearch"); ok {
			// if defined we must have only one entry
			if esc := v.([]interface{})[0]; esc != nil {
				esConfig := esc.(map[string]interface{})
				if username, ok := esConfig["username"]; ok {
					config.Username = username.(string)
				}
				if password, ok := esConfig["password"]; ok {
					config.Password = password.(string)
				}

				// default endpoints taken from Env if set
				if es := os.Getenv("ELASTICSEARCH_ENDPOINTS"); es != "" {
					endpoints := make([]string, 0)
					for _, e := range strings.Split(es, ",") {
						endpoints = append(endpoints, strings.TrimSpace(e))
					}
					config.Addresses = endpoints
				}
				// setting endpoints from config block if provided
				if eps, ok := esConfig["endpoints"]; ok && len(eps.([]interface{})) > 0 {
					endpoints := make([]string, 0)
					for _, e := range eps.([]interface{}) {
						endpoints = append(endpoints, e.(string))
					}
					config.Addresses = endpoints
				}

				if insecure, ok := esConfig["insecure"]; ok && insecure.(bool) {
					tr := http.DefaultTransport.(*http.Transport)
					tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
					config.Transport = tr
				}
			}
		}

		es, err := elasticsearch.NewClient(config)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Elasticsearch client",
				Detail:   err.Error(),
			})
		}

		return &ApiClient{es, version}, diags
	}
}

func NewApiClient(d *schema.ResourceData, meta interface{}) (*ApiClient, error) {
	defaultClient := meta.(*ApiClient)
	// if the config provided let's use it
	if esConn, ok := d.GetOk("elasticsearch_connection"); ok {
		config := elasticsearch.Config{}
		config.Header = http.Header{"User-Agent": []string{fmt.Sprintf("elasticstack-terraform-provider/%s", defaultClient.version)}}

		// there is always only 1 connection per resource
		conn := esConn.([]interface{})[0].(map[string]interface{})

		if u := conn["username"]; u != nil {
			config.Username = u.(string)
		}
		if p := conn["password"]; p != nil {
			config.Password = p.(string)
		}
		if endpoints := conn["endpoints"]; endpoints != nil {
			var addrs []string
			for _, e := range endpoints.([]interface{}) {
				addrs = append(addrs, e.(string))
			}
			config.Addresses = addrs
		}
		if insecure := conn["insecure"]; insecure.(bool) {
			tr := http.DefaultTransport.(*http.Transport)
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			config.Transport = tr
		}

		es, err := elasticsearch.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("Unable to create Elasticsearch client")
		}
		return &ApiClient{es, defaultClient.version}, nil
	} else { // or return the default client
		return defaultClient, nil
	}
}

func (a *ApiClient) GetESClient() *elasticsearch.Client {
	return a.es
}

func (a *ApiClient) ID(resourceId string) (*CompositeId, diag.Diagnostics) {
	var diags diag.Diagnostics
	clusterId, diags := a.ClusterID()
	if diags.HasError() {
		return nil, diags
	}
	log.Printf("[TRACE] cluster UUID: %s", *clusterId)
	return &CompositeId{*clusterId, resourceId}, diags
}

func (a *ApiClient) ClusterID() (*string, diag.Diagnostics) {
	var diags diag.Diagnostics
	res, err := a.es.Info()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to connect to the Elasticsearch cluster"); diags.HasError() {
		return nil, diags
	}

	info := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, diag.FromErr(err)
	}
	if uuid := info["cluster_uuid"].(string); uuid != "" && uuid != "_na_" {
		log.Printf("[TRACE] cluster UUID: %s", uuid)
		return &uuid, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to get cluster UUID",
		Detail: `Unable to get cluster UUID.
		There might be a problem with permissions or cluster is still starting up and UUID has not been populated yet.`,
	})
	return nil, diags
}

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

func (a *ApiClient) PutElasticsearchIlm(policy *models.Policy) diag.Diagnostics {
	var diags diag.Diagnostics
	policyBytes, err := json.Marshal(map[string]interface{}{"policy": policy})
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending new ILM policy to ES API: %s", policyBytes)
	req := a.es.ILM.PutLifecycle.WithBody(bytes.NewReader(policyBytes))
	res, err := a.es.ILM.PutLifecycle(policy.Name, req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the ILM policy"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) GetElasticsearchIlm(policyName string) (*models.PolicyDefinition, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.ILM.GetLifecycle.WithPolicy(policyName)
	res, err := a.es.ILM.GetLifecycle(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	if diags := utils.CheckError(res, "Unable to fetch ILM policy from the cluster."); diags.HasError() {
		return nil, diags
	}
	defer res.Body.Close()

	// our API response
	ilm := make(map[string]models.PolicyDefinition)
	if err := json.NewDecoder(res.Body).Decode(&ilm); err != nil {
		return nil, diag.FromErr(err)
	}

	if ilm, ok := ilm[policyName]; ok {
		return &ilm, diags
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find a ILM policy in the cluster",
		Detail:   fmt.Sprintf(`Unable to find "%s" ILM policy in the cluster`, policyName),
	})
	return nil, diags
}

func (a *ApiClient) DeleteElasticsearchIlm(policyName string) diag.Diagnostics {
	var diags diag.Diagnostics

	res, err := a.es.ILM.DeleteLifecycle(policyName)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete ILM policy."); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchIndexTemplate(template *models.IndexTemplate) diag.Diagnostics {
	var diags diag.Diagnostics
	templateBytes, err := json.Marshal(template)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending request to ES: %s to create template '%s' ", templateBytes, template.Name)

	res, err := a.es.Indices.PutIndexTemplate(template.Name, bytes.NewReader(templateBytes))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create index template"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchIndexTemplate(templateName string) (*models.IndexTemplateResponse, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Indices.GetIndexTemplate.WithName(templateName)
	res, err := a.es.Indices.GetIndexTemplate(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to request index template."); diags.HasError() {
		return nil, diags
	}

	var indexTemplates models.IndexTemplatesResponse
	if err := json.NewDecoder(res.Body).Decode(&indexTemplates); err != nil {
		return nil, diag.FromErr(err)
	}

	// we requested only 1 template
	if len(indexTemplates.IndexTemplates) != 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong number of templates returned",
			Detail:   fmt.Sprintf("Elasticsearch API returned %d when requsted '%s' template.", len(indexTemplates.IndexTemplates), templateName),
		})
		return nil, diags
	}
	tpl := indexTemplates.IndexTemplates[0]
	return &tpl, diags
}

func (a *ApiClient) DeleteElasticsearchIndexTemplate(templateName string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Indices.DeleteIndexTemplate(templateName)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to delete index template"); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchSnapshotRepository(repository *models.SnapshotRepository) diag.Diagnostics {
	var diags diag.Diagnostics
	snapRepoBytes, err := json.Marshal(repository)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending snapshot repository definition to ES API: %s", snapRepoBytes)
	res, err := a.es.Snapshot.CreateRepository(repository.Name, bytes.NewReader(snapRepoBytes))
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the snapshot repository"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchSnapshotRepository(name string) (*models.SnapshotRepository, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Snapshot.GetRepository.WithRepository(name)
	res, err := a.es.Snapshot.GetRepository(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to get the information about snapshot repository: %s", name)); diags.HasError() {
		return nil, diags
	}
	snapRepoResponse := make(map[string]models.SnapshotRepository)
	if err := json.NewDecoder(res.Body).Decode(&snapRepoResponse); err != nil {
		return nil, diag.FromErr(err)
	}
	log.Printf("[TRACE] response ES API snapshot repository: %+v", snapRepoResponse)

	if currentRepo, ok := snapRepoResponse[name]; ok {
		return &currentRepo, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find requested repository",
		Detail:   fmt.Sprintf(`Repository "%s" is missing in the ES API response`, name),
	})
	return nil, diags
}

func (a *ApiClient) DeleteElasticsearchSnapshotRepository(name string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.Snapshot.DeleteRepository([]string{name})
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete snapshot repository: %s", name)); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) PutElasticsearchSlm(slm *models.SnapshotPolicy) diag.Diagnostics {
	var diags diag.Diagnostics

	slmBytes, err := json.Marshal(slm)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] sending SLM to ES API: %s", slmBytes)
	req := a.es.SlmPutLifecycle.WithBody(bytes.NewReader(slmBytes))
	res, err := a.es.SlmPutLifecycle(slm.Id, req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to create or update the SLM"); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) GetElasticsearchSlm(slmName string) (*models.SnapshotPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.SlmGetLifecycle.WithPolicyID(slmName)
	res, err := a.es.SlmGetLifecycle(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to get SLM policy from ES API"); diags.HasError() {
		return nil, diags
	}
	type SlmReponse = map[string]struct {
		Policy models.SnapshotPolicy `json:"policy"`
	}
	var slmResponse SlmReponse
	if err := json.NewDecoder(res.Body).Decode(&slmResponse); err != nil {
		return nil, diag.FromErr(err)
	}
	if slm, ok := slmResponse[slmName]; ok {
		return &slm.Policy, diags
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to find the SLM policy in the response",
		Detail:   fmt.Sprintf(`Unable to find "%s" policy in the ES API response.`, slmName),
	})
	return nil, diags
}

func (a *ApiClient) DeleteElasticsearchSlm(slmName string) diag.Diagnostics {
	var diags diag.Diagnostics
	res, err := a.es.SlmDeleteLifecycle(slmName)
	if err != nil {
		return diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, fmt.Sprintf("Unable to delete SLM policy: %s", slmName)); diags.HasError() {
		return diags
	}

	return diags
}

func (a *ApiClient) PutElasticsearchSettings(settings map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	settingsBytes, err := json.Marshal(settings)
	if err != nil {
		diag.FromErr(err)
	}
	log.Printf("[TRACE] settings to set: %s", settingsBytes)
	res, err := a.es.Cluster.PutSettings(bytes.NewReader(settingsBytes))
	if err != nil {
		diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to update cluster settings."); diags.HasError() {
		return diags
	}
	return diags
}

func (a *ApiClient) GetElasticsearchSettings() (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := a.es.Cluster.GetSettings.WithFlatSettings(true)
	res, err := a.es.Cluster.GetSettings(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()
	if diags := utils.CheckError(res, "Unable to read cluster settings."); diags.HasError() {
		return nil, diags
	}

	clusterSettings := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&clusterSettings); err != nil {
		return nil, diag.FromErr(err)
	}
	return clusterSettings, diags
}
