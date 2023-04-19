# ConfigPropertiesServicenow

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiUrl** | **string** | The ServiceNow instance URL. | [default to null]
**ClientId** | **string** | The client ID assigned to your OAuth application. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;.  | [optional] [default to null]
**IsOAuth** | **bool** | The type of authentication to use. The default value is false, which means basic authentication is used instead of open authorization (OAuth).  | [optional] [default to false]
**JwtKeyId** | **string** | The key identifier assigned to the JWT verifier map of your OAuth application. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;.  | [optional] [default to null]
**UserIdentifierValue** | **string** | The identifier to use for OAuth authentication. This identifier should be the user field you selected when you created an OAuth JWT API endpoint for external clients in your ServiceNow instance. For example, if the selected user field is &#x60;Email&#x60;, the user identifier should be the user&#x27;s email address. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;.  | [optional] [default to null]
**UsesTableApi** | **bool** | Determines whether the connector uses the Table API or the Import Set API. This property is supported only for ServiceNow ITSM and ServiceNow SecOps connectors.  NOTE: If this property is set to &#x60;false&#x60;, the Elastic application should be installed in ServiceNow.  | [optional] [default to true]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

