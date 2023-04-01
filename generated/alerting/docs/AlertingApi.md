# \AlertingApi

All URIs are relative to *http://localhost:5601*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRule**](AlertingApi.md#CreateRule) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId} | Creates a rule.
[**DeleteRule**](AlertingApi.md#DeleteRule) | **Delete** /s/{spaceId}/api/alerting/rule/{ruleId} | Deletes a rule.
[**DisableRule**](AlertingApi.md#DisableRule) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/_disable | Disables a rule.
[**EnableRule**](AlertingApi.md#EnableRule) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/_enable | Enables a rule.
[**FindRules**](AlertingApi.md#FindRules) | **Get** /s/{spaceId}/api/alerting/rules/_find | Retrieves information about rules.
[**GetAlertingHealth**](AlertingApi.md#GetAlertingHealth) | **Get** /s/{spaceId}/api/alerting/_health | Retrieves the health status of the alerting framework.
[**GetRule**](AlertingApi.md#GetRule) | **Get** /s/{spaceId}/api/alerting/rule/{ruleId} | Retrieves a rule by its identifier.
[**GetRuleTypes**](AlertingApi.md#GetRuleTypes) | **Get** /s/{spaceId}/api/alerting/rule_types | Retrieves a list of rule types.
[**LegacyCreateAlert**](AlertingApi.md#LegacyCreateAlert) | **Post** /s/{spaceId}/api/alerts/alert/{alertId} | Create an alert.
[**LegacyDisableAlert**](AlertingApi.md#LegacyDisableAlert) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/_disable | Disables an alert.
[**LegacyEnableAlert**](AlertingApi.md#LegacyEnableAlert) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/_enable | Enables an alert.
[**LegacyFindAlerts**](AlertingApi.md#LegacyFindAlerts) | **Get** /s/{spaceId}/api/alerts/alerts/_find | Retrieves a paginated set of alerts.
[**LegacyGetAlert**](AlertingApi.md#LegacyGetAlert) | **Get** /s/{spaceId}/api/alerts/alert/{alertId} | Retrieves an alert by its identifier.
[**LegacyGetAlertTypes**](AlertingApi.md#LegacyGetAlertTypes) | **Get** /s/{spaceId}/api/alerts/alerts/list_alert_types | Retrieves a list of alert types.
[**LegacyGetAlertingHealth**](AlertingApi.md#LegacyGetAlertingHealth) | **Get** /s/{spaceId}/api/alerts/alerts/_health | Retrieves the health status of the alerting framework.
[**LegacyMuteAlertInstance**](AlertingApi.md#LegacyMuteAlertInstance) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/alert_instance/{alertInstanceId}/_mute | Mutes an alert instance.
[**LegacyMuteAllAlertInstances**](AlertingApi.md#LegacyMuteAllAlertInstances) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/_mute_all | Mutes all alert instances.
[**LegacyUnmuteAlertInstance**](AlertingApi.md#LegacyUnmuteAlertInstance) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/alert_instance/{alertInstanceId}/_unmute | Unmutes an alert instance.
[**LegacyUnmuteAllAlertInstances**](AlertingApi.md#LegacyUnmuteAllAlertInstances) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/_unmute_all | Unmutes all alert instances.
[**LegacyUpdateAlert**](AlertingApi.md#LegacyUpdateAlert) | **Put** /s/{spaceId}/api/alerts/alert/{alertId} | Updates the attributes for an alert.
[**LegaryDeleteAlert**](AlertingApi.md#LegaryDeleteAlert) | **Delete** /s/{spaceId}/api/alerts/alert/{alertId} | Permanently removes an alert.
[**MuteAlert**](AlertingApi.md#MuteAlert) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/alert/{alertId}/_mute | Mutes an alert.
[**MuteAllAlerts**](AlertingApi.md#MuteAllAlerts) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/_mute_all | Mutes all alerts.
[**UnmuteAlert**](AlertingApi.md#UnmuteAlert) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/alert/{alertId}/_unmute | Unmutes an alert.
[**UnmuteAllAlerts**](AlertingApi.md#UnmuteAllAlerts) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/_unmute_all | Unmutes all alerts.
[**UpdateRule**](AlertingApi.md#UpdateRule) | **Put** /s/{spaceId}/api/alerting/rule/{ruleId} | Updates the attributes for a rule.



## CreateRule

> RuleResponseProperties CreateRule(ctx, spaceId, ruleId).KbnXsrf(kbnXsrf).CreateRuleRequest(createRuleRequest).Execute()

Creates a rule.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    ruleId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An UUID v1 or v4 identifier for the rule. If you omit this parameter, an identifier is randomly generated. 
    createRuleRequest := *openapiclient.NewCreateRuleRequest("Consumer_example", "cluster_health_rule", map[string]interface{}{"key": interface{}(123)}, "RuleTypeId_example", *openapiclient.NewSchedule()) // CreateRuleRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.CreateRule(context.Background(), spaceId, ruleId).KbnXsrf(kbnXsrf).CreateRuleRequest(createRuleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.CreateRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateRule`: RuleResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.CreateRule`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**ruleId** | **string** | An UUID v1 or v4 identifier for the rule. If you omit this parameter, an identifier is randomly generated.  | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 


 **createRuleRequest** | [**CreateRuleRequest**](CreateRuleRequest.md) |  | 

### Return type

[**RuleResponseProperties**](RuleResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRule

> DeleteRule(ctx, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()

Deletes a rule.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    ruleId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the rule.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.DeleteRule(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.DeleteRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | **string** | An identifier for the rule. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DisableRule

> DisableRule(ctx, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()

Disables a rule.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    ruleId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the rule.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.DisableRule(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.DisableRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | **string** | An identifier for the rule. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDisableRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## EnableRule

> EnableRule(ctx, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()

Enables a rule.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    ruleId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the rule.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.EnableRule(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.EnableRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | **string** | An identifier for the rule. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiEnableRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FindRules

> FindRules200Response FindRules(ctx, spaceId).DefaultSearchOperator(defaultSearchOperator).Fields(fields).Filter(filter).HasReference(hasReference).Page(page).PerPage(perPage).Search(search).SearchFields(searchFields).SortField(sortField).SortOrder(sortOrder).Execute()

Retrieves information about rules.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    defaultSearchOperator := "OR" // string | The default operator to use for the simple_query_string. (optional) (default to "OR")
    fields := []string{"Inner_example"} // []string | The fields to return in the `attributes` key of the response. (optional)
    filter := "filter_example" // string | A KQL string that you filter with an attribute from your saved object. It should look like `savedObjectType.attributes.title: \"myTitle\"`. However, if you used a direct attribute of a saved object, such as `updatedAt`, you must define your filter, for example, `savedObjectType.updatedAt > 2018-12-22`.  (optional)
    hasReference := map[string][]openapiclient.FindRulesHasReferenceParameter{ ... } // FindRulesHasReferenceParameter | Filters the rules that have a relation with the reference objects with a specific type and identifier. (optional)
    page := int32(1) // int32 | The page number to return. (optional) (default to 1)
    perPage := int32(20) // int32 | The number of rules to return per page. (optional) (default to 20)
    search := "search_example" // string | An Elasticsearch simple_query_string query that filters the objects in the response. (optional)
    searchFields := openapiclient.findRules_search_fields_parameter{ArrayOfString: new([]string)} // FindRulesSearchFieldsParameter | The fields to perform the simple_query_string parsed query against. (optional)
    sortField := "sortField_example" // string | Determines which field is used to sort the results. The field must exist in the `attributes` key of the response.  (optional)
    sortOrder := "asc" // string | Determines the sort order. (optional) (default to "desc")

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.FindRules(context.Background(), spaceId).DefaultSearchOperator(defaultSearchOperator).Fields(fields).Filter(filter).HasReference(hasReference).Page(page).PerPage(perPage).Search(search).SearchFields(searchFields).SortField(sortField).SortOrder(sortOrder).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.FindRules``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `FindRules`: FindRules200Response
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.FindRules`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiFindRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **defaultSearchOperator** | **string** | The default operator to use for the simple_query_string. | [default to &quot;OR&quot;]
 **fields** | **[]string** | The fields to return in the &#x60;attributes&#x60; key of the response. | 
 **filter** | **string** | A KQL string that you filter with an attribute from your saved object. It should look like &#x60;savedObjectType.attributes.title: \&quot;myTitle\&quot;&#x60;. However, if you used a direct attribute of a saved object, such as &#x60;updatedAt&#x60;, you must define your filter, for example, &#x60;savedObjectType.updatedAt &gt; 2018-12-22&#x60;.  | 
 **hasReference** | [**FindRulesHasReferenceParameter**](FindRulesHasReferenceParameter.md) | Filters the rules that have a relation with the reference objects with a specific type and identifier. | 
 **page** | **int32** | The page number to return. | [default to 1]
 **perPage** | **int32** | The number of rules to return per page. | [default to 20]
 **search** | **string** | An Elasticsearch simple_query_string query that filters the objects in the response. | 
 **searchFields** | [**FindRulesSearchFieldsParameter**](FindRulesSearchFieldsParameter.md) | The fields to perform the simple_query_string parsed query against. | 
 **sortField** | **string** | Determines which field is used to sort the results. The field must exist in the &#x60;attributes&#x60; key of the response.  | 
 **sortOrder** | **string** | Determines the sort order. | [default to &quot;desc&quot;]

### Return type

[**FindRules200Response**](FindRules200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAlertingHealth

> GetAlertingHealth200Response GetAlertingHealth(ctx, spaceId).Execute()

Retrieves the health status of the alerting framework.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.GetAlertingHealth(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.GetAlertingHealth``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAlertingHealth`: GetAlertingHealth200Response
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.GetAlertingHealth`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAlertingHealthRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetAlertingHealth200Response**](GetAlertingHealth200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRule

> RuleResponseProperties GetRule(ctx, ruleId, spaceId).Execute()

Retrieves a rule by its identifier.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    ruleId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the rule.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.GetRule(context.Background(), ruleId, spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.GetRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetRule`: RuleResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.GetRule`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | **string** | An identifier for the rule. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**RuleResponseProperties**](RuleResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRuleTypes

> []GetRuleTypes200ResponseInner GetRuleTypes(ctx, spaceId).Execute()

Retrieves a list of rule types.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.GetRuleTypes(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.GetRuleTypes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetRuleTypes`: []GetRuleTypes200ResponseInner
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.GetRuleTypes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRuleTypesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]GetRuleTypes200ResponseInner**](GetRuleTypes200ResponseInner.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyCreateAlert

> AlertResponseProperties LegacyCreateAlert(ctx, alertId, spaceId).KbnXsrf(kbnXsrf).LegacyCreateAlertRequestProperties(legacyCreateAlertRequestProperties).Execute()

Create an alert.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    alertId := "41893910-6bca-11eb-9e0d-85d233e3ee35" // string | An UUID v1 or v4 identifier for the alert. If this parameter is omitted, the identifier is randomly generated.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    legacyCreateAlertRequestProperties := *openapiclient.NewLegacyCreateAlertRequestProperties("AlertTypeId_example", "Consumer_example", "Name_example", "NotifyWhen_example", map[string]interface{}(123), *openapiclient.NewLegacyCreateAlertRequestPropertiesSchedule()) // LegacyCreateAlertRequestProperties | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.LegacyCreateAlert(context.Background(), alertId, spaceId).KbnXsrf(kbnXsrf).LegacyCreateAlertRequestProperties(legacyCreateAlertRequestProperties).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyCreateAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyCreateAlert`: AlertResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.LegacyCreateAlert`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**alertId** | **string** | An UUID v1 or v4 identifier for the alert. If this parameter is omitted, the identifier is randomly generated. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyCreateAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 


 **legacyCreateAlertRequestProperties** | [**LegacyCreateAlertRequestProperties**](LegacyCreateAlertRequestProperties.md) |  | 

### Return type

[**AlertResponseProperties**](AlertResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyDisableAlert

> LegacyDisableAlert(ctx, spaceId, alertId).KbnXsrf(kbnXsrf).Execute()

Disables an alert.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "41893910-6bca-11eb-9e0d-85d233e3ee35" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.LegacyDisableAlert(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyDisableAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyDisableAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyEnableAlert

> LegacyEnableAlert(ctx, spaceId, alertId).KbnXsrf(kbnXsrf).Execute()

Enables an alert.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "41893910-6bca-11eb-9e0d-85d233e3ee35" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.LegacyEnableAlert(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyEnableAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyEnableAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyFindAlerts

> LegacyFindAlerts200Response LegacyFindAlerts(ctx, spaceId).DefaultSearchOperator(defaultSearchOperator).Fields(fields).Filter(filter).HasReference(hasReference).Page(page).PerPage(perPage).Search(search).SearchFields(searchFields).SortField(sortField).SortOrder(sortOrder).Execute()

Retrieves a paginated set of alerts.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    defaultSearchOperator := "OR" // string | The default operator to use for the `simple_query_string`. (optional) (default to "OR")
    fields := []string{"Inner_example"} // []string | The fields to return in the `attributes` key of the response. (optional)
    filter := "filter_example" // string | A KQL string that you filter with an attribute from your saved object. It should look like `savedObjectType.attributes.title: \"myTitle\"`. However, if you used a direct attribute of a saved object, such as `updatedAt`, you must define your filter, for example, `savedObjectType.updatedAt > 2018-12-22`.  (optional)
    hasReference := map[string][]openapiclient.FindRulesHasReferenceParameter{ ... } // FindRulesHasReferenceParameter | Filters the rules that have a relation with the reference objects with a specific type and identifier. (optional)
    page := int32(1) // int32 | The page number to return. (optional) (default to 1)
    perPage := int32(20) // int32 | The number of alerts to return per page. (optional) (default to 20)
    search := "search_example" // string | An Elasticsearch `simple_query_string` query that filters the alerts in the response. (optional)
    searchFields := openapiclient.findRules_search_fields_parameter{ArrayOfString: new([]string)} // FindRulesSearchFieldsParameter | The fields to perform the `simple_query_string` parsed query against. (optional)
    sortField := "sortField_example" // string | Determines which field is used to sort the results. The field must exist in the `attributes` key of the response.  (optional)
    sortOrder := "asc" // string | Determines the sort order. (optional) (default to "desc")

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.LegacyFindAlerts(context.Background(), spaceId).DefaultSearchOperator(defaultSearchOperator).Fields(fields).Filter(filter).HasReference(hasReference).Page(page).PerPage(perPage).Search(search).SearchFields(searchFields).SortField(sortField).SortOrder(sortOrder).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyFindAlerts``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyFindAlerts`: LegacyFindAlerts200Response
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.LegacyFindAlerts`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyFindAlertsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **defaultSearchOperator** | **string** | The default operator to use for the &#x60;simple_query_string&#x60;. | [default to &quot;OR&quot;]
 **fields** | **[]string** | The fields to return in the &#x60;attributes&#x60; key of the response. | 
 **filter** | **string** | A KQL string that you filter with an attribute from your saved object. It should look like &#x60;savedObjectType.attributes.title: \&quot;myTitle\&quot;&#x60;. However, if you used a direct attribute of a saved object, such as &#x60;updatedAt&#x60;, you must define your filter, for example, &#x60;savedObjectType.updatedAt &gt; 2018-12-22&#x60;.  | 
 **hasReference** | [**FindRulesHasReferenceParameter**](FindRulesHasReferenceParameter.md) | Filters the rules that have a relation with the reference objects with a specific type and identifier. | 
 **page** | **int32** | The page number to return. | [default to 1]
 **perPage** | **int32** | The number of alerts to return per page. | [default to 20]
 **search** | **string** | An Elasticsearch &#x60;simple_query_string&#x60; query that filters the alerts in the response. | 
 **searchFields** | [**FindRulesSearchFieldsParameter**](FindRulesSearchFieldsParameter.md) | The fields to perform the &#x60;simple_query_string&#x60; parsed query against. | 
 **sortField** | **string** | Determines which field is used to sort the results. The field must exist in the &#x60;attributes&#x60; key of the response.  | 
 **sortOrder** | **string** | Determines the sort order. | [default to &quot;desc&quot;]

### Return type

[**LegacyFindAlerts200Response**](LegacyFindAlerts200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyGetAlert

> AlertResponseProperties LegacyGetAlert(ctx, spaceId, alertId).Execute()

Retrieves an alert by its identifier.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "41893910-6bca-11eb-9e0d-85d233e3ee35" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.LegacyGetAlert(context.Background(), spaceId, alertId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyGetAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyGetAlert`: AlertResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.LegacyGetAlert`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyGetAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**AlertResponseProperties**](AlertResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyGetAlertTypes

> []LegacyGetAlertTypes200ResponseInner LegacyGetAlertTypes(ctx, spaceId).Execute()

Retrieves a list of alert types.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.LegacyGetAlertTypes(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyGetAlertTypes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyGetAlertTypes`: []LegacyGetAlertTypes200ResponseInner
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.LegacyGetAlertTypes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyGetAlertTypesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]LegacyGetAlertTypes200ResponseInner**](LegacyGetAlertTypes200ResponseInner.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyGetAlertingHealth

> LegacyGetAlertingHealth200Response LegacyGetAlertingHealth(ctx, spaceId).Execute()

Retrieves the health status of the alerting framework.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.LegacyGetAlertingHealth(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyGetAlertingHealth``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyGetAlertingHealth`: LegacyGetAlertingHealth200Response
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.LegacyGetAlertingHealth`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyGetAlertingHealthRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**LegacyGetAlertingHealth200Response**](LegacyGetAlertingHealth200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyMuteAlertInstance

> LegacyMuteAlertInstance(ctx, spaceId, alertId, alertInstanceId).KbnXsrf(kbnXsrf).Execute()

Mutes an alert instance.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "41893910-6bca-11eb-9e0d-85d233e3ee35" // string | An identifier for the alert.
    alertInstanceId := "dceeb5d0-6b41-11eb-802b-85b0c1bc8ba2" // string | An identifier for the alert instance.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.LegacyMuteAlertInstance(context.Background(), spaceId, alertId, alertInstanceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyMuteAlertInstance``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | An identifier for the alert. | 
**alertInstanceId** | **string** | An identifier for the alert instance. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyMuteAlertInstanceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 




### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyMuteAllAlertInstances

> LegacyMuteAllAlertInstances(ctx, spaceId, alertId).KbnXsrf(kbnXsrf).Execute()

Mutes all alert instances.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "41893910-6bca-11eb-9e0d-85d233e3ee35" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.LegacyMuteAllAlertInstances(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyMuteAllAlertInstances``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyMuteAllAlertInstancesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyUnmuteAlertInstance

> LegacyUnmuteAlertInstance(ctx, spaceId, alertId, alertInstanceId).KbnXsrf(kbnXsrf).Execute()

Unmutes an alert instance.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "41893910-6bca-11eb-9e0d-85d233e3ee35" // string | An identifier for the alert.
    alertInstanceId := "dceeb5d0-6b41-11eb-802b-85b0c1bc8ba2" // string | An identifier for the alert instance.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.LegacyUnmuteAlertInstance(context.Background(), spaceId, alertId, alertInstanceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyUnmuteAlertInstance``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | An identifier for the alert. | 
**alertInstanceId** | **string** | An identifier for the alert instance. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyUnmuteAlertInstanceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 




### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyUnmuteAllAlertInstances

> LegacyUnmuteAllAlertInstances(ctx, spaceId, alertId).KbnXsrf(kbnXsrf).Execute()

Unmutes all alert instances.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "41893910-6bca-11eb-9e0d-85d233e3ee35" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.LegacyUnmuteAllAlertInstances(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyUnmuteAllAlertInstances``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyUnmuteAllAlertInstancesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyUpdateAlert

> AlertResponseProperties LegacyUpdateAlert(ctx, spaceId, alertId).KbnXsrf(kbnXsrf).LegacyUpdateAlertRequestProperties(legacyUpdateAlertRequestProperties).Execute()

Updates the attributes for an alert.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "41893910-6bca-11eb-9e0d-85d233e3ee35" // string | The identifier for the alert.
    legacyUpdateAlertRequestProperties := *openapiclient.NewLegacyUpdateAlertRequestProperties("Name_example", "NotifyWhen_example", map[string]interface{}(123), *openapiclient.NewLegacyUpdateAlertRequestPropertiesSchedule()) // LegacyUpdateAlertRequestProperties | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.LegacyUpdateAlert(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).LegacyUpdateAlertRequestProperties(legacyUpdateAlertRequestProperties).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegacyUpdateAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyUpdateAlert`: AlertResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.LegacyUpdateAlert`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyUpdateAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 


 **legacyUpdateAlertRequestProperties** | [**LegacyUpdateAlertRequestProperties**](LegacyUpdateAlertRequestProperties.md) |  | 

### Return type

[**AlertResponseProperties**](AlertResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegaryDeleteAlert

> LegaryDeleteAlert(ctx, spaceId, alertId).KbnXsrf(kbnXsrf).Execute()

Permanently removes an alert.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "41893910-6bca-11eb-9e0d-85d233e3ee35" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.LegaryDeleteAlert(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.LegaryDeleteAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegaryDeleteAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MuteAlert

> MuteAlert(ctx, alertId, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()

Mutes an alert.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    alertId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the alert. The identifier is generated by the rule and might be any arbitrary string.
    ruleId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the rule.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.MuteAlert(context.Background(), alertId, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.MuteAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**alertId** | **string** | An identifier for the alert. The identifier is generated by the rule and might be any arbitrary string. | 
**ruleId** | **string** | An identifier for the rule. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiMuteAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 




### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MuteAllAlerts

> MuteAllAlerts(ctx, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()

Mutes all alerts.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    ruleId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the rule.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.MuteAllAlerts(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.MuteAllAlerts``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | **string** | An identifier for the rule. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiMuteAllAlertsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UnmuteAlert

> UnmuteAlert(ctx, alertId, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()

Unmutes an alert.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    alertId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the alert. The identifier is generated by the rule and might be any arbitrary string.
    ruleId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the rule.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.UnmuteAlert(context.Background(), alertId, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.UnmuteAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**alertId** | **string** | An identifier for the alert. The identifier is generated by the rule and might be any arbitrary string. | 
**ruleId** | **string** | An identifier for the rule. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUnmuteAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 




### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UnmuteAllAlerts

> UnmuteAllAlerts(ctx, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()

Unmutes all alerts.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    ruleId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the rule.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingApi.UnmuteAllAlerts(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.UnmuteAllAlerts``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | **string** | An identifier for the rule. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUnmuteAllAlertsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateRule

> RuleResponseProperties UpdateRule(ctx, ruleId, spaceId).KbnXsrf(kbnXsrf).UpdateRuleRequest(updateRuleRequest).Execute()

Updates the attributes for a rule.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/alerting"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    ruleId := "ac4e6b90-6be7-11eb-ba0d-9b1c1f912d74" // string | An identifier for the rule.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    updateRuleRequest := *openapiclient.NewUpdateRuleRequest("cluster_health_rule", map[string]interface{}{"key": interface{}(123)}, *openapiclient.NewSchedule()) // UpdateRuleRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingApi.UpdateRule(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).UpdateRuleRequest(updateRuleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingApi.UpdateRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateRule`: RuleResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingApi.UpdateRule`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | **string** | An identifier for the rule. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 


 **updateRuleRequest** | [**UpdateRuleRequest**](UpdateRuleRequest.md) |  | 

### Return type

[**RuleResponseProperties**](RuleResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

