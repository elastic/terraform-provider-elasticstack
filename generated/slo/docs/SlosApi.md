# \SlosApi

All URIs are relative to *http://localhost:5601*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateSlo**](SlosApi.md#CreateSlo) | **Post** /s/{spaceId}/api/observability/slos | Creates an SLO.
[**DeleteSlo**](SlosApi.md#DeleteSlo) | **Delete** /s/{spaceId}/api/observability/slos/{sloId} | Deletes an SLO
[**DisableSlo**](SlosApi.md#DisableSlo) | **Post** /s/{spaceId}/api/observability/slos/{sloId}/disable | Disables an SLO
[**EnableSlo**](SlosApi.md#EnableSlo) | **Post** /s/{spaceId}/api/observability/slos/{sloId}/enable | Enables an SLO
[**FindSlos**](SlosApi.md#FindSlos) | **Get** /s/{spaceId}/api/observability/slos | Retrieves a paginated list of SLOs
[**GetSlo**](SlosApi.md#GetSlo) | **Get** /s/{spaceId}/api/observability/slos/{sloId} | Retrieves a SLO
[**HistoricalSummary**](SlosApi.md#HistoricalSummary) | **Post** /s/{spaceId}/internal/observability/slos/_historical_summary | Retrieves the historical summary for a list of SLOs
[**UpdateSlo**](SlosApi.md#UpdateSlo) | **Put** /s/{spaceId}/api/observability/slos/{sloId} | Updates an SLO



## CreateSlo

> CreateSloResponse CreateSlo(ctx, spaceId).KbnXsrf(kbnXsrf).CreateSloRequest(createSloRequest).Execute()

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
    createSloRequest := *openapiclient.NewCreateSloRequest("Name_example", "Description_example", openapiclient.slo_response_indicator{IndicatorPropertiesApmAvailability: openapiclient.NewIndicatorPropertiesApmAvailability(*openapiclient.NewIndicatorPropertiesApmAvailabilityParams("o11y-app", "production", "request", "GET /my/api", "metrics-apm*,apm*"), "sli.apm.transactionDuration")}, openapiclient.slo_response_timeWindow{TimeWindowCalendarAligned: openapiclient.NewTimeWindowCalendarAligned("1M", *openapiclient.NewTimeWindowCalendarAlignedCalendar())}, openapiclient.budgeting_method("occurrences"), *openapiclient.NewObjective(float32(0.99))) // CreateSloRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SlosApi.CreateSlo(context.Background(), spaceId).KbnXsrf(kbnXsrf).CreateSloRequest(createSloRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SlosApi.CreateSlo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateSlo`: CreateSloResponse
    fmt.Fprintf(os.Stdout, "Response from `SlosApi.CreateSlo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateSloRequest struct via the builder pattern


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


## DeleteSlo

> DeleteSlo(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

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
    r, err := apiClient.SlosApi.DeleteSlo(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SlosApi.DeleteSlo``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteSloRequest struct via the builder pattern


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


## DisableSlo

> DisableSlo(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

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
    r, err := apiClient.SlosApi.DisableSlo(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SlosApi.DisableSlo``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDisableSloRequest struct via the builder pattern


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


## EnableSlo

> EnableSlo(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

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
    r, err := apiClient.SlosApi.EnableSlo(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SlosApi.EnableSlo``: %v\n", err)
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

Other parameters are passed through a pointer to a apiEnableSloRequest struct via the builder pattern


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


## FindSlos

> FindSloResponse FindSlos(ctx, spaceId).KbnXsrf(kbnXsrf).Name(name).IndicatorTypes(indicatorTypes).Page(page).PerPage(perPage).SortBy(sortBy).SortDirection(sortDirection).Execute()

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
    sortBy := "name" // string | Sort by field (optional) (default to "name")
    sortDirection := "asc" // string | Sort order (optional) (default to "asc")

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SlosApi.FindSlos(context.Background(), spaceId).KbnXsrf(kbnXsrf).Name(name).IndicatorTypes(indicatorTypes).Page(page).PerPage(perPage).SortBy(sortBy).SortDirection(sortDirection).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SlosApi.FindSlos``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `FindSlos`: FindSloResponse
    fmt.Fprintf(os.Stdout, "Response from `SlosApi.FindSlos`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiFindSlosRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **name** | **string** | Filter by name | 
 **indicatorTypes** | **[]string** | Filter by indicator type | 
 **page** | **int32** | The page number to return | [default to 1]
 **perPage** | **int32** | The number of SLOs to return per page | [default to 25]
 **sortBy** | **string** | Sort by field | [default to &quot;name&quot;]
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


## GetSlo

> SloResponse GetSlo(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

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
    resp, r, err := apiClient.SlosApi.GetSlo(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SlosApi.GetSlo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSlo`: SloResponse
    fmt.Fprintf(os.Stdout, "Response from `SlosApi.GetSlo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**sloId** | **string** | An identifier for the slo. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSloRequest struct via the builder pattern


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


## HistoricalSummary

> map[string][]HistoricalSummaryResponseInner HistoricalSummary(ctx, spaceId).KbnXsrf(kbnXsrf).HistoricalSummaryRequest(historicalSummaryRequest).Execute()

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
    resp, r, err := apiClient.SlosApi.HistoricalSummary(context.Background(), spaceId).KbnXsrf(kbnXsrf).HistoricalSummaryRequest(historicalSummaryRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SlosApi.HistoricalSummary``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `HistoricalSummary`: map[string][]HistoricalSummaryResponseInner
    fmt.Fprintf(os.Stdout, "Response from `SlosApi.HistoricalSummary`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiHistoricalSummaryRequest struct via the builder pattern


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


## UpdateSlo

> SloResponse UpdateSlo(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).UpdateSloRequest(updateSloRequest).Execute()

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
    resp, r, err := apiClient.SlosApi.UpdateSlo(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).UpdateSloRequest(updateSloRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SlosApi.UpdateSlo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateSlo`: SloResponse
    fmt.Fprintf(os.Stdout, "Response from `SlosApi.UpdateSlo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**sloId** | **string** | An identifier for the slo. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSloRequest struct via the builder pattern


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

