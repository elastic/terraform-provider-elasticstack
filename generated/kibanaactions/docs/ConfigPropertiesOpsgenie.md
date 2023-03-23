# ConfigPropertiesOpsgenie

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiUrl** | **string** | The Opsgenie URL. For example, &#x60;https://api.opsgenie.com&#x60; or &#x60;https://api.eu.opsgenie.com&#x60;. If you are using the &#x60;xpack.actions.allowedHosts&#x60; setting, add the hostname to the allowed hosts.  | 

## Methods

### NewConfigPropertiesOpsgenie

`func NewConfigPropertiesOpsgenie(apiUrl string, ) *ConfigPropertiesOpsgenie`

NewConfigPropertiesOpsgenie instantiates a new ConfigPropertiesOpsgenie object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigPropertiesOpsgenieWithDefaults

`func NewConfigPropertiesOpsgenieWithDefaults() *ConfigPropertiesOpsgenie`

NewConfigPropertiesOpsgenieWithDefaults instantiates a new ConfigPropertiesOpsgenie object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiUrl

`func (o *ConfigPropertiesOpsgenie) GetApiUrl() string`

GetApiUrl returns the ApiUrl field if non-nil, zero value otherwise.

### GetApiUrlOk

`func (o *ConfigPropertiesOpsgenie) GetApiUrlOk() (*string, bool)`

GetApiUrlOk returns a tuple with the ApiUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiUrl

`func (o *ConfigPropertiesOpsgenie) SetApiUrl(v string)`

SetApiUrl sets ApiUrl field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


