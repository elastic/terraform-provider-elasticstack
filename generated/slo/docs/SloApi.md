# \SloAPI

All URIs are relative to *http://localhost:5601*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateSloOp**](SloAPI.md#CreateSloOp) | **Post** /s/{spaceId}/api/observability/slos | Creates an SLO.
[**DeleteSloOp**](SloAPI.md#DeleteSloOp) | **Delete** /s/{spaceId}/api/observability/slos/{sloId} | Deletes an SLO
[**DisableSloOp**](SloAPI.md#DisableSloOp) | **Post** /s/{spaceId}/api/observability/slos/{sloId}/disable | Disables an SLO
[**EnableSloOp**](SloAPI.md#EnableSloOp) | **Post** /s/{spaceId}/api/observability/slos/{sloId}/enable | Enables an SLO
[**FindSlosOp**](SloAPI.md#FindSlosOp) | **Get** /s/{spaceId}/api/observability/slos | Retrieves a paginated list of SLOs
[**GetSloOp**](SloAPI.md#GetSloOp) | **Get** /s/{spaceId}/api/observability/slos/{sloId} | Retrieves a SLO
[**HistoricalSummaryOp**](SloAPI.md#HistoricalSummaryOp) | **Post** /s/{spaceId}/internal/observability/slos/_historical_summary | Retrieves the historical summary for a list of SLOs
[**UpdateSloOp**](SloAPI.md#UpdateSloOp) | **Put** /s/{spaceId}/api/observability/slos/{sloId} | Updates an SLO



## CreateSloOp

> CreateSloResponse CreateSloOp(ctx, spaceId).KbnXsrf(kbnXsrf).CreateSloRequest(createSloRequest).Execute()

Creates an SLO.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/slo"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    createSloRequest := *openapiclient.NewCreateSloRequest("Name_example", "Description_example", openapiclient.create_slo_request_indicator{IndicatorPropertiesApmAvailability: openapiclient.NewIndicatorPropertiesApmAvailability(*openapiclient.NewIndicatorPropertiesApmAvailabilityParams("o11y-app", "production", "request", "GET /my/api", "metrics-apm*,apm*"), "sli.apm.transactionDuration")}, *openapiclient.NewTimeWindow("30d", "rolling"), openapiclient.budgeting_method("occurrences"), *openapiclient.NewObjective(float64(0.99))) // CreateSloRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.CreateSloOp(context.Background(), spaceId).KbnXsrf(kbnXsrf).CreateSloRequest(createSloRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.CreateSloOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateSloOp`: CreateSloResponse
    fmt.Fprintf(os.Stdout, "Response from `SloAPI.CreateSloOp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateSloOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **createSloRequest** | [**CreateSloRequest**](CreateSloRequest.md) |  | 

### Return type

[**CreateSloResponse**](CreateSloResponse.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteSloOp

> DeleteSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

Deletes an SLO



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/slo"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    sloId := "9c235211-6834-11ea-a78c-6feb38a34414" // string | An identifier for the slo.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.SloAPI.DeleteSloOp(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.DeleteSloOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**sloId** | **string** | An identifier for the slo. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteSloOpRequest struct via the builder pattern


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


## DisableSloOp

> DisableSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

Disables an SLO



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/slo"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    sloId := "9c235211-6834-11ea-a78c-6feb38a34414" // string | An identifier for the slo.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.SloAPI.DisableSloOp(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.DisableSloOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**sloId** | **string** | An identifier for the slo. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDisableSloOpRequest struct via the builder pattern


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


## EnableSloOp

> EnableSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

Enables an SLO



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/slo"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    sloId := "9c235211-6834-11ea-a78c-6feb38a34414" // string | An identifier for the slo.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.SloAPI.EnableSloOp(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.EnableSloOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**sloId** | **string** | An identifier for the slo. | 

### Other Parameters

Other parameters are passed through a pointer to a apiEnableSloOpRequest struct via the builder pattern


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


## FindSlosOp

> FindSloResponse FindSlosOp(ctx, spaceId).KbnXsrf(kbnXsrf).Name(name).IndicatorTypes(indicatorTypes).Page(page).PerPage(perPage).SortBy(sortBy).SortDirection(sortDirection).Execute()

Retrieves a paginated list of SLOs



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/slo"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    name := "awesome-service" // string | Filter by name (optional)
    indicatorTypes := []string{"Inner_example"} // []string | Filter by indicator type (optional)
    page := int32(1) // int32 | The page number to return (optional) (default to 1)
    perPage := int32(20) // int32 | The number of SLOs to return per page (optional) (default to 25)
    sortBy := "creationTime" // string | Sort by field (optional) (default to "creationTime")
    sortDirection := "asc" // string | Sort order (optional) (default to "asc")

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.FindSlosOp(context.Background(), spaceId).KbnXsrf(kbnXsrf).Name(name).IndicatorTypes(indicatorTypes).Page(page).PerPage(perPage).SortBy(sortBy).SortDirection(sortDirection).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.FindSlosOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `FindSlosOp`: FindSloResponse
    fmt.Fprintf(os.Stdout, "Response from `SloAPI.FindSlosOp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiFindSlosOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **name** | **string** | Filter by name | 
 **indicatorTypes** | **[]string** | Filter by indicator type | 
 **page** | **int32** | The page number to return | [default to 1]
 **perPage** | **int32** | The number of SLOs to return per page | [default to 25]
 **sortBy** | **string** | Sort by field | [default to &quot;creationTime&quot;]
 **sortDirection** | **string** | Sort order | [default to &quot;asc&quot;]

### Return type

[**FindSloResponse**](FindSloResponse.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSloOp

> SloResponse GetSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

Retrieves a SLO



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/slo"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    sloId := "9c235211-6834-11ea-a78c-6feb38a34414" // string | An identifier for the slo.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.GetSloOp(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.GetSloOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSloOp`: SloResponse
    fmt.Fprintf(os.Stdout, "Response from `SloAPI.GetSloOp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**sloId** | **string** | An identifier for the slo. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSloOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

[**SloResponse**](SloResponse.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## HistoricalSummaryOp

> map[string][]HistoricalSummaryResponseInner HistoricalSummaryOp(ctx, spaceId).KbnXsrf(kbnXsrf).HistoricalSummaryRequest(historicalSummaryRequest).Execute()

Retrieves the historical summary for a list of SLOs



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/slo"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    historicalSummaryRequest := *openapiclient.NewHistoricalSummaryRequest([]string{"8853df00-ae2e-11ed-90af-09bb6422b258"}) // HistoricalSummaryRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.HistoricalSummaryOp(context.Background(), spaceId).KbnXsrf(kbnXsrf).HistoricalSummaryRequest(historicalSummaryRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.HistoricalSummaryOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `HistoricalSummaryOp`: map[string][]HistoricalSummaryResponseInner
    fmt.Fprintf(os.Stdout, "Response from `SloAPI.HistoricalSummaryOp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiHistoricalSummaryOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **historicalSummaryRequest** | [**HistoricalSummaryRequest**](HistoricalSummaryRequest.md) |  | 

### Return type

[**map[string][]HistoricalSummaryResponseInner**](array.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateSloOp

> SloResponse UpdateSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).UpdateSloRequest(updateSloRequest).Execute()

Updates an SLO



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/slo"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    spaceId := "default" // string | An identifier for the space. If `/s/` and the identifier are omitted from the path, the default space is used.
    sloId := "9c235211-6834-11ea-a78c-6feb38a34414" // string | An identifier for the slo.
    updateSloRequest := *openapiclient.NewUpdateSloRequest() // UpdateSloRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.UpdateSloOp(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).UpdateSloRequest(updateSloRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.UpdateSloOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateSloOp`: SloResponse
    fmt.Fprintf(os.Stdout, "Response from `SloAPI.UpdateSloOp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**sloId** | **string** | An identifier for the slo. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSloOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 


 **updateSloRequest** | [**UpdateSloRequest**](UpdateSloRequest.md) |  | 

### Return type

[**SloResponse**](SloResponse.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

