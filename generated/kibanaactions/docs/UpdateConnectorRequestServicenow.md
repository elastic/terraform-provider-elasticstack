# UpdateConnectorRequestServicenow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesServicenow**](ConfigPropertiesServicenow.md) |  | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesServicenow**](SecretsPropertiesServicenow.md) |  | 

## Methods

### NewUpdateConnectorRequestServicenow

`func NewUpdateConnectorRequestServicenow(config ConfigPropertiesServicenow, name string, secrets SecretsPropertiesServicenow, ) *UpdateConnectorRequestServicenow`

NewUpdateConnectorRequestServicenow instantiates a new UpdateConnectorRequestServicenow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateConnectorRequestServicenowWithDefaults

`func NewUpdateConnectorRequestServicenowWithDefaults() *UpdateConnectorRequestServicenow`

NewUpdateConnectorRequestServicenowWithDefaults instantiates a new UpdateConnectorRequestServicenow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *UpdateConnectorRequestServicenow) GetConfig() ConfigPropertiesServicenow`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *UpdateConnectorRequestServicenow) GetConfigOk() (*ConfigPropertiesServicenow, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *UpdateConnectorRequestServicenow) SetConfig(v ConfigPropertiesServicenow)`

SetConfig sets Config field to given value.


### GetName

`func (o *UpdateConnectorRequestServicenow) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateConnectorRequestServicenow) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateConnectorRequestServicenow) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *UpdateConnectorRequestServicenow) GetSecrets() SecretsPropertiesServicenow`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *UpdateConnectorRequestServicenow) GetSecretsOk() (*SecretsPropertiesServicenow, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *UpdateConnectorRequestServicenow) SetSecrets(v SecretsPropertiesServicenow)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


