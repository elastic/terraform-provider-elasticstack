# LegacyUpdateConnectorRequestBodyProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | Pointer to **map[string]interface{}** | The new connector configuration. Configuration properties vary depending on the connector type. | [optional] 
**Name** | Pointer to **string** | The new name for the connector. | [optional] 
**Secrets** | Pointer to **map[string]interface{}** | The updated secrets configuration for the connector. Secrets properties vary depending on the connector type. | [optional] 

## Methods

### NewLegacyUpdateConnectorRequestBodyProperties

`func NewLegacyUpdateConnectorRequestBodyProperties() *LegacyUpdateConnectorRequestBodyProperties`

NewLegacyUpdateConnectorRequestBodyProperties instantiates a new LegacyUpdateConnectorRequestBodyProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyUpdateConnectorRequestBodyPropertiesWithDefaults

`func NewLegacyUpdateConnectorRequestBodyPropertiesWithDefaults() *LegacyUpdateConnectorRequestBodyProperties`

NewLegacyUpdateConnectorRequestBodyPropertiesWithDefaults instantiates a new LegacyUpdateConnectorRequestBodyProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *LegacyUpdateConnectorRequestBodyProperties) GetConfig() map[string]interface{}`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *LegacyUpdateConnectorRequestBodyProperties) GetConfigOk() (*map[string]interface{}, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *LegacyUpdateConnectorRequestBodyProperties) SetConfig(v map[string]interface{})`

SetConfig sets Config field to given value.

### HasConfig

`func (o *LegacyUpdateConnectorRequestBodyProperties) HasConfig() bool`

HasConfig returns a boolean if a field has been set.

### GetName

`func (o *LegacyUpdateConnectorRequestBodyProperties) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *LegacyUpdateConnectorRequestBodyProperties) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *LegacyUpdateConnectorRequestBodyProperties) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *LegacyUpdateConnectorRequestBodyProperties) HasName() bool`

HasName returns a boolean if a field has been set.

### GetSecrets

`func (o *LegacyUpdateConnectorRequestBodyProperties) GetSecrets() map[string]interface{}`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *LegacyUpdateConnectorRequestBodyProperties) GetSecretsOk() (*map[string]interface{}, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *LegacyUpdateConnectorRequestBodyProperties) SetSecrets(v map[string]interface{})`

SetSecrets sets Secrets field to given value.

### HasSecrets

`func (o *LegacyUpdateConnectorRequestBodyProperties) HasSecrets() bool`

HasSecrets returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


