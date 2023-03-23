# UpdateConnectorRequestResilient

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesResilient**](ConfigPropertiesResilient.md) |  | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesResilient**](SecretsPropertiesResilient.md) |  | 

## Methods

### NewUpdateConnectorRequestResilient

`func NewUpdateConnectorRequestResilient(config ConfigPropertiesResilient, name string, secrets SecretsPropertiesResilient, ) *UpdateConnectorRequestResilient`

NewUpdateConnectorRequestResilient instantiates a new UpdateConnectorRequestResilient object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateConnectorRequestResilientWithDefaults

`func NewUpdateConnectorRequestResilientWithDefaults() *UpdateConnectorRequestResilient`

NewUpdateConnectorRequestResilientWithDefaults instantiates a new UpdateConnectorRequestResilient object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *UpdateConnectorRequestResilient) GetConfig() ConfigPropertiesResilient`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *UpdateConnectorRequestResilient) GetConfigOk() (*ConfigPropertiesResilient, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *UpdateConnectorRequestResilient) SetConfig(v ConfigPropertiesResilient)`

SetConfig sets Config field to given value.


### GetName

`func (o *UpdateConnectorRequestResilient) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateConnectorRequestResilient) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateConnectorRequestResilient) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *UpdateConnectorRequestResilient) GetSecrets() SecretsPropertiesResilient`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *UpdateConnectorRequestResilient) GetSecretsOk() (*SecretsPropertiesResilient, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *UpdateConnectorRequestResilient) SetSecrets(v SecretsPropertiesResilient)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


