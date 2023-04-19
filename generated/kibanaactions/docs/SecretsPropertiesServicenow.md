# SecretsPropertiesServicenow

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClientSecret** | **string** | The client secret assigned to your OAuth application. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;. | [optional] [default to null]
**Password** | **string** | The password for HTTP basic authentication. This property is required when &#x60;isOAuth&#x60; is &#x60;false&#x60;. | [optional] [default to null]
**PrivateKey** | **string** | The RSA private key that you created for use in ServiceNow. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;. | [optional] [default to null]
**PrivateKeyPassword** | **string** | The password for the RSA private key. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60; and you set a password on your private key. | [optional] [default to null]
**Username** | **string** | The username for HTTP basic authentication. This property is required when &#x60;isOAuth&#x60; is &#x60;false&#x60;. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

