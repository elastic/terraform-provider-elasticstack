# ConnectorResponsePropertiesServicenowItom

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesServicenowItom**](ConfigPropertiesServicenowItom.md) |  | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Id** | **string** | The identifier for the connector. | 
**IsDeprecated** | **bool** | Indicates whether the connector type is deprecated. | 
**IsMissingSecrets** | Pointer to **bool** | Indicates whether secrets are missing for the connector. Secrets configuration properties vary depending on the connector type. | [optional] 
**IsPreconfigured** | **bool** | Indicates whether it is a preconfigured connector. If true, the &#x60;config&#x60; and &#x60;is_missing_secrets&#x60; properties are omitted from the response. | 
**Name** | **string** | The display name for the connector. | 

## Methods

### NewConnectorResponsePropertiesServicenowItom

`func NewConnectorResponsePropertiesServicenowItom(config ConfigPropertiesServicenowItom, connectorTypeId string, id string, isDeprecated bool, isPreconfigured bool, name string, ) *ConnectorResponsePropertiesServicenowItom`

NewConnectorResponsePropertiesServicenowItom instantiates a new ConnectorResponsePropertiesServicenowItom object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnectorResponsePropertiesServicenowItomWithDefaults

`func NewConnectorResponsePropertiesServicenowItomWithDefaults() *ConnectorResponsePropertiesServicenowItom`

NewConnectorResponsePropertiesServicenowItomWithDefaults instantiates a new ConnectorResponsePropertiesServicenowItom object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *ConnectorResponsePropertiesServicenowItom) GetConfig() ConfigPropertiesServicenowItom`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *ConnectorResponsePropertiesServicenowItom) GetConfigOk() (*ConfigPropertiesServicenowItom, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *ConnectorResponsePropertiesServicenowItom) SetConfig(v ConfigPropertiesServicenowItom)`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *ConnectorResponsePropertiesServicenowItom) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *ConnectorResponsePropertiesServicenowItom) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *ConnectorResponsePropertiesServicenowItom) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetId

`func (o *ConnectorResponsePropertiesServicenowItom) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ConnectorResponsePropertiesServicenowItom) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ConnectorResponsePropertiesServicenowItom) SetId(v string)`

SetId sets Id field to given value.


### GetIsDeprecated

`func (o *ConnectorResponsePropertiesServicenowItom) GetIsDeprecated() bool`

GetIsDeprecated returns the IsDeprecated field if non-nil, zero value otherwise.

### GetIsDeprecatedOk

`func (o *ConnectorResponsePropertiesServicenowItom) GetIsDeprecatedOk() (*bool, bool)`

GetIsDeprecatedOk returns a tuple with the IsDeprecated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDeprecated

`func (o *ConnectorResponsePropertiesServicenowItom) SetIsDeprecated(v bool)`

SetIsDeprecated sets IsDeprecated field to given value.


### GetIsMissingSecrets

`func (o *ConnectorResponsePropertiesServicenowItom) GetIsMissingSecrets() bool`

GetIsMissingSecrets returns the IsMissingSecrets field if non-nil, zero value otherwise.

### GetIsMissingSecretsOk

`func (o *ConnectorResponsePropertiesServicenowItom) GetIsMissingSecretsOk() (*bool, bool)`

GetIsMissingSecretsOk returns a tuple with the IsMissingSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsMissingSecrets

`func (o *ConnectorResponsePropertiesServicenowItom) SetIsMissingSecrets(v bool)`

SetIsMissingSecrets sets IsMissingSecrets field to given value.

### HasIsMissingSecrets

`func (o *ConnectorResponsePropertiesServicenowItom) HasIsMissingSecrets() bool`

HasIsMissingSecrets returns a boolean if a field has been set.

### GetIsPreconfigured

`func (o *ConnectorResponsePropertiesServicenowItom) GetIsPreconfigured() bool`

GetIsPreconfigured returns the IsPreconfigured field if non-nil, zero value otherwise.

### GetIsPreconfiguredOk

`func (o *ConnectorResponsePropertiesServicenowItom) GetIsPreconfiguredOk() (*bool, bool)`

GetIsPreconfiguredOk returns a tuple with the IsPreconfigured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsPreconfigured

`func (o *ConnectorResponsePropertiesServicenowItom) SetIsPreconfigured(v bool)`

SetIsPreconfigured sets IsPreconfigured field to given value.


### GetName

`func (o *ConnectorResponsePropertiesServicenowItom) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ConnectorResponsePropertiesServicenowItom) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ConnectorResponsePropertiesServicenowItom) SetName(v string)`

SetName sets Name field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


