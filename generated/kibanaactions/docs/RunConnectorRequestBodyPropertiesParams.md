# RunConnectorRequestBodyPropertiesParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Documents** | **[]map[string]interface{}** | The documents in JSON format for index connectors. | 
**Level** | Pointer to **string** | The log level of the message for server log connectors. | [optional] [default to "info"]
**Message** | **string** | The message for server log connectors. | 

## Methods

### NewRunConnectorRequestBodyPropertiesParams

`func NewRunConnectorRequestBodyPropertiesParams(documents []map[string]interface{}, message string, ) *RunConnectorRequestBodyPropertiesParams`

NewRunConnectorRequestBodyPropertiesParams instantiates a new RunConnectorRequestBodyPropertiesParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunConnectorRequestBodyPropertiesParamsWithDefaults

`func NewRunConnectorRequestBodyPropertiesParamsWithDefaults() *RunConnectorRequestBodyPropertiesParams`

NewRunConnectorRequestBodyPropertiesParamsWithDefaults instantiates a new RunConnectorRequestBodyPropertiesParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDocuments

`func (o *RunConnectorRequestBodyPropertiesParams) GetDocuments() []map[string]interface{}`

GetDocuments returns the Documents field if non-nil, zero value otherwise.

### GetDocumentsOk

`func (o *RunConnectorRequestBodyPropertiesParams) GetDocumentsOk() (*[]map[string]interface{}, bool)`

GetDocumentsOk returns a tuple with the Documents field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDocuments

`func (o *RunConnectorRequestBodyPropertiesParams) SetDocuments(v []map[string]interface{})`

SetDocuments sets Documents field to given value.


### GetLevel

`func (o *RunConnectorRequestBodyPropertiesParams) GetLevel() string`

GetLevel returns the Level field if non-nil, zero value otherwise.

### GetLevelOk

`func (o *RunConnectorRequestBodyPropertiesParams) GetLevelOk() (*string, bool)`

GetLevelOk returns a tuple with the Level field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLevel

`func (o *RunConnectorRequestBodyPropertiesParams) SetLevel(v string)`

SetLevel sets Level field to given value.

### HasLevel

`func (o *RunConnectorRequestBodyPropertiesParams) HasLevel() bool`

HasLevel returns a boolean if a field has been set.

### GetMessage

`func (o *RunConnectorRequestBodyPropertiesParams) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *RunConnectorRequestBodyPropertiesParams) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *RunConnectorRequestBodyPropertiesParams) SetMessage(v string)`

SetMessage sets Message field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


