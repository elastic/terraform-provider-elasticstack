# ConfigPropertiesJira

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiUrl** | **string** | The Jira instance URL. | 
**ProjectKey** | **string** | The Jira project key. | 

## Methods

### NewConfigPropertiesJira

`func NewConfigPropertiesJira(apiUrl string, projectKey string, ) *ConfigPropertiesJira`

NewConfigPropertiesJira instantiates a new ConfigPropertiesJira object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigPropertiesJiraWithDefaults

`func NewConfigPropertiesJiraWithDefaults() *ConfigPropertiesJira`

NewConfigPropertiesJiraWithDefaults instantiates a new ConfigPropertiesJira object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiUrl

`func (o *ConfigPropertiesJira) GetApiUrl() string`

GetApiUrl returns the ApiUrl field if non-nil, zero value otherwise.

### GetApiUrlOk

`func (o *ConfigPropertiesJira) GetApiUrlOk() (*string, bool)`

GetApiUrlOk returns a tuple with the ApiUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiUrl

`func (o *ConfigPropertiesJira) SetApiUrl(v string)`

SetApiUrl sets ApiUrl field to given value.


### GetProjectKey

`func (o *ConfigPropertiesJira) GetProjectKey() string`

GetProjectKey returns the ProjectKey field if non-nil, zero value otherwise.

### GetProjectKeyOk

`func (o *ConfigPropertiesJira) GetProjectKeyOk() (*string, bool)`

GetProjectKeyOk returns a tuple with the ProjectKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectKey

`func (o *ConfigPropertiesJira) SetProjectKey(v string)`

SetProjectKey sets ProjectKey field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


