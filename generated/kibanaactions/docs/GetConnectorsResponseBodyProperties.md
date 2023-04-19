# GetConnectorsResponseBodyProperties

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConnectorTypeId** | [***ConnectorTypes**](connector_types.md) |  | [default to null]
**Config** | [**ModelMap**](interface{}.md) | The configuration for the connector. Configuration properties vary depending on the connector type. | [optional] [default to null]
**Id** | **string** | The identifier for the connector. | [default to null]
**IsDeprecated** | **bool** |  | [default to null]
**IsMissingSecrets** | **bool** |  | [optional] [default to null]
**IsPreconfigured** | **bool** |  | [default to null]
**Name** | **string** | The display name for the connector. | [default to null]
**ReferencedByCount** | **int32** | Indicates the number of saved objects that reference the connector. If &#x60;is_preconfigured&#x60; is true, this value is not calculated. | [default to 0]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

