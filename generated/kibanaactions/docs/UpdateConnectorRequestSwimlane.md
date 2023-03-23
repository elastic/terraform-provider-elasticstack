# UpdateConnectorRequestSwimlane

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesSwimlane**](ConfigPropertiesSwimlane.md) |  | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesSwimlane**](SecretsPropertiesSwimlane.md) |  | 

## Methods

### NewUpdateConnectorRequestSwimlane

`func NewUpdateConnectorRequestSwimlane(config ConfigPropertiesSwimlane, name string, secrets SecretsPropertiesSwimlane, ) *UpdateConnectorRequestSwimlane`

NewUpdateConnectorRequestSwimlane instantiates a new UpdateConnectorRequestSwimlane object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateConnectorRequestSwimlaneWithDefaults

`func NewUpdateConnectorRequestSwimlaneWithDefaults() *UpdateConnectorRequestSwimlane`

NewUpdateConnectorRequestSwimlaneWithDefaults instantiates a new UpdateConnectorRequestSwimlane object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *UpdateConnectorRequestSwimlane) GetConfig() ConfigPropertiesSwimlane`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *UpdateConnectorRequestSwimlane) GetConfigOk() (*ConfigPropertiesSwimlane, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *UpdateConnectorRequestSwimlane) SetConfig(v ConfigPropertiesSwimlane)`

SetConfig sets Config field to given value.


### GetName

`func (o *UpdateConnectorRequestSwimlane) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateConnectorRequestSwimlane) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateConnectorRequestSwimlane) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *UpdateConnectorRequestSwimlane) GetSecrets() SecretsPropertiesSwimlane`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *UpdateConnectorRequestSwimlane) GetSecretsOk() (*SecretsPropertiesSwimlane, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *UpdateConnectorRequestSwimlane) SetSecrets(v SecretsPropertiesSwimlane)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


