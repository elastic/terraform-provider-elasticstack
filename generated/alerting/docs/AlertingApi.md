# \AlertingAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRule**](AlertingAPI.md#CreateRule) | **Post** /s/{spaceId}/api/alerting/rule | Creates a rule with a randomly generated rule identifier.
[**CreateRuleId**](AlertingAPI.md#CreateRuleId) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId} | Creates a rule with a specific rule identifier.
[**DeleteRule**](AlertingAPI.md#DeleteRule) | **Delete** /s/{spaceId}/api/alerting/rule/{ruleId} | Deletes a rule.
[**DisableRule**](AlertingAPI.md#DisableRule) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/_disable | Disables a rule.
[**EnableRule**](AlertingAPI.md#EnableRule) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/_enable | Enables a rule.
[**FindRules**](AlertingAPI.md#FindRules) | **Get** /s/{spaceId}/api/alerting/rules/_find | Retrieves information about rules.
[**GetAlertingHealth**](AlertingAPI.md#GetAlertingHealth) | **Get** /s/{spaceId}/api/alerting/_health | Retrieves the health status of the alerting framework.
[**GetRule**](AlertingAPI.md#GetRule) | **Get** /s/{spaceId}/api/alerting/rule/{ruleId} | Retrieves a rule by its identifier.
[**GetRuleTypes**](AlertingAPI.md#GetRuleTypes) | **Get** /s/{spaceId}/api/alerting/rule_types | Retrieves a list of rule types.
[**LegacyCreateAlert**](AlertingAPI.md#LegacyCreateAlert) | **Post** /s/{spaceId}/api/alerts/alert/{alertId} | Create an alert.
[**LegacyDisableAlert**](AlertingAPI.md#LegacyDisableAlert) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/_disable | Disables an alert.
[**LegacyEnableAlert**](AlertingAPI.md#LegacyEnableAlert) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/_enable | Enables an alert.
[**LegacyFindAlerts**](AlertingAPI.md#LegacyFindAlerts) | **Get** /s/{spaceId}/api/alerts/alerts/_find | Retrieves a paginated set of alerts.
[**LegacyGetAlert**](AlertingAPI.md#LegacyGetAlert) | **Get** /s/{spaceId}/api/alerts/alert/{alertId} | Retrieves an alert by its identifier.
[**LegacyGetAlertTypes**](AlertingAPI.md#LegacyGetAlertTypes) | **Get** /s/{spaceId}/api/alerts/alerts/list_alert_types | Retrieves a list of alert types.
[**LegacyGetAlertingHealth**](AlertingAPI.md#LegacyGetAlertingHealth) | **Get** /s/{spaceId}/api/alerts/alerts/_health | Retrieves the health status of the alerting framework.
[**LegacyMuteAlertInstance**](AlertingAPI.md#LegacyMuteAlertInstance) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/alert_instance/{alertInstanceId}/_mute | Mutes an alert instance.
[**LegacyMuteAllAlertInstances**](AlertingAPI.md#LegacyMuteAllAlertInstances) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/_mute_all | Mutes all alert instances.
[**LegacyUnmuteAlertInstance**](AlertingAPI.md#LegacyUnmuteAlertInstance) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/alert_instance/{alertInstanceId}/_unmute | Unmutes an alert instance.
[**LegacyUnmuteAllAlertInstances**](AlertingAPI.md#LegacyUnmuteAllAlertInstances) | **Post** /s/{spaceId}/api/alerts/alert/{alertId}/_unmute_all | Unmutes all alert instances.
[**LegacyUpdateAlert**](AlertingAPI.md#LegacyUpdateAlert) | **Put** /s/{spaceId}/api/alerts/alert/{alertId} | Updates the attributes for an alert.
[**LegaryDeleteAlert**](AlertingAPI.md#LegaryDeleteAlert) | **Delete** /s/{spaceId}/api/alerts/alert/{alertId} | Permanently removes an alert.
[**MuteAlert**](AlertingAPI.md#MuteAlert) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/alert/{alertId}/_mute | Mutes an alert.
[**MuteAllAlerts**](AlertingAPI.md#MuteAllAlerts) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/_mute_all | Mutes all alerts.
[**UnmuteAlert**](AlertingAPI.md#UnmuteAlert) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/alert/{alertId}/_unmute | Unmutes an alert.
[**UnmuteAllAlerts**](AlertingAPI.md#UnmuteAllAlerts) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/_unmute_all | Unmutes all alerts.
[**UpdateRule**](AlertingAPI.md#UpdateRule) | **Put** /s/{spaceId}/api/alerting/rule/{ruleId} | Updates the attributes for a rule.
[**UpdateRuleAPIKey**](AlertingAPI.md#UpdateRuleAPIKey) | **Post** /s/{spaceId}/api/alerting/rule/{ruleId}/_update_api_key | Updates the API key for a rule.



## CreateRule

> RuleResponseProperties CreateRule(ctx, spaceId).KbnXsrf(kbnXsrf).CreateRuleRequest(createRuleRequest).Execute()

Creates a rule with a randomly generated rule identifier.



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    createRuleRequest := *openapiclient.NewCreateRuleRequest("Consumer_example", "cluster_health_rule", map[string]interface{}{"key": interface{}(123)}, "RuleTypeId_example", *openapiclient.NewSchedule()) // CreateRuleRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.CreateRule(context.Background(), spaceId).KbnXsrf(kbnXsrf).CreateRuleRequest(createRuleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.CreateRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateRule`: RuleResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.CreateRule`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 

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


## CreateRuleId

> RuleResponseProperties CreateRuleId(ctx, spaceId, ruleId).KbnXsrf(kbnXsrf).CreateRuleRequest(createRuleRequest).Execute()

Creates a rule with a specific rule identifier.



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    ruleId := "ruleId_example" // string | An UUID v1 or v4 identifier for the rule. If you omit this parameter, an identifier is randomly generated. 
    createRuleRequest := *openapiclient.NewCreateRuleRequest("Consumer_example", "cluster_health_rule", map[string]interface{}{"key": interface{}(123)}, "RuleTypeId_example", *openapiclient.NewSchedule()) // CreateRuleRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.CreateRuleId(context.Background(), spaceId, ruleId).KbnXsrf(kbnXsrf).CreateRuleRequest(createRuleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.CreateRuleId``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateRuleId`: RuleResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.CreateRuleId`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**ruleId** | **string** | An UUID v1 or v4 identifier for the rule. If you omit this parameter, an identifier is randomly generated.  | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateRuleIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 


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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    ruleId := TODO // interface{} | An identifier for the rule.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.DeleteRule(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.DeleteRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | [**interface{}**](.md) | An identifier for the rule. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    ruleId := TODO // interface{} | An identifier for the rule.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.DisableRule(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.DisableRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | [**interface{}**](.md) | An identifier for the rule. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDisableRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    ruleId := TODO // interface{} | An identifier for the rule.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.EnableRule(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.EnableRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | [**interface{}**](.md) | An identifier for the rule. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiEnableRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    defaultSearchOperator := "defaultSearchOperator_example" // string | The default operator to use for the simple_query_string. (optional) (default to "OR")
    fields := []*string{"Inner_example"} // []*string | The fields to return in the `attributes` key of the response. (optional)
    filter := "filter_example" // string | A KQL string that you filter with an attribute from your saved object. It should look like `savedObjectType.attributes.title: \"myTitle\"`. However, if you used a direct attribute of a saved object, such as `updatedAt`, you must define your filter, for example, `savedObjectType.updatedAt > 2018-12-22`.  (optional)
    hasReference := *openapiclient.NewFindRulesHasReferenceParameter() // FindRulesHasReferenceParameter | Filters the rules that have a relation with the reference objects with a specific type and identifier. (optional)
    page := int32(56) // int32 | The page number to return. (optional) (default to 1)
    perPage := int32(56) // int32 | The number of rules to return per page. (optional) (default to 20)
    search := "search_example" // string | An Elasticsearch simple_query_string query that filters the objects in the response. (optional)
    searchFields := "searchFields_example" // string | The fields to perform the simple_query_string parsed query against. (optional)
    sortField := "sortField_example" // string | Determines which field is used to sort the results. The field must exist in the `attributes` key of the response.  (optional)
    sortOrder := "sortOrder_example" // string | Determines the sort order. (optional) (default to "desc")

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.FindRules(context.Background(), spaceId).DefaultSearchOperator(defaultSearchOperator).Fields(fields).Filter(filter).HasReference(hasReference).Page(page).PerPage(perPage).Search(search).SearchFields(searchFields).SortField(sortField).SortOrder(sortOrder).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.FindRules``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `FindRules`: FindRules200Response
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.FindRules`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

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
 **searchFields** | **string** | The fields to perform the simple_query_string parsed query against. | 
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
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.GetAlertingHealth(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.GetAlertingHealth``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAlertingHealth`: GetAlertingHealth200Response
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.GetAlertingHealth`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

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
    ruleId := TODO // interface{} | An identifier for the rule.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.GetRule(context.Background(), ruleId, spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.GetRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetRule`: RuleResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.GetRule`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | [**interface{}**](.md) | An identifier for the rule. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

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
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.GetRuleTypes(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.GetRuleTypes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetRuleTypes`: []GetRuleTypes200ResponseInner
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.GetRuleTypes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    alertId := "alertId_example" // string | An UUID v1 or v4 identifier for the alert. If this parameter is omitted, the identifier is randomly generated.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    legacyCreateAlertRequestProperties := *openapiclient.NewLegacyCreateAlertRequestProperties("AlertTypeId_example", "Consumer_example", "Name_example", "NotifyWhen_example", map[string]interface{}(123), *openapiclient.NewLegacyUpdateAlertRequestPropertiesSchedule()) // LegacyCreateAlertRequestProperties | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.LegacyCreateAlert(context.Background(), alertId, spaceId).KbnXsrf(kbnXsrf).LegacyCreateAlertRequestProperties(legacyCreateAlertRequestProperties).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyCreateAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyCreateAlert`: AlertResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.LegacyCreateAlert`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**alertId** | **string** | An UUID v1 or v4 identifier for the alert. If this parameter is omitted, the identifier is randomly generated. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyCreateAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 


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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "alertId_example" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.LegacyDisableAlert(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyDisableAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyDisableAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "alertId_example" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.LegacyEnableAlert(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyEnableAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyEnableAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    defaultSearchOperator := "defaultSearchOperator_example" // string | The default operator to use for the `simple_query_string`. (optional) (default to "OR")
    fields := []string{"Inner_example"} // []string | The fields to return in the `attributes` key of the response. (optional)
    filter := "filter_example" // string | A KQL string that you filter with an attribute from your saved object. It should look like `savedObjectType.attributes.title: \"myTitle\"`. However, if you used a direct attribute of a saved object, such as `updatedAt`, you must define your filter, for example, `savedObjectType.updatedAt > 2018-12-22`.  (optional)
    hasReference := *openapiclient.NewLegacyFindAlertsHasReferenceParameter() // LegacyFindAlertsHasReferenceParameter | Filters the rules that have a relation with the reference objects with a specific type and identifier. (optional)
    page := int32(56) // int32 | The page number to return. (optional) (default to 1)
    perPage := int32(56) // int32 | The number of alerts to return per page. (optional) (default to 20)
    search := "search_example" // string | An Elasticsearch `simple_query_string` query that filters the alerts in the response. (optional)
    searchFields := "searchFields_example" // string | The fields to perform the `simple_query_string` parsed query against. (optional)
    sortField := "sortField_example" // string | Determines which field is used to sort the results. The field must exist in the `attributes` key of the response.  (optional)
    sortOrder := "sortOrder_example" // string | Determines the sort order. (optional) (default to "desc")

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.LegacyFindAlerts(context.Background(), spaceId).DefaultSearchOperator(defaultSearchOperator).Fields(fields).Filter(filter).HasReference(hasReference).Page(page).PerPage(perPage).Search(search).SearchFields(searchFields).SortField(sortField).SortOrder(sortOrder).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyFindAlerts``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyFindAlerts`: LegacyFindAlerts200Response
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.LegacyFindAlerts`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyFindAlertsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **defaultSearchOperator** | **string** | The default operator to use for the &#x60;simple_query_string&#x60;. | [default to &quot;OR&quot;]
 **fields** | **[]string** | The fields to return in the &#x60;attributes&#x60; key of the response. | 
 **filter** | **string** | A KQL string that you filter with an attribute from your saved object. It should look like &#x60;savedObjectType.attributes.title: \&quot;myTitle\&quot;&#x60;. However, if you used a direct attribute of a saved object, such as &#x60;updatedAt&#x60;, you must define your filter, for example, &#x60;savedObjectType.updatedAt &gt; 2018-12-22&#x60;.  | 
 **hasReference** | [**LegacyFindAlertsHasReferenceParameter**](LegacyFindAlertsHasReferenceParameter.md) | Filters the rules that have a relation with the reference objects with a specific type and identifier. | 
 **page** | **int32** | The page number to return. | [default to 1]
 **perPage** | **int32** | The number of alerts to return per page. | [default to 20]
 **search** | **string** | An Elasticsearch &#x60;simple_query_string&#x60; query that filters the alerts in the response. | 
 **searchFields** | **string** | The fields to perform the &#x60;simple_query_string&#x60; parsed query against. | 
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
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "alertId_example" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.LegacyGetAlert(context.Background(), spaceId, alertId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyGetAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyGetAlert`: AlertResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.LegacyGetAlert`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
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
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.LegacyGetAlertTypes(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyGetAlertTypes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyGetAlertTypes`: []LegacyGetAlertTypes200ResponseInner
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.LegacyGetAlertTypes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

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
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.LegacyGetAlertingHealth(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyGetAlertingHealth``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyGetAlertingHealth`: LegacyGetAlertingHealth200Response
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.LegacyGetAlertingHealth`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "alertId_example" // string | An identifier for the alert.
    alertInstanceId := "alertInstanceId_example" // string | An identifier for the alert instance.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.LegacyMuteAlertInstance(context.Background(), spaceId, alertId, alertInstanceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyMuteAlertInstance``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | An identifier for the alert. | 
**alertInstanceId** | **string** | An identifier for the alert instance. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyMuteAlertInstanceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 




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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "alertId_example" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.LegacyMuteAllAlertInstances(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyMuteAllAlertInstances``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyMuteAllAlertInstancesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "alertId_example" // string | An identifier for the alert.
    alertInstanceId := "alertInstanceId_example" // string | An identifier for the alert instance.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.LegacyUnmuteAlertInstance(context.Background(), spaceId, alertId, alertInstanceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyUnmuteAlertInstance``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | An identifier for the alert. | 
**alertInstanceId** | **string** | An identifier for the alert instance. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyUnmuteAlertInstanceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 




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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "alertId_example" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.LegacyUnmuteAllAlertInstances(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyUnmuteAllAlertInstances``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyUnmuteAllAlertInstancesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "alertId_example" // string | The identifier for the alert.
    legacyUpdateAlertRequestProperties := *openapiclient.NewLegacyUpdateAlertRequestProperties("Name_example", "NotifyWhen_example", map[string]interface{}(123), *openapiclient.NewLegacyUpdateAlertRequestPropertiesSchedule()) // LegacyUpdateAlertRequestProperties | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.LegacyUpdateAlert(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).LegacyUpdateAlertRequestProperties(legacyUpdateAlertRequestProperties).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegacyUpdateAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyUpdateAlert`: AlertResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.LegacyUpdateAlert`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyUpdateAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 


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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    alertId := "alertId_example" // string | The identifier for the alert.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.LegaryDeleteAlert(context.Background(), spaceId, alertId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.LegaryDeleteAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**alertId** | **string** | The identifier for the alert. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegaryDeleteAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    alertId := TODO // interface{} | An identifier for the alert. The identifier is generated by the rule and might be any arbitrary string.
    ruleId := TODO // interface{} | An identifier for the rule.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.MuteAlert(context.Background(), alertId, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.MuteAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**alertId** | [**interface{}**](.md) | An identifier for the alert. The identifier is generated by the rule and might be any arbitrary string. | 
**ruleId** | [**interface{}**](.md) | An identifier for the rule. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiMuteAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 




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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    ruleId := TODO // interface{} | An identifier for the rule.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.MuteAllAlerts(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.MuteAllAlerts``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | [**interface{}**](.md) | An identifier for the rule. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiMuteAllAlertsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    alertId := TODO // interface{} | An identifier for the alert. The identifier is generated by the rule and might be any arbitrary string.
    ruleId := TODO // interface{} | An identifier for the rule.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.UnmuteAlert(context.Background(), alertId, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.UnmuteAlert``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**alertId** | [**interface{}**](.md) | An identifier for the alert. The identifier is generated by the rule and might be any arbitrary string. | 
**ruleId** | [**interface{}**](.md) | An identifier for the rule. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUnmuteAlertRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 




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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    ruleId := TODO // interface{} | An identifier for the rule.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.UnmuteAllAlerts(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.UnmuteAllAlerts``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | [**interface{}**](.md) | An identifier for the rule. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUnmuteAllAlertsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    ruleId := TODO // interface{} | An identifier for the rule.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    updateRuleRequest := *openapiclient.NewUpdateRuleRequest("Name_example", map[string]interface{}{"key": interface{}(123)}, *openapiclient.NewSchedule()) // UpdateRuleRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AlertingAPI.UpdateRule(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).UpdateRuleRequest(updateRuleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.UpdateRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateRule`: RuleResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `AlertingAPI.UpdateRule`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | [**interface{}**](.md) | An identifier for the rule. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 


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


## UpdateRuleAPIKey

> UpdateRuleAPIKey(ctx, ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()

Updates the API key for a rule.



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
    kbnXsrf := TODO // interface{} | Cross-site request forgery protection
    ruleId := TODO // interface{} | An identifier for the rule.
    spaceId := TODO // interface{} | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.AlertingAPI.UpdateRuleAPIKey(context.Background(), ruleId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AlertingAPI.UpdateRuleAPIKey``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**ruleId** | [**interface{}**](.md) | An identifier for the rule. | 
**spaceId** | [**interface{}**](.md) | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRuleAPIKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | [**interface{}**](interface{}.md) | Cross-site request forgery protection | 



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

