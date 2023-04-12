# UpdateConnectorRequestJira

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesJira**](ConfigPropertiesJira.md) |  | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesJira**](SecretsPropertiesJira.md) |  | 

## Methods

### NewUpdateConnectorRequestJira

`func NewUpdateConnectorRequestJira(config ConfigPropertiesJira, name string, secrets SecretsPropertiesJira, ) *UpdateConnectorRequestJira`

NewUpdateConnectorRequestJira instantiates a new UpdateConnectorRequestJira object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateConnectorRequestJiraWithDefaults

`func NewUpdateConnectorRequestJiraWithDefaults() *UpdateConnectorRequestJira`

NewUpdateConnectorRequestJiraWithDefaults instantiates a new UpdateConnectorRequestJira object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *UpdateConnectorRequestJira) GetConfig() ConfigPropertiesJira`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *UpdateConnectorRequestJira) GetConfigOk() (*ConfigPropertiesJira, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *UpdateConnectorRequestJira) SetConfig(v ConfigPropertiesJira)`

SetConfig sets Config field to given value.


### GetName

`func (o *UpdateConnectorRequestJira) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateConnectorRequestJira) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateConnectorRequestJira) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *UpdateConnectorRequestJira) GetSecrets() SecretsPropertiesJira`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *UpdateConnectorRequestJira) GetSecretsOk() (*SecretsPropertiesJira, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *UpdateConnectorRequestJira) SetSecrets(v SecretsPropertiesJira)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


