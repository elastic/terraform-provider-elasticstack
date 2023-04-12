# CreateConnectorRequestTeams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConnectorTypeId** | **string** | The type of connector. | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | **map[string]interface{}** | Defines secrets for connectors when type is &#x60;.teams&#x60;. | 

## Methods

### NewCreateConnectorRequestTeams

`func NewCreateConnectorRequestTeams(connectorTypeId string, name string, secrets map[string]interface{}, ) *CreateConnectorRequestTeams`

NewCreateConnectorRequestTeams instantiates a new CreateConnectorRequestTeams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateConnectorRequestTeamsWithDefaults

`func NewCreateConnectorRequestTeamsWithDefaults() *CreateConnectorRequestTeams`

NewCreateConnectorRequestTeamsWithDefaults instantiates a new CreateConnectorRequestTeams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConnectorTypeId

`func (o *CreateConnectorRequestTeams) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *CreateConnectorRequestTeams) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *CreateConnectorRequestTeams) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetName

`func (o *CreateConnectorRequestTeams) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateConnectorRequestTeams) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateConnectorRequestTeams) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *CreateConnectorRequestTeams) GetSecrets() map[string]interface{}`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *CreateConnectorRequestTeams) GetSecretsOk() (*map[string]interface{}, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *CreateConnectorRequestTeams) SetSecrets(v map[string]interface{})`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


