# ConnectorResponsePropertiesIndex

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesIndex**](ConfigPropertiesIndex.md) |  | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Id** | **string** | The identifier for the connector. | 
**IsDeprecated** | **bool** | Indicates whether the connector type is deprecated. | 
**IsMissingSecrets** | Pointer to **bool** | Indicates whether secrets are missing for the connector. Secrets configuration properties vary depending on the connector type. | [optional] 
**IsPreconfigured** | **bool** | Indicates whether it is a preconfigured connector. If true, the &#x60;config&#x60; and &#x60;is_missing_secrets&#x60; properties are omitted from the response. | 
**Name** | **string** | The display name for the connector. | 

## Methods

### NewConnectorResponsePropertiesIndex

`func NewConnectorResponsePropertiesIndex(config ConfigPropertiesIndex, connectorTypeId string, id string, isDeprecated bool, isPreconfigured bool, name string, ) *ConnectorResponsePropertiesIndex`

NewConnectorResponsePropertiesIndex instantiates a new ConnectorResponsePropertiesIndex object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnectorResponsePropertiesIndexWithDefaults

`func NewConnectorResponsePropertiesIndexWithDefaults() *ConnectorResponsePropertiesIndex`

NewConnectorResponsePropertiesIndexWithDefaults instantiates a new ConnectorResponsePropertiesIndex object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *ConnectorResponsePropertiesIndex) GetConfig() ConfigPropertiesIndex`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *ConnectorResponsePropertiesIndex) GetConfigOk() (*ConfigPropertiesIndex, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *ConnectorResponsePropertiesIndex) SetConfig(v ConfigPropertiesIndex)`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *ConnectorResponsePropertiesIndex) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *ConnectorResponsePropertiesIndex) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *ConnectorResponsePropertiesIndex) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetId

`func (o *ConnectorResponsePropertiesIndex) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ConnectorResponsePropertiesIndex) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ConnectorResponsePropertiesIndex) SetId(v string)`

SetId sets Id field to given value.


### GetIsDeprecated

`func (o *ConnectorResponsePropertiesIndex) GetIsDeprecated() bool`

GetIsDeprecated returns the IsDeprecated field if non-nil, zero value otherwise.

### GetIsDeprecatedOk

`func (o *ConnectorResponsePropertiesIndex) GetIsDeprecatedOk() (*bool, bool)`

GetIsDeprecatedOk returns a tuple with the IsDeprecated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDeprecated

`func (o *ConnectorResponsePropertiesIndex) SetIsDeprecated(v bool)`

SetIsDeprecated sets IsDeprecated field to given value.


### GetIsMissingSecrets

`func (o *ConnectorResponsePropertiesIndex) GetIsMissingSecrets() bool`

GetIsMissingSecrets returns the IsMissingSecrets field if non-nil, zero value otherwise.

### GetIsMissingSecretsOk

`func (o *ConnectorResponsePropertiesIndex) GetIsMissingSecretsOk() (*bool, bool)`

GetIsMissingSecretsOk returns a tuple with the IsMissingSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsMissingSecrets

`func (o *ConnectorResponsePropertiesIndex) SetIsMissingSecrets(v bool)`

SetIsMissingSecrets sets IsMissingSecrets field to given value.

### HasIsMissingSecrets

`func (o *ConnectorResponsePropertiesIndex) HasIsMissingSecrets() bool`

HasIsMissingSecrets returns a boolean if a field has been set.

### GetIsPreconfigured

`func (o *ConnectorResponsePropertiesIndex) GetIsPreconfigured() bool`

GetIsPreconfigured returns the IsPreconfigured field if non-nil, zero value otherwise.

### GetIsPreconfiguredOk

`func (o *ConnectorResponsePropertiesIndex) GetIsPreconfiguredOk() (*bool, bool)`

GetIsPreconfiguredOk returns a tuple with the IsPreconfigured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsPreconfigured

`func (o *ConnectorResponsePropertiesIndex) SetIsPreconfigured(v bool)`

SetIsPreconfigured sets IsPreconfigured field to given value.


### GetName

`func (o *ConnectorResponsePropertiesIndex) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ConnectorResponsePropertiesIndex) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ConnectorResponsePropertiesIndex) SetName(v string)`

SetName sets Name field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


