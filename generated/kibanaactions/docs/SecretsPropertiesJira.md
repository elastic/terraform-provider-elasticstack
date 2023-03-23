# SecretsPropertiesJira

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiToken** | **string** | The Jira API authentication token for HTTP basic authentication. | 
**Email** | **string** | The account email for HTTP Basic authentication. | 

## Methods

### NewSecretsPropertiesJira

`func NewSecretsPropertiesJira(apiToken string, email string, ) *SecretsPropertiesJira`

NewSecretsPropertiesJira instantiates a new SecretsPropertiesJira object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretsPropertiesJiraWithDefaults

`func NewSecretsPropertiesJiraWithDefaults() *SecretsPropertiesJira`

NewSecretsPropertiesJiraWithDefaults instantiates a new SecretsPropertiesJira object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiToken

`func (o *SecretsPropertiesJira) GetApiToken() string`

GetApiToken returns the ApiToken field if non-nil, zero value otherwise.

### GetApiTokenOk

`func (o *SecretsPropertiesJira) GetApiTokenOk() (*string, bool)`

GetApiTokenOk returns a tuple with the ApiToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiToken

`func (o *SecretsPropertiesJira) SetApiToken(v string)`

SetApiToken sets ApiToken field to given value.


### GetEmail

`func (o *SecretsPropertiesJira) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *SecretsPropertiesJira) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *SecretsPropertiesJira) SetEmail(v string)`

SetEmail sets Email field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


