# \SloAPI

All URIs are relative to *https://localhost:5601*

Method | HTTP request | Description
------------- | ------------- | -------------
[**BulkDeleteOp**](SloAPI.md#BulkDeleteOp) | **Post** /s/{spaceId}/api/observability/slos/_bulk_delete | Bulk delete SLO definitions and their associated summary and rollup data.
[**BulkDeleteStatusOp**](SloAPI.md#BulkDeleteStatusOp) | **Get** /s/{spaceId}/api/observability/slos/_bulk_delete/{taskId} | Retrieve the status of the bulk deletion
[**CreateSloOp**](SloAPI.md#CreateSloOp) | **Post** /s/{spaceId}/api/observability/slos | Create an SLO
[**DeleteRollupDataOp**](SloAPI.md#DeleteRollupDataOp) | **Post** /s/{spaceId}/api/observability/slos/_bulk_purge_rollup | Batch delete rollup and summary data
[**DeleteSloInstancesOp**](SloAPI.md#DeleteSloInstancesOp) | **Post** /s/{spaceId}/api/observability/slos/_delete_instances | Batch delete rollup and summary data
[**DeleteSloOp**](SloAPI.md#DeleteSloOp) | **Delete** /s/{spaceId}/api/observability/slos/{sloId} | Delete an SLO
[**DisableSloOp**](SloAPI.md#DisableSloOp) | **Post** /s/{spaceId}/api/observability/slos/{sloId}/disable | Disable an SLO
[**EnableSloOp**](SloAPI.md#EnableSloOp) | **Post** /s/{spaceId}/api/observability/slos/{sloId}/enable | Enable an SLO
[**FindSlosOp**](SloAPI.md#FindSlosOp) | **Get** /s/{spaceId}/api/observability/slos | Get a paginated list of SLOs
[**GetDefinitionsOp**](SloAPI.md#GetDefinitionsOp) | **Get** /s/{spaceId}/internal/observability/slos/_definitions | Get the SLO definitions
[**GetSloOp**](SloAPI.md#GetSloOp) | **Get** /s/{spaceId}/api/observability/slos/{sloId} | Get an SLO
[**ResetSloOp**](SloAPI.md#ResetSloOp) | **Post** /s/{spaceId}/api/observability/slos/{sloId}/_reset | Reset an SLO
[**UpdateSloOp**](SloAPI.md#UpdateSloOp) | **Put** /s/{spaceId}/api/observability/slos/{sloId} | Update an SLO



## BulkDeleteOp

> BulkDeleteResponse BulkDeleteOp(ctx, spaceId).KbnXsrf(kbnXsrf).BulkDeleteRequest(bulkDeleteRequest).Execute()

Bulk delete SLO definitions and their associated summary and rollup data.



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
    bulkDeleteRequest := *openapiclient.NewBulkDeleteRequest([]string{"8853df00-ae2e-11ed-90af-09bb6422b258"}) // BulkDeleteRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.BulkDeleteOp(context.Background(), spaceId).KbnXsrf(kbnXsrf).BulkDeleteRequest(bulkDeleteRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.BulkDeleteOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BulkDeleteOp`: BulkDeleteResponse
    fmt.Fprintf(os.Stdout, "Response from `SloAPI.BulkDeleteOp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiBulkDeleteOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **bulkDeleteRequest** | [**BulkDeleteRequest**](BulkDeleteRequest.md) |  | 

### Return type

[**BulkDeleteResponse**](BulkDeleteResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BulkDeleteStatusOp

> BulkDeleteStatusResponse BulkDeleteStatusOp(ctx, spaceId, taskId).KbnXsrf(kbnXsrf).Execute()

Retrieve the status of the bulk deletion



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
    taskId := "8853df00-ae2e-11ed-90af-09bb6422b258" // string | The task id of the bulk delete operation

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.BulkDeleteStatusOp(context.Background(), spaceId, taskId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.BulkDeleteStatusOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `BulkDeleteStatusOp`: BulkDeleteStatusResponse
    fmt.Fprintf(os.Stdout, "Response from `SloAPI.BulkDeleteStatusOp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**taskId** | **string** | The task id of the bulk delete operation | 

### Other Parameters

Other parameters are passed through a pointer to a apiBulkDeleteStatusOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

[**BulkDeleteStatusResponse**](BulkDeleteStatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateSloOp

> CreateSloResponse CreateSloOp(ctx, spaceId).KbnXsrf(kbnXsrf).CreateSloRequest(createSloRequest).Execute()

Create an SLO



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

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRollupDataOp

> BulkPurgeRollupResponse DeleteRollupDataOp(ctx, spaceId).KbnXsrf(kbnXsrf).BulkPurgeRollupRequest(bulkPurgeRollupRequest).Execute()

Batch delete rollup and summary data



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
    bulkPurgeRollupRequest := *openapiclient.NewBulkPurgeRollupRequest([]string{"8853df00-ae2e-11ed-90af-09bb6422b258"}, openapiclient.bulk_purge_rollup_request_purgePolicy{BulkPurgeRollupRequestPurgePolicyOneOf: openapiclient.NewBulkPurgeRollupRequestPurgePolicyOneOf()}) // BulkPurgeRollupRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.DeleteRollupDataOp(context.Background(), spaceId).KbnXsrf(kbnXsrf).BulkPurgeRollupRequest(bulkPurgeRollupRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.DeleteRollupDataOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteRollupDataOp`: BulkPurgeRollupResponse
    fmt.Fprintf(os.Stdout, "Response from `SloAPI.DeleteRollupDataOp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRollupDataOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **bulkPurgeRollupRequest** | [**BulkPurgeRollupRequest**](BulkPurgeRollupRequest.md) |  | 

### Return type

[**BulkPurgeRollupResponse**](BulkPurgeRollupResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteSloInstancesOp

> DeleteSloInstancesOp(ctx, spaceId).KbnXsrf(kbnXsrf).DeleteSloInstancesRequest(deleteSloInstancesRequest).Execute()

Batch delete rollup and summary data



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
    deleteSloInstancesRequest := *openapiclient.NewDeleteSloInstancesRequest([]openapiclient.DeleteSloInstancesRequestListInner{*openapiclient.NewDeleteSloInstancesRequestListInner("8853df00-ae2e-11ed-90af-09bb6422b258", "8853df00-ae2e-11ed-90af-09bb6422b258")}) // DeleteSloInstancesRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.SloAPI.DeleteSloInstancesOp(context.Background(), spaceId).KbnXsrf(kbnXsrf).DeleteSloInstancesRequest(deleteSloInstancesRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.DeleteSloInstancesOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteSloInstancesOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **deleteSloInstancesRequest** | [**DeleteSloInstancesRequest**](DeleteSloInstancesRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteSloOp

> DeleteSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

Delete an SLO



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

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DisableSloOp

> DisableSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

Disable an SLO



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

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## EnableSloOp

> EnableSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

Enable an SLO



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

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## FindSlosOp

> FindSloResponse FindSlosOp(ctx, spaceId).KbnXsrf(kbnXsrf).KqlQuery(kqlQuery).Size(size).SearchAfter(searchAfter).Page(page).PerPage(perPage).SortBy(sortBy).SortDirection(sortDirection).HideStale(hideStale).Execute()

Get a paginated list of SLOs



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
    kqlQuery := "slo.name:latency* and slo.tags : "prod"" // string | A valid kql query to filter the SLO with (optional)
    size := int32(1) // int32 | The page size to use for cursor-based pagination, must be greater or equal than 1 (optional) (default to 1)
    searchAfter := []string{"Inner_example"} // []string | The cursor to use for fetching the results from, when using a cursor-base pagination. (optional)
    page := int32(1) // int32 | The page to use for pagination, must be greater or equal than 1 (optional) (default to 1)
    perPage := int32(25) // int32 | Number of SLOs returned by page (optional) (default to 25)
    sortBy := "status" // string | Sort by field (optional) (default to "status")
    sortDirection := "asc" // string | Sort order (optional) (default to "asc")
    hideStale := true // bool | Hide stale SLOs from the list as defined by stale SLO threshold in SLO settings (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.FindSlosOp(context.Background(), spaceId).KbnXsrf(kbnXsrf).KqlQuery(kqlQuery).Size(size).SearchAfter(searchAfter).Page(page).PerPage(perPage).SortBy(sortBy).SortDirection(sortDirection).HideStale(hideStale).Execute()
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

 **kqlQuery** | **string** | A valid kql query to filter the SLO with | 
 **size** | **int32** | The page size to use for cursor-based pagination, must be greater or equal than 1 | [default to 1]
 **searchAfter** | **[]string** | The cursor to use for fetching the results from, when using a cursor-base pagination. | 
 **page** | **int32** | The page to use for pagination, must be greater or equal than 1 | [default to 1]
 **perPage** | **int32** | Number of SLOs returned by page | [default to 25]
 **sortBy** | **string** | Sort by field | [default to &quot;status&quot;]
 **sortDirection** | **string** | Sort order | [default to &quot;asc&quot;]
 **hideStale** | **bool** | Hide stale SLOs from the list as defined by stale SLO threshold in SLO settings | 

### Return type

[**FindSloResponse**](FindSloResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDefinitionsOp

> FindSloDefinitionsResponse GetDefinitionsOp(ctx, spaceId).KbnXsrf(kbnXsrf).IncludeOutdatedOnly(includeOutdatedOnly).Tags(tags).Search(search).Page(page).PerPage(perPage).Execute()

Get the SLO definitions



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
    includeOutdatedOnly := true // bool | Indicates if the API returns only outdated SLO or all SLO definitions (optional)
    tags := "tags_example" // string | Filters the SLOs by tag (optional)
    search := "my service availability" // string | Filters the SLOs by name (optional)
    page := float64(1) // float64 | The page to use for pagination, must be greater or equal than 1 (optional)
    perPage := int32(100) // int32 | Number of SLOs returned by page (optional) (default to 100)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.GetDefinitionsOp(context.Background(), spaceId).KbnXsrf(kbnXsrf).IncludeOutdatedOnly(includeOutdatedOnly).Tags(tags).Search(search).Page(page).PerPage(perPage).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.GetDefinitionsOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDefinitionsOp`: FindSloDefinitionsResponse
    fmt.Fprintf(os.Stdout, "Response from `SloAPI.GetDefinitionsOp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDefinitionsOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **includeOutdatedOnly** | **bool** | Indicates if the API returns only outdated SLO or all SLO definitions | 
 **tags** | **string** | Filters the SLOs by tag | 
 **search** | **string** | Filters the SLOs by name | 
 **page** | **float64** | The page to use for pagination, must be greater or equal than 1 | 
 **perPage** | **int32** | Number of SLOs returned by page | [default to 100]

### Return type

[**FindSloDefinitionsResponse**](FindSloDefinitionsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSloOp

> SloWithSummaryResponse GetSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).InstanceId(instanceId).Execute()

Get an SLO



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
    instanceId := "host-abcde" // string | the specific instanceId used by the summary calculation (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SloAPI.GetSloOp(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).InstanceId(instanceId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.GetSloOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSloOp`: SloWithSummaryResponse
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


 **instanceId** | **string** | the specific instanceId used by the summary calculation | 

### Return type

[**SloWithSummaryResponse**](SloWithSummaryResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ResetSloOp

> SloDefinitionResponse ResetSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).Execute()

Reset an SLO



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
    resp, r, err := apiClient.SloAPI.ResetSloOp(context.Background(), spaceId, sloId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SloAPI.ResetSloOp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ResetSloOp`: SloDefinitionResponse
    fmt.Fprintf(os.Stdout, "Response from `SloAPI.ResetSloOp`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**spaceId** | **string** | An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
**sloId** | **string** | An identifier for the slo. | 

### Other Parameters

Other parameters are passed through a pointer to a apiResetSloOpRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 



### Return type

[**SloDefinitionResponse**](SloDefinitionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateSloOp

> SloDefinitionResponse UpdateSloOp(ctx, spaceId, sloId).KbnXsrf(kbnXsrf).UpdateSloRequest(updateSloRequest).Execute()

Update an SLO



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
    // response from `UpdateSloOp`: SloDefinitionResponse
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

[**SloDefinitionResponse**](SloDefinitionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

