# UpdateConnectorRequestServicenowItom

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesServicenowItom**](ConfigPropertiesServicenowItom.md) |  | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesServicenow**](SecretsPropertiesServicenow.md) |  | 

## Methods

### NewUpdateConnectorRequestServicenowItom

`func NewUpdateConnectorRequestServicenowItom(config ConfigPropertiesServicenowItom, name string, secrets SecretsPropertiesServicenow, ) *UpdateConnectorRequestServicenowItom`

NewUpdateConnectorRequestServicenowItom instantiates a new UpdateConnectorRequestServicenowItom object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateConnectorRequestServicenowItomWithDefaults

`func NewUpdateConnectorRequestServicenowItomWithDefaults() *UpdateConnectorRequestServicenowItom`

NewUpdateConnectorRequestServicenowItomWithDefaults instantiates a new UpdateConnectorRequestServicenowItom object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *UpdateConnectorRequestServicenowItom) GetConfig() ConfigPropertiesServicenowItom`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *UpdateConnectorRequestServicenowItom) GetConfigOk() (*ConfigPropertiesServicenowItom, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *UpdateConnectorRequestServicenowItom) SetConfig(v ConfigPropertiesServicenowItom)`

SetConfig sets Config field to given value.


### GetName

`func (o *UpdateConnectorRequestServicenowItom) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateConnectorRequestServicenowItom) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateConnectorRequestServicenowItom) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *UpdateConnectorRequestServicenowItom) GetSecrets() SecretsPropertiesServicenow`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *UpdateConnectorRequestServicenowItom) GetSecretsOk() (*SecretsPropertiesServicenow, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *UpdateConnectorRequestServicenowItom) SetSecrets(v SecretsPropertiesServicenow)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


