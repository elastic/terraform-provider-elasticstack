# {{classname}}

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

# **CreateConnector**
> ConnectorResponseProperties CreateConnector(ctx, body, kbnXsrf, spaceId)
Creates a connector.

You must have `all` privileges for the **Actions and Connectors** feature in the **Management** section of the Kibana feature privileges. 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CreateConnectorRequestBodyProperties**](CreateConnectorRequestBodyProperties.md)|  | 
  **kbnXsrf** | **string**| Cross-site request forgery protection | 
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**ConnectorResponseProperties**](connector_response_properties.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteConnector**
> DeleteConnector(ctx, kbnXsrf, connectorId, spaceId)
Deletes a connector.

You must have `all` privileges for the **Actions and Connectors** feature in the **Management** section of the Kibana feature privileges. WARNING: When you delete a connector, it cannot be recovered. 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **kbnXsrf** | **string**| Cross-site request forgery protection | 
  **connectorId** | **string**| An identifier for the connector. | 
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

 (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetConnector**
> ConnectorResponseProperties GetConnector(ctx, connectorId, spaceId)
Retrieves a connector by ID.

You must have `read` privileges for the **Actions and Connectors** feature in the **Management** section of the Kibana feature privileges. 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **connectorId** | **string**| An identifier for the connector. | 
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**ConnectorResponseProperties**](connector_response_properties.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetConnectorTypes**
> []InlineResponse200 GetConnectorTypes(ctx, spaceId, optional)
Retrieves a list of all connector types.

You do not need any Kibana feature privileges to run this API. 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 
 **optional** | ***ConnectorsApiGetConnectorTypesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ConnectorsApiGetConnectorTypesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **featureId** | [**optional.Interface of Features**](.md)| A filter to limit the retrieved connector types to those that support a specific feature (such as alerting or cases). | 

### Return type

[**[]InlineResponse200**](inline_response_200.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetConnectors**
> []GetConnectorsResponseBodyProperties GetConnectors(ctx, spaceId)
Retrieves all connectors.

You must have `read` privileges for the **Actions and Connectors** feature in the **Management** section of the Kibana feature privileges. 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**[]GetConnectorsResponseBodyProperties**](Get connectors response body properties.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LegacyCreateConnector**
> ActionResponseProperties LegacyCreateConnector(ctx, body, kbnXsrf, spaceId)
Creates a connector.

Deprecated in 7.13.0. Use the create connector API instead.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**LegacyCreateConnectorRequestProperties**](LegacyCreateConnectorRequestProperties.md)|  | 
  **kbnXsrf** | **string**| Cross-site request forgery protection | 
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**ActionResponseProperties**](action_response_properties.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LegacyDeleteConnector**
> LegacyDeleteConnector(ctx, kbnXsrf, actionId, spaceId)
Deletes a connector.

Deprecated in 7.13.0. Use the delete connector API instead. WARNING: When you delete a connector, it cannot be recovered. 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **kbnXsrf** | **string**| Cross-site request forgery protection | 
  **actionId** | **string**| An identifier for the action. | 
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

 (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LegacyGetConnector**
> ActionResponseProperties LegacyGetConnector(ctx, actionId, spaceId)
Retrieves a connector by ID.

Deprecated in 7.13.0. Use the get connector API instead.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **actionId** | **string**| An identifier for the action. | 
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**ActionResponseProperties**](action_response_properties.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LegacyGetConnectorTypes**
> []InlineResponse2002 LegacyGetConnectorTypes(ctx, spaceId)
Retrieves a list of all connector types.

Deprecated in 7.13.0. Use the get all connector types API instead.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**[]InlineResponse2002**](inline_response_200_2.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LegacyGetConnectors**
> []ActionResponseProperties LegacyGetConnectors(ctx, spaceId)
Retrieves all connectors.

Deprecated in 7.13.0. Use the get all connectors API instead.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**[]ActionResponseProperties**](action_response_properties.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LegacyRunConnector**
> InlineResponse2003 LegacyRunConnector(ctx, body, kbnXsrf, actionId, spaceId)
Runs a connector.

Deprecated in 7.13.0. Use the run connector API instead.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**LegacyRunConnectorRequestBodyProperties**](LegacyRunConnectorRequestBodyProperties.md)|  | 
  **kbnXsrf** | **string**| Cross-site request forgery protection | 
  **actionId** | **string**| An identifier for the action. | 
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**InlineResponse2003**](inline_response_200_3.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LegacyUpdateConnector**
> ActionResponseProperties LegacyUpdateConnector(ctx, body, kbnXsrf, actionId, spaceId)
Updates the attributes for a connector.

Deprecated in 7.13.0. Use the update connector API instead.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**LegacyUpdateConnectorRequestBodyProperties**](LegacyUpdateConnectorRequestBodyProperties.md)|  | 
  **kbnXsrf** | **string**| Cross-site request forgery protection | 
  **actionId** | **string**| An identifier for the action. | 
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**ActionResponseProperties**](action_response_properties.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RunConnector**
> InlineResponse2001 RunConnector(ctx, body, kbnXsrf, connectorId, spaceId)
Runs a connector.

You can use this API to test an action that involves interaction with Kibana services or integrations with third-party systems. You must have `read` privileges for the **Actions and Connectors** feature in the **Management** section of the Kibana feature privileges. If you use an index connector, you must also have `all`, `create`, `index`, or `write` indices privileges. 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**RunConnectorRequestBodyProperties**](RunConnectorRequestBodyProperties.md)|  | 
  **kbnXsrf** | **string**| Cross-site request forgery protection | 
  **connectorId** | **string**| An identifier for the connector. | 
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**InlineResponse2001**](inline_response_200_1.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateConnector**
> ConnectorResponseProperties UpdateConnector(ctx, body, kbnXsrf, connectorId, spaceId)
Updates the attributes for a connector.

You must have `all` privileges for the **Actions and Connectors** feature in the **Management** section of the Kibana feature privileges. 

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UpdateConnectorRequestBodyProperties**](UpdateConnectorRequestBodyProperties.md)|  | 
  **kbnXsrf** | **string**| Cross-site request forgery protection | 
  **connectorId** | **string**| An identifier for the connector. | 
  **spaceId** | **string**| An identifier for the space. If &#x60;/s/&#x60; and the identifier are omitted from the path, the default space is used. | 

### Return type

[**ConnectorResponseProperties**](connector_response_properties.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [basicAuth](../README.md#basicAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

