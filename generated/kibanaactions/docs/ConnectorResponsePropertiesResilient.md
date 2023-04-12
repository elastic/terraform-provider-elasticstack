# ConnectorResponsePropertiesResilient

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesResilient**](ConfigPropertiesResilient.md) |  | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Id** | **string** | The identifier for the connector. | 
**IsDeprecated** | **bool** | Indicates whether the connector type is deprecated. | 
**IsMissingSecrets** | Pointer to **bool** | Indicates whether secrets are missing for the connector. Secrets configuration properties vary depending on the connector type. | [optional] 
**IsPreconfigured** | **bool** | Indicates whether it is a preconfigured connector. If true, the &#x60;config&#x60; and &#x60;is_missing_secrets&#x60; properties are omitted from the response. | 
**Name** | **string** | The display name for the connector. | 

## Methods

### NewConnectorResponsePropertiesResilient

`func NewConnectorResponsePropertiesResilient(config ConfigPropertiesResilient, connectorTypeId string, id string, isDeprecated bool, isPreconfigured bool, name string, ) *ConnectorResponsePropertiesResilient`

NewConnectorResponsePropertiesResilient instantiates a new ConnectorResponsePropertiesResilient object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnectorResponsePropertiesResilientWithDefaults

`func NewConnectorResponsePropertiesResilientWithDefaults() *ConnectorResponsePropertiesResilient`

NewConnectorResponsePropertiesResilientWithDefaults instantiates a new ConnectorResponsePropertiesResilient object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *ConnectorResponsePropertiesResilient) GetConfig() ConfigPropertiesResilient`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *ConnectorResponsePropertiesResilient) GetConfigOk() (*ConfigPropertiesResilient, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *ConnectorResponsePropertiesResilient) SetConfig(v ConfigPropertiesResilient)`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *ConnectorResponsePropertiesResilient) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *ConnectorResponsePropertiesResilient) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *ConnectorResponsePropertiesResilient) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetId

`func (o *ConnectorResponsePropertiesResilient) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ConnectorResponsePropertiesResilient) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ConnectorResponsePropertiesResilient) SetId(v string)`

SetId sets Id field to given value.


### GetIsDeprecated

`func (o *ConnectorResponsePropertiesResilient) GetIsDeprecated() bool`

GetIsDeprecated returns the IsDeprecated field if non-nil, zero value otherwise.

### GetIsDeprecatedOk

`func (o *ConnectorResponsePropertiesResilient) GetIsDeprecatedOk() (*bool, bool)`

GetIsDeprecatedOk returns a tuple with the IsDeprecated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDeprecated

`func (o *ConnectorResponsePropertiesResilient) SetIsDeprecated(v bool)`

SetIsDeprecated sets IsDeprecated field to given value.


### GetIsMissingSecrets

`func (o *ConnectorResponsePropertiesResilient) GetIsMissingSecrets() bool`

GetIsMissingSecrets returns the IsMissingSecrets field if non-nil, zero value otherwise.

### GetIsMissingSecretsOk

`func (o *ConnectorResponsePropertiesResilient) GetIsMissingSecretsOk() (*bool, bool)`

GetIsMissingSecretsOk returns a tuple with the IsMissingSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsMissingSecrets

`func (o *ConnectorResponsePropertiesResilient) SetIsMissingSecrets(v bool)`

SetIsMissingSecrets sets IsMissingSecrets field to given value.

### HasIsMissingSecrets

`func (o *ConnectorResponsePropertiesResilient) HasIsMissingSecrets() bool`

HasIsMissingSecrets returns a boolean if a field has been set.

### GetIsPreconfigured

`func (o *ConnectorResponsePropertiesResilient) GetIsPreconfigured() bool`

GetIsPreconfigured returns the IsPreconfigured field if non-nil, zero value otherwise.

### GetIsPreconfiguredOk

`func (o *ConnectorResponsePropertiesResilient) GetIsPreconfiguredOk() (*bool, bool)`

GetIsPreconfiguredOk returns a tuple with the IsPreconfigured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsPreconfigured

`func (o *ConnectorResponsePropertiesResilient) SetIsPreconfigured(v bool)`

SetIsPreconfigured sets IsPreconfigured field to given value.


### GetName

`func (o *ConnectorResponsePropertiesResilient) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ConnectorResponsePropertiesResilient) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ConnectorResponsePropertiesResilient) SetName(v string)`

SetName sets Name field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


