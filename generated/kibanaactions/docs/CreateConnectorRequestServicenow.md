# CreateConnectorRequestServicenow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesServicenow**](ConfigPropertiesServicenow.md) |  | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesServicenow**](SecretsPropertiesServicenow.md) |  | 

## Methods

### NewCreateConnectorRequestServicenow

`func NewCreateConnectorRequestServicenow(config ConfigPropertiesServicenow, connectorTypeId string, name string, secrets SecretsPropertiesServicenow, ) *CreateConnectorRequestServicenow`

NewCreateConnectorRequestServicenow instantiates a new CreateConnectorRequestServicenow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateConnectorRequestServicenowWithDefaults

`func NewCreateConnectorRequestServicenowWithDefaults() *CreateConnectorRequestServicenow`

NewCreateConnectorRequestServicenowWithDefaults instantiates a new CreateConnectorRequestServicenow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *CreateConnectorRequestServicenow) GetConfig() ConfigPropertiesServicenow`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *CreateConnectorRequestServicenow) GetConfigOk() (*ConfigPropertiesServicenow, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *CreateConnectorRequestServicenow) SetConfig(v ConfigPropertiesServicenow)`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *CreateConnectorRequestServicenow) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *CreateConnectorRequestServicenow) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *CreateConnectorRequestServicenow) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetName

`func (o *CreateConnectorRequestServicenow) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateConnectorRequestServicenow) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateConnectorRequestServicenow) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *CreateConnectorRequestServicenow) GetSecrets() SecretsPropertiesServicenow`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *CreateConnectorRequestServicenow) GetSecretsOk() (*SecretsPropertiesServicenow, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *CreateConnectorRequestServicenow) SetSecrets(v SecretsPropertiesServicenow)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


