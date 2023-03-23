# ConfigPropertiesSwimlane

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiUrl** | **string** | The Swimlane instance URL. | 
**AppId** | **string** | The Swimlane application ID. | 
**ConnectorType** | **string** | The type of connector. Valid values are &#x60;all&#x60;, &#x60;alerts&#x60;, and &#x60;cases&#x60;. | 
**Mappings** | Pointer to [**ConnectorMappingsPropertiesForASwimlaneConnector**](ConnectorMappingsPropertiesForASwimlaneConnector.md) |  | [optional] 

## Methods

### NewConfigPropertiesSwimlane

`func NewConfigPropertiesSwimlane(apiUrl string, appId string, connectorType string, ) *ConfigPropertiesSwimlane`

NewConfigPropertiesSwimlane instantiates a new ConfigPropertiesSwimlane object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigPropertiesSwimlaneWithDefaults

`func NewConfigPropertiesSwimlaneWithDefaults() *ConfigPropertiesSwimlane`

NewConfigPropertiesSwimlaneWithDefaults instantiates a new ConfigPropertiesSwimlane object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiUrl

`func (o *ConfigPropertiesSwimlane) GetApiUrl() string`

GetApiUrl returns the ApiUrl field if non-nil, zero value otherwise.

### GetApiUrlOk

`func (o *ConfigPropertiesSwimlane) GetApiUrlOk() (*string, bool)`

GetApiUrlOk returns a tuple with the ApiUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiUrl

`func (o *ConfigPropertiesSwimlane) SetApiUrl(v string)`

SetApiUrl sets ApiUrl field to given value.


### GetAppId

`func (o *ConfigPropertiesSwimlane) GetAppId() string`

GetAppId returns the AppId field if non-nil, zero value otherwise.

### GetAppIdOk

`func (o *ConfigPropertiesSwimlane) GetAppIdOk() (*string, bool)`

GetAppIdOk returns a tuple with the AppId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAppId

`func (o *ConfigPropertiesSwimlane) SetAppId(v string)`

SetAppId sets AppId field to given value.


### GetConnectorType

`func (o *ConfigPropertiesSwimlane) GetConnectorType() string`

GetConnectorType returns the ConnectorType field if non-nil, zero value otherwise.

### GetConnectorTypeOk

`func (o *ConfigPropertiesSwimlane) GetConnectorTypeOk() (*string, bool)`

GetConnectorTypeOk returns a tuple with the ConnectorType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorType

`func (o *ConfigPropertiesSwimlane) SetConnectorType(v string)`

SetConnectorType sets ConnectorType field to given value.


### GetMappings

`func (o *ConfigPropertiesSwimlane) GetMappings() ConnectorMappingsPropertiesForASwimlaneConnector`

GetMappings returns the Mappings field if non-nil, zero value otherwise.

### GetMappingsOk

`func (o *ConfigPropertiesSwimlane) GetMappingsOk() (*ConnectorMappingsPropertiesForASwimlaneConnector, bool)`

GetMappingsOk returns a tuple with the Mappings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMappings

`func (o *ConfigPropertiesSwimlane) SetMappings(v ConnectorMappingsPropertiesForASwimlaneConnector)`

SetMappings sets Mappings field to given value.

### HasMappings

`func (o *ConfigPropertiesSwimlane) HasMappings() bool`

HasMappings returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


