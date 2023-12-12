# \DataViewsAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDataView**](DataViewsAPI.md#CreateDataView) | **Post** /api/data_views/data_view | Creates a data view.
[**CreateRuntimeField**](DataViewsAPI.md#CreateRuntimeField) | **Post** /api/data_views/data_view/{viewId}/runtime_field | Creates a runtime field.
[**CreateUpdateRuntimeField**](DataViewsAPI.md#CreateUpdateRuntimeField) | **Put** /api/data_views/data_view/{viewId}/runtime_field | Create or update an existing runtime field.
[**DeleteDataView**](DataViewsAPI.md#DeleteDataView) | **Delete** /api/data_views/data_view/{viewId} | Deletes a data view.
[**DeleteRuntimeField**](DataViewsAPI.md#DeleteRuntimeField) | **Delete** /api/data_views/data_view/{viewId}/runtime_field/{fieldName} | Delete a runtime field from a data view.
[**GetAllDataViews**](DataViewsAPI.md#GetAllDataViews) | **Get** /api/data_views | Retrieves a list of all data views.
[**GetDataView**](DataViewsAPI.md#GetDataView) | **Get** /api/data_views/data_view/{viewId} | Retrieves a single data view by identifier.
[**GetDefaultDataView**](DataViewsAPI.md#GetDefaultDataView) | **Get** /api/data_views/default | Retrieves the default data view identifier.
[**GetRuntimeField**](DataViewsAPI.md#GetRuntimeField) | **Get** /api/data_views/data_view/{viewId}/runtime_field/{fieldName} | Retrieves a runtime field.
[**SetDefaultDatailView**](DataViewsAPI.md#SetDefaultDatailView) | **Post** /api/data_views/default | Sets the default data view identifier.
[**UpdateDataView**](DataViewsAPI.md#UpdateDataView) | **Post** /api/data_views/data_view/{viewId} | Updates a data view.
[**UpdateFieldsMetadata**](DataViewsAPI.md#UpdateFieldsMetadata) | **Post** /api/data_views/data_view/{viewId}/fields | Update fields presentation metadata such as count, customLabel and format.
[**UpdateRuntimeField**](DataViewsAPI.md#UpdateRuntimeField) | **Post** /api/data_views/data_view/{viewId}/runtime_field/{fieldName} | Update an existing runtime field.



## CreateDataView

> DataViewResponseObject CreateDataView(ctx).KbnXsrf(kbnXsrf).CreateDataViewRequestObject(createDataViewRequestObject).Execute()

Creates a data view.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    createDataViewRequestObject := *openapiclient.NewCreateDataViewRequestObject(*openapiclient.NewCreateDataViewRequestObjectDataView("Title_example")) // CreateDataViewRequestObject | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DataViewsAPI.CreateDataView(context.Background()).KbnXsrf(kbnXsrf).CreateDataViewRequestObject(createDataViewRequestObject).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.CreateDataView``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDataView`: DataViewResponseObject
    fmt.Fprintf(os.Stdout, "Response from `DataViewsAPI.CreateDataView`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateDataViewRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 
 **createDataViewRequestObject** | [**CreateDataViewRequestObject**](CreateDataViewRequestObject.md) |  | 

### Return type

[**DataViewResponseObject**](DataViewResponseObject.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateRuntimeField

> CreateRuntimeField(ctx, viewId).KbnXsrf(kbnXsrf).CreateUpdateRuntimeFieldRequest(createUpdateRuntimeFieldRequest).Execute()

Creates a runtime field.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    viewId := "ff959d40-b880-11e8-a6d9-e546fe2bba5f" // string | An identifier for the data view.
    createUpdateRuntimeFieldRequest := *openapiclient.NewCreateUpdateRuntimeFieldRequest("Name_example", map[string]interface{}(123)) // CreateUpdateRuntimeFieldRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.DataViewsAPI.CreateRuntimeField(context.Background(), viewId).KbnXsrf(kbnXsrf).CreateUpdateRuntimeFieldRequest(createUpdateRuntimeFieldRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.CreateRuntimeField``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**viewId** | **string** | An identifier for the data view. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateRuntimeFieldRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **createUpdateRuntimeFieldRequest** | [**CreateUpdateRuntimeFieldRequest**](CreateUpdateRuntimeFieldRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateUpdateRuntimeField

> CreateUpdateRuntimeField200Response CreateUpdateRuntimeField(ctx, viewId).KbnXsrf(kbnXsrf).CreateUpdateRuntimeFieldRequest(createUpdateRuntimeFieldRequest).Execute()

Create or update an existing runtime field.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    viewId := "viewId_example" // string | The ID of the data view fields you want to update. 
    createUpdateRuntimeFieldRequest := *openapiclient.NewCreateUpdateRuntimeFieldRequest("Name_example", map[string]interface{}(123)) // CreateUpdateRuntimeFieldRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DataViewsAPI.CreateUpdateRuntimeField(context.Background(), viewId).KbnXsrf(kbnXsrf).CreateUpdateRuntimeFieldRequest(createUpdateRuntimeFieldRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.CreateUpdateRuntimeField``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateUpdateRuntimeField`: CreateUpdateRuntimeField200Response
    fmt.Fprintf(os.Stdout, "Response from `DataViewsAPI.CreateUpdateRuntimeField`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**viewId** | **string** | The ID of the data view fields you want to update.  | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateUpdateRuntimeFieldRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **createUpdateRuntimeFieldRequest** | [**CreateUpdateRuntimeFieldRequest**](CreateUpdateRuntimeFieldRequest.md) |  | 

### Return type

[**CreateUpdateRuntimeField200Response**](CreateUpdateRuntimeField200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDataView

> DeleteDataView(ctx, viewId).KbnXsrf(kbnXsrf).Execute()

Deletes a data view.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    viewId := "ff959d40-b880-11e8-a6d9-e546fe2bba5f" // string | An identifier for the data view.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.DataViewsAPI.DeleteDataView(context.Background(), viewId).KbnXsrf(kbnXsrf).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.DeleteDataView``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**viewId** | **string** | An identifier for the data view. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDataViewRequest struct via the builder pattern


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


## DeleteRuntimeField

> DeleteRuntimeField(ctx, fieldName, viewId).Execute()

Delete a runtime field from a data view.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    fieldName := "hour_of_day" // string | The name of the runtime field.
    viewId := "ff959d40-b880-11e8-a6d9-e546fe2bba5f" // string | An identifier for the data view.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.DataViewsAPI.DeleteRuntimeField(context.Background(), fieldName, viewId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.DeleteRuntimeField``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fieldName** | **string** | The name of the runtime field. | 
**viewId** | **string** | An identifier for the data view. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRuntimeFieldRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



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


## GetAllDataViews

> GetAllDataViews200Response GetAllDataViews(ctx).Execute()

Retrieves a list of all data views.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DataViewsAPI.GetAllDataViews(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.GetAllDataViews``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAllDataViews`: GetAllDataViews200Response
    fmt.Fprintf(os.Stdout, "Response from `DataViewsAPI.GetAllDataViews`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetAllDataViewsRequest struct via the builder pattern


### Return type

[**GetAllDataViews200Response**](GetAllDataViews200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDataView

> DataViewResponseObject GetDataView(ctx, viewId).Execute()

Retrieves a single data view by identifier.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    viewId := "ff959d40-b880-11e8-a6d9-e546fe2bba5f" // string | An identifier for the data view.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DataViewsAPI.GetDataView(context.Background(), viewId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.GetDataView``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDataView`: DataViewResponseObject
    fmt.Fprintf(os.Stdout, "Response from `DataViewsAPI.GetDataView`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**viewId** | **string** | An identifier for the data view. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDataViewRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DataViewResponseObject**](DataViewResponseObject.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDefaultDataView

> GetDefaultDataView200Response GetDefaultDataView(ctx).Execute()

Retrieves the default data view identifier.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DataViewsAPI.GetDefaultDataView(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.GetDefaultDataView``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetDefaultDataView`: GetDefaultDataView200Response
    fmt.Fprintf(os.Stdout, "Response from `DataViewsAPI.GetDefaultDataView`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetDefaultDataViewRequest struct via the builder pattern


### Return type

[**GetDefaultDataView200Response**](GetDefaultDataView200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRuntimeField

> CreateUpdateRuntimeField200Response GetRuntimeField(ctx, fieldName, viewId).Execute()

Retrieves a runtime field.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    fieldName := "hour_of_day" // string | The name of the runtime field.
    viewId := "ff959d40-b880-11e8-a6d9-e546fe2bba5f" // string | An identifier for the data view.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DataViewsAPI.GetRuntimeField(context.Background(), fieldName, viewId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.GetRuntimeField``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetRuntimeField`: CreateUpdateRuntimeField200Response
    fmt.Fprintf(os.Stdout, "Response from `DataViewsAPI.GetRuntimeField`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fieldName** | **string** | The name of the runtime field. | 
**viewId** | **string** | An identifier for the data view. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRuntimeFieldRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**CreateUpdateRuntimeField200Response**](CreateUpdateRuntimeField200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetDefaultDatailView

> UpdateFieldsMetadata200Response SetDefaultDatailView(ctx).KbnXsrf(kbnXsrf).SetDefaultDatailViewRequest(setDefaultDatailViewRequest).Execute()

Sets the default data view identifier.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    setDefaultDatailViewRequest := *openapiclient.NewSetDefaultDatailViewRequest(interface{}(123)) // SetDefaultDatailViewRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DataViewsAPI.SetDefaultDatailView(context.Background()).KbnXsrf(kbnXsrf).SetDefaultDatailViewRequest(setDefaultDatailViewRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.SetDefaultDatailView``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SetDefaultDatailView`: UpdateFieldsMetadata200Response
    fmt.Fprintf(os.Stdout, "Response from `DataViewsAPI.SetDefaultDatailView`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetDefaultDatailViewRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 
 **setDefaultDatailViewRequest** | [**SetDefaultDatailViewRequest**](SetDefaultDatailViewRequest.md) |  | 

### Return type

[**UpdateFieldsMetadata200Response**](UpdateFieldsMetadata200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDataView

> DataViewResponseObject UpdateDataView(ctx, viewId).KbnXsrf(kbnXsrf).UpdateDataViewRequestObject(updateDataViewRequestObject).Execute()

Updates a data view.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    viewId := "ff959d40-b880-11e8-a6d9-e546fe2bba5f" // string | An identifier for the data view.
    updateDataViewRequestObject := *openapiclient.NewUpdateDataViewRequestObject(*openapiclient.NewUpdateDataViewRequestObjectDataView()) // UpdateDataViewRequestObject | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DataViewsAPI.UpdateDataView(context.Background(), viewId).KbnXsrf(kbnXsrf).UpdateDataViewRequestObject(updateDataViewRequestObject).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.UpdateDataView``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateDataView`: DataViewResponseObject
    fmt.Fprintf(os.Stdout, "Response from `DataViewsAPI.UpdateDataView`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**viewId** | **string** | An identifier for the data view. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDataViewRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **updateDataViewRequestObject** | [**UpdateDataViewRequestObject**](UpdateDataViewRequestObject.md) |  | 

### Return type

[**DataViewResponseObject**](DataViewResponseObject.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateFieldsMetadata

> UpdateFieldsMetadata200Response UpdateFieldsMetadata(ctx, viewId).KbnXsrf(kbnXsrf).UpdateFieldsMetadataRequest(updateFieldsMetadataRequest).Execute()

Update fields presentation metadata such as count, customLabel and format.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    kbnXsrf := "kbnXsrf_example" // string | Cross-site request forgery protection
    viewId := "ff959d40-b880-11e8-a6d9-e546fe2bba5f" // string | An identifier for the data view.
    updateFieldsMetadataRequest := *openapiclient.NewUpdateFieldsMetadataRequest(map[string]interface{}(123)) // UpdateFieldsMetadataRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DataViewsAPI.UpdateFieldsMetadata(context.Background(), viewId).KbnXsrf(kbnXsrf).UpdateFieldsMetadataRequest(updateFieldsMetadataRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.UpdateFieldsMetadata``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateFieldsMetadata`: UpdateFieldsMetadata200Response
    fmt.Fprintf(os.Stdout, "Response from `DataViewsAPI.UpdateFieldsMetadata`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**viewId** | **string** | An identifier for the data view. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateFieldsMetadataRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kbnXsrf** | **string** | Cross-site request forgery protection | 

 **updateFieldsMetadataRequest** | [**UpdateFieldsMetadataRequest**](UpdateFieldsMetadataRequest.md) |  | 

### Return type

[**UpdateFieldsMetadata200Response**](UpdateFieldsMetadata200Response.md)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateRuntimeField

> UpdateRuntimeField(ctx, fieldName, viewId).UpdateRuntimeFieldRequest(updateRuntimeFieldRequest).Execute()

Update an existing runtime field.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/elastic/terraform-provider-elasticstack/data_views"
)

func main() {
    fieldName := "hour_of_day" // string | The name of the runtime field.
    viewId := "ff959d40-b880-11e8-a6d9-e546fe2bba5f" // string | An identifier for the data view.
    updateRuntimeFieldRequest := *openapiclient.NewUpdateRuntimeFieldRequest(map[string]interface{}(123)) // UpdateRuntimeFieldRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    r, err := apiClient.DataViewsAPI.UpdateRuntimeField(context.Background(), fieldName, viewId).UpdateRuntimeFieldRequest(updateRuntimeFieldRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DataViewsAPI.UpdateRuntimeField``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fieldName** | **string** | The name of the runtime field. | 
**viewId** | **string** | An identifier for the data view. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRuntimeFieldRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **updateRuntimeFieldRequest** | [**UpdateRuntimeFieldRequest**](UpdateRuntimeFieldRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[basicAuth](../README.md#basicAuth), [apiKeyAuth](../README.md#apiKeyAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

