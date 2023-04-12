# UpdateConnectorRequestOpsgenie

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesOpsgenie**](ConfigPropertiesOpsgenie.md) |  | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesOpsgenie**](SecretsPropertiesOpsgenie.md) |  | 

## Methods

### NewUpdateConnectorRequestOpsgenie

`func NewUpdateConnectorRequestOpsgenie(config ConfigPropertiesOpsgenie, name string, secrets SecretsPropertiesOpsgenie, ) *UpdateConnectorRequestOpsgenie`

NewUpdateConnectorRequestOpsgenie instantiates a new UpdateConnectorRequestOpsgenie object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateConnectorRequestOpsgenieWithDefaults

`func NewUpdateConnectorRequestOpsgenieWithDefaults() *UpdateConnectorRequestOpsgenie`

NewUpdateConnectorRequestOpsgenieWithDefaults instantiates a new UpdateConnectorRequestOpsgenie object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *UpdateConnectorRequestOpsgenie) GetConfig() ConfigPropertiesOpsgenie`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *UpdateConnectorRequestOpsgenie) GetConfigOk() (*ConfigPropertiesOpsgenie, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *UpdateConnectorRequestOpsgenie) SetConfig(v ConfigPropertiesOpsgenie)`

SetConfig sets Config field to given value.


### GetName

`func (o *UpdateConnectorRequestOpsgenie) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateConnectorRequestOpsgenie) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateConnectorRequestOpsgenie) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *UpdateConnectorRequestOpsgenie) GetSecrets() SecretsPropertiesOpsgenie`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *UpdateConnectorRequestOpsgenie) GetSecretsOk() (*SecretsPropertiesOpsgenie, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *UpdateConnectorRequestOpsgenie) SetSecrets(v SecretsPropertiesOpsgenie)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


