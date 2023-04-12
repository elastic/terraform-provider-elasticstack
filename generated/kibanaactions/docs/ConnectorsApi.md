# \ConnectorsApi

All URIs are relative to *http://localhost:5601*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateConnector**](ConnectorsApi.md#CreateConnector) | **Post** /s/{spaceId}/api/actions/connector | Creates a connector.
[**DeleteConnector**](ConnectorsApi.md#DeleteConnector) | **Delete** /s/{spaceId}/api/actions/connector/{connectorId} | Deletes a connector.
[**GetConnector**](ConnectorsApi.md#GetConnector) | **Get** /s/{spaceId}/api/actions/connector/{connectorId} | Retrieves a connector by ID.
[**GetConnectorTypes**](ConnectorsApi.md#GetConnectorTypes) | **Get** /s/{spaceId}/api/actions/connector_types | Retrieves a list of all connector types.
[**GetConnectors**](ConnectorsApi.md#GetConnectors) | **Get** /s/{spaceId}/api/actions/connectors | Retrieves all connectors.
[**LegacyCreateConnector**](ConnectorsApi.md#LegacyCreateConnector) | **Post** /s/{spaceId}/api/actions | Creates a connector.
[**LegacyDeleteConnector**](ConnectorsApi.md#LegacyDeleteConnector) | **Delete** /s/{spaceId}/api/actions/action/{actionId} | Deletes a connector.
[**LegacyGetConnector**](ConnectorsApi.md#LegacyGetConnector) | **Get** /s/{spaceId}/api/actions/action/{actionId} | Retrieves a connector by ID.
[**LegacyGetConnectorTypes**](ConnectorsApi.md#LegacyGetConnectorTypes) | **Get** /s/{spaceId}/api/actions/list_action_types | Retrieves a list of all connector types.
[**LegacyGetConnectors**](ConnectorsApi.md#LegacyGetConnectors) | **Get** /s/{spaceId}/api/actions | Retrieves all connectors.
[**LegacyRunConnector**](ConnectorsApi.md#LegacyRunConnector) | **Post** /s/{spaceId}/api/actions/action/{actionId}/_execute | Runs a connector.
[**LegacyUpdateConnector**](ConnectorsApi.md#LegacyUpdateConnector) | **Put** /s/{spaceId}/api/actions/action/{actionId} | Updates the attributes for a connector.
[**RunConnector**](ConnectorsApi.md#RunConnector) | **Post** /s/{spaceId}/api/actions/connector/{connectorId}/_execute | Runs a connector.
[**UpdateConnector**](ConnectorsApi.md#UpdateConnector) | **Put** /s/{spaceId}/api/actions/connector/{connectorId} | Updates the attributes for a connector.



## CreateConnector

> ConnectorResponseProperties CreateConnector(ctx, spaceId).KbnXsrf(kbnXsrf).CreateConnectorRequestBodyProperties(createConnectorRequestBodyProperties).Execute()

Creates a connector.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    createConnectorRequestBodyProperties := openapiclient.Create_connector_request_body_properties{CreateConnectorRequestCasesWebhook: openapiclient.NewCreateConnectorRequestCasesWebhook(*openapiclient.NewConfigPropertiesCasesWebhook("{"fields":{"summary":{"[object Object]":null},"description":{"[object Object]":null},"labels":{"[object Object]":null}}}", "CreateIncidentResponseKey_example", "CreateIncidentUrl_example", "GetIncidentResponseExternalTitleKey_example", "https://testing-jira.atlassian.net/rest/api/2/issue/{{{external.system.id}}}", "{"fields":{"summary":{"[object Object]":null},"description":{"[object Object]":null},"labels":{"[object Object]":null}}}", "https://testing-jira.atlassian.net/rest/api/2/issue/{{{external.system.ID}}}", "https://testing-jira.atlassian.net/browse/{{{external.system.title}}}"), ".cases-webhook", "my-connector")} // CreateConnectorRequestBodyProperties | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.CreateConnector(context.Background(), spaceId).KbnXsrf(kbnXsrf).CreateConnectorRequestBodyProperties(createConnectorRequestBodyProperties).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.CreateConnector``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateConnector`: ConnectorResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.CreateConnector`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateConnectorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **createConnectorRequestBodyProperties** | [**CreateConnectorRequestBodyProperties**](CreateConnectorRequestBodyProperties.md) |  | 

### Return type

[**ConnectorResponseProperties**](ConnectorResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteConnector

> DeleteConnector(ctx, connectorId, spaceId).KbnXsrf(kbnXsrf).Execute()

Deletes a connector.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    connectorId := "df770e30-8b8b-11ed-a780-3b746c987a81" // string | An identifier for the connector.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.ConnectorsApi.DeleteConnector(context.Background(), connectorId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.DeleteConnector``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**connectorId** | **string** | An identifier for the connector. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteConnectorRequest struct via the builder pattern


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


## GetConnector

> ConnectorResponseProperties GetConnector(ctx, connectorId, spaceId).Execute()

Retrieves a connector by ID.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    connectorId := "df770e30-8b8b-11ed-a780-3b746c987a81" // string | An identifier for the connector.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.GetConnector(context.Background(), connectorId, spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.GetConnector``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetConnector`: ConnectorResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.GetConnector`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**connectorId** | **string** | An identifier for the connector. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetConnectorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**ConnectorResponseProperties**](ConnectorResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetConnectorTypes

> []GetConnectorTypesResponseBodyPropertiesInner GetConnectorTypes(ctx, spaceId).FeatureId(featureId).Execute()

Retrieves a list of all connector types.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    featureId := openapiclient.features("alerting") // Features | A filter to limit the retrieved connector types to those that support a specific feature (such as alerting or cases). (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.GetConnectorTypes(context.Background(), spaceId).FeatureId(featureId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.GetConnectorTypes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetConnectorTypes`: []GetConnectorTypesResponseBodyPropertiesInner
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.GetConnectorTypes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetConnectorTypesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **featureId** | [**Features**](Features.md) | A filter to limit the retrieved connector types to those that support a specific feature (such as alerting or cases). | 

### Return type

[**[]GetConnectorTypesResponseBodyPropertiesInner**](GetConnectorTypesResponseBodyPropertiesInner.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetConnectors

> []GetConnectorsResponseBodyProperties GetConnectors(ctx, spaceId).Execute()

Retrieves all connectors.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.GetConnectors(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.GetConnectors``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetConnectors`: []GetConnectorsResponseBodyProperties
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.GetConnectors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetConnectorsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]GetConnectorsResponseBodyProperties**](GetConnectorsResponseBodyProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyCreateConnector

> ActionResponseProperties LegacyCreateConnector(ctx, spaceId).KbnXsrf(kbnXsrf).LegacyCreateConnectorRequestProperties(legacyCreateConnectorRequestProperties).Execute()

Creates a connector.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    legacyCreateConnectorRequestProperties := *openapiclient.NewLegacyCreateConnectorRequestProperties() // LegacyCreateConnectorRequestProperties | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.LegacyCreateConnector(context.Background(), spaceId).KbnXsrf(kbnXsrf).LegacyCreateConnectorRequestProperties(legacyCreateConnectorRequestProperties).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.LegacyCreateConnector``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyCreateConnector`: ActionResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.LegacyCreateConnector`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyCreateConnectorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **legacyCreateConnectorRequestProperties** | [**LegacyCreateConnectorRequestProperties**](LegacyCreateConnectorRequestProperties.md) |  | 

### Return type

[**ActionResponseProperties**](ActionResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyDeleteConnector

> LegacyDeleteConnector(ctx, actionId, spaceId).KbnXsrf(kbnXsrf).Execute()

Deletes a connector.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    actionId := "c55b6eb0-6bad-11eb-9f3b-611eebc6c3ad" // string | An identifier for the action.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.ConnectorsApi.LegacyDeleteConnector(context.Background(), actionId, spaceId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.LegacyDeleteConnector``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**actionId** | **string** | An identifier for the action. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyDeleteConnectorRequest struct via the builder pattern


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


## LegacyGetConnector

> ActionResponseProperties LegacyGetConnector(ctx, actionId, spaceId).Execute()

Retrieves a connector by ID.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    actionId := "c55b6eb0-6bad-11eb-9f3b-611eebc6c3ad" // string | An identifier for the action.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.LegacyGetConnector(context.Background(), actionId, spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.LegacyGetConnector``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyGetConnector`: ActionResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.LegacyGetConnector`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**actionId** | **string** | An identifier for the action. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyGetConnectorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**ActionResponseProperties**](ActionResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyGetConnectorTypes

> []LegacyGetConnectorTypesResponseBodyPropertiesInner LegacyGetConnectorTypes(ctx, spaceId).Execute()

Retrieves a list of all connector types.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.LegacyGetConnectorTypes(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.LegacyGetConnectorTypes``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyGetConnectorTypes`: []LegacyGetConnectorTypesResponseBodyPropertiesInner
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.LegacyGetConnectorTypes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyGetConnectorTypesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]LegacyGetConnectorTypesResponseBodyPropertiesInner**](LegacyGetConnectorTypesResponseBodyPropertiesInner.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyGetConnectors

> []ActionResponseProperties LegacyGetConnectors(ctx, spaceId).Execute()

Retrieves all connectors.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.LegacyGetConnectors(context.Background(), spaceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.LegacyGetConnectors``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyGetConnectors`: []ActionResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.LegacyGetConnectors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyGetConnectorsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]ActionResponseProperties**](ActionResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyRunConnector

> LegacyRunConnector200Response LegacyRunConnector(ctx, actionId, spaceId).KbnXsrf(kbnXsrf).LegacyRunConnectorRequestBodyProperties(legacyRunConnectorRequestBodyProperties).Execute()

Runs a connector.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    actionId := "c55b6eb0-6bad-11eb-9f3b-611eebc6c3ad" // string | An identifier for the action.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    legacyRunConnectorRequestBodyProperties := *openapiclient.NewLegacyRunConnectorRequestBodyProperties(map[string]interface{}(123)) // LegacyRunConnectorRequestBodyProperties | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.LegacyRunConnector(context.Background(), actionId, spaceId).KbnXsrf(kbnXsrf).LegacyRunConnectorRequestBodyProperties(legacyRunConnectorRequestBodyProperties).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.LegacyRunConnector``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyRunConnector`: LegacyRunConnector200Response
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.LegacyRunConnector`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**actionId** | **string** | An identifier for the action. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyRunConnectorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 


 **legacyRunConnectorRequestBodyProperties** | [**LegacyRunConnectorRequestBodyProperties**](LegacyRunConnectorRequestBodyProperties.md) |  | 

### Return type

[**LegacyRunConnector200Response**](LegacyRunConnector200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LegacyUpdateConnector

> ActionResponseProperties LegacyUpdateConnector(ctx, actionId, spaceId).KbnXsrf(kbnXsrf).LegacyUpdateConnectorRequestBodyProperties(legacyUpdateConnectorRequestBodyProperties).Execute()

Updates the attributes for a connector.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    actionId := "c55b6eb0-6bad-11eb-9f3b-611eebc6c3ad" // string | An identifier for the action.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    legacyUpdateConnectorRequestBodyProperties := *openapiclient.NewLegacyUpdateConnectorRequestBodyProperties() // LegacyUpdateConnectorRequestBodyProperties | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.LegacyUpdateConnector(context.Background(), actionId, spaceId).KbnXsrf(kbnXsrf).LegacyUpdateConnectorRequestBodyProperties(legacyUpdateConnectorRequestBodyProperties).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.LegacyUpdateConnector``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `LegacyUpdateConnector`: ActionResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.LegacyUpdateConnector`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**actionId** | **string** | An identifier for the action. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiLegacyUpdateConnectorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 


 **legacyUpdateConnectorRequestBodyProperties** | [**LegacyUpdateConnectorRequestBodyProperties**](LegacyUpdateConnectorRequestBodyProperties.md) |  | 

### Return type

[**ActionResponseProperties**](ActionResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RunConnector

> RunConnector200Response RunConnector(ctx, connectorId, spaceId).KbnXsrf(kbnXsrf).RunConnectorRequestBodyProperties(runConnectorRequestBodyProperties).Execute()

Runs a connector.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    connectorId := "df770e30-8b8b-11ed-a780-3b746c987a81" // string | An identifier for the connector.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    runConnectorRequestBodyProperties := *openapiclient.NewRunConnectorRequestBodyProperties(openapiclient.Run_connector_request_body_properties_params{RunConnectorParamsDocuments: openapiclient.NewRunConnectorParamsDocuments([]map[string]interface{}{map[string]interface{}{"key": interface{}(123)}})}) // RunConnectorRequestBodyProperties | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.RunConnector(context.Background(), connectorId, spaceId).KbnXsrf(kbnXsrf).RunConnectorRequestBodyProperties(runConnectorRequestBodyProperties).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.RunConnector``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `RunConnector`: RunConnector200Response
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.RunConnector`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**connectorId** | **string** | An identifier for the connector. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiRunConnectorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 


 **runConnectorRequestBodyProperties** | [**RunConnectorRequestBodyProperties**](RunConnectorRequestBodyProperties.md) |  | 

### Return type

[**RunConnector200Response**](RunConnector200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateConnector

> ConnectorResponseProperties UpdateConnector(ctx, connectorId, spaceId).KbnXsrf(kbnXsrf).UpdateConnectorRequestBodyProperties(updateConnectorRequestBodyProperties).Execute()

Updates the attributes for a connector.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/kibanaactions"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    connectorId := "df770e30-8b8b-11ed-a780-3b746c987a81" // string | An identifier for the connector.
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    updateConnectorRequestBodyProperties := openapiclient.Update_connector_request_body_properties{UpdateConnectorRequestCasesWebhook: openapiclient.NewUpdateConnectorRequestCasesWebhook(*openapiclient.NewConfigPropertiesCasesWebhook("{"fields":{"summary":{"[object Object]":null},"description":{"[object Object]":null},"labels":{"[object Object]":null}}}", "CreateIncidentResponseKey_example", "CreateIncidentUrl_example", "GetIncidentResponseExternalTitleKey_example", "https://testing-jira.atlassian.net/rest/api/2/issue/{{{external.system.id}}}", "{"fields":{"summary":{"[object Object]":null},"description":{"[object Object]":null},"labels":{"[object Object]":null}}}", "https://testing-jira.atlassian.net/rest/api/2/issue/{{{external.system.ID}}}", "https://testing-jira.atlassian.net/browse/{{{external.system.title}}}"), "my-connector")} // UpdateConnectorRequestBodyProperties | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ConnectorsApi.UpdateConnector(context.Background(), connectorId, spaceId).KbnXsrf(kbnXsrf).UpdateConnectorRequestBodyProperties(updateConnectorRequestBodyProperties).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ConnectorsApi.UpdateConnector``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateConnector`: ConnectorResponseProperties
    fmt.Fprintf(os.Stdout, "Response from `ConnectorsApi.UpdateConnector`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**connectorId** | **string** | An identifier for the connector. | 
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateConnectorRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 


 **updateConnectorRequestBodyProperties** | [**UpdateConnectorRequestBodyProperties**](UpdateConnectorRequestBodyProperties.md) |  | 

### Return type

[**ConnectorResponseProperties**](ConnectorResponseProperties.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

