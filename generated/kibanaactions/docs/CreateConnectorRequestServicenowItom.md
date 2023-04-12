# CreateConnectorRequestServicenowItom

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesServicenowItom**](ConfigPropertiesServicenowItom.md) |  | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesServicenow**](SecretsPropertiesServicenow.md) |  | 

## Methods

### NewCreateConnectorRequestServicenowItom

`func NewCreateConnectorRequestServicenowItom(config ConfigPropertiesServicenowItom, connectorTypeId string, name string, secrets SecretsPropertiesServicenow, ) *CreateConnectorRequestServicenowItom`

NewCreateConnectorRequestServicenowItom instantiates a new CreateConnectorRequestServicenowItom object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateConnectorRequestServicenowItomWithDefaults

`func NewCreateConnectorRequestServicenowItomWithDefaults() *CreateConnectorRequestServicenowItom`

NewCreateConnectorRequestServicenowItomWithDefaults instantiates a new CreateConnectorRequestServicenowItom object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *CreateConnectorRequestServicenowItom) GetConfig() ConfigPropertiesServicenowItom`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *CreateConnectorRequestServicenowItom) GetConfigOk() (*ConfigPropertiesServicenowItom, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *CreateConnectorRequestServicenowItom) SetConfig(v ConfigPropertiesServicenowItom)`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *CreateConnectorRequestServicenowItom) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *CreateConnectorRequestServicenowItom) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *CreateConnectorRequestServicenowItom) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetName

`func (o *CreateConnectorRequestServicenowItom) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateConnectorRequestServicenowItom) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateConnectorRequestServicenowItom) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *CreateConnectorRequestServicenowItom) GetSecrets() SecretsPropertiesServicenow`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *CreateConnectorRequestServicenowItom) GetSecretsOk() (*SecretsPropertiesServicenow, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *CreateConnectorRequestServicenowItom) SetSecrets(v SecretsPropertiesServicenow)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


