# LegacyCreateConnectorRequestProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActionTypeId** | Pointer to **string** | The connector type identifier. | [optional] 
**Config** | Pointer to **map[string]interface{}** | The configuration for the connector. Configuration properties vary depending on the connector type. | [optional] 
**Name** | Pointer to **string** | The display name for the connector. | [optional] 
**Secrets** | Pointer to **map[string]interface{}** | The secrets configuration for the connector. Secrets configuration properties vary depending on the connector type. NOTE: Remember these values. You must provide them each time you update the connector.  | [optional] 

## Methods

### NewLegacyCreateConnectorRequestProperties

`func NewLegacyCreateConnectorRequestProperties() *LegacyCreateConnectorRequestProperties`

NewLegacyCreateConnectorRequestProperties instantiates a new LegacyCreateConnectorRequestProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyCreateConnectorRequestPropertiesWithDefaults

`func NewLegacyCreateConnectorRequestPropertiesWithDefaults() *LegacyCreateConnectorRequestProperties`

NewLegacyCreateConnectorRequestPropertiesWithDefaults instantiates a new LegacyCreateConnectorRequestProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetActionTypeId

`func (o *LegacyCreateConnectorRequestProperties) GetActionTypeId() string`

GetActionTypeId returns the ActionTypeId field if non-nil, zero value otherwise.

### GetActionTypeIdOk

`func (o *LegacyCreateConnectorRequestProperties) GetActionTypeIdOk() (*string, bool)`

GetActionTypeIdOk returns a tuple with the ActionTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActionTypeId

`func (o *LegacyCreateConnectorRequestProperties) SetActionTypeId(v string)`

SetActionTypeId sets ActionTypeId field to given value.

### HasActionTypeId

`func (o *LegacyCreateConnectorRequestProperties) HasActionTypeId() bool`

HasActionTypeId returns a boolean if a field has been set.

### GetConfig

`func (o *LegacyCreateConnectorRequestProperties) GetConfig() map[string]interface{}`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *LegacyCreateConnectorRequestProperties) GetConfigOk() (*map[string]interface{}, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *LegacyCreateConnectorRequestProperties) SetConfig(v map[string]interface{})`

SetConfig sets Config field to given value.

### HasConfig

`func (o *LegacyCreateConnectorRequestProperties) HasConfig() bool`

HasConfig returns a boolean if a field has been set.

### GetName

`func (o *LegacyCreateConnectorRequestProperties) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *LegacyCreateConnectorRequestProperties) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *LegacyCreateConnectorRequestProperties) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *LegacyCreateConnectorRequestProperties) HasName() bool`

HasName returns a boolean if a field has been set.

### GetSecrets

`func (o *LegacyCreateConnectorRequestProperties) GetSecrets() map[string]interface{}`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *LegacyCreateConnectorRequestProperties) GetSecretsOk() (*map[string]interface{}, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *LegacyCreateConnectorRequestProperties) SetSecrets(v map[string]interface{})`

SetSecrets sets Secrets field to given value.

### HasSecrets

`func (o *LegacyCreateConnectorRequestProperties) HasSecrets() bool`

HasSecrets returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


