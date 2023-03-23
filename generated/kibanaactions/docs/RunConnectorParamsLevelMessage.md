# RunConnectorParamsLevelMessage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Level** | Pointer to **string** | The log level of the message for server log connectors. | [optional] [default to "info"]
**Message** | **string** | The message for server log connectors. | 

## Methods

### NewRunConnectorParamsLevelMessage

`func NewRunConnectorParamsLevelMessage(message string, ) *RunConnectorParamsLevelMessage`

NewRunConnectorParamsLevelMessage instantiates a new RunConnectorParamsLevelMessage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRunConnectorParamsLevelMessageWithDefaults

`func NewRunConnectorParamsLevelMessageWithDefaults() *RunConnectorParamsLevelMessage`

NewRunConnectorParamsLevelMessageWithDefaults instantiates a new RunConnectorParamsLevelMessage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLevel

`func (o *RunConnectorParamsLevelMessage) GetLevel() string`

GetLevel returns the Level field if non-nil, zero value otherwise.

### GetLevelOk

`func (o *RunConnectorParamsLevelMessage) GetLevelOk() (*string, bool)`

GetLevelOk returns a tuple with the Level field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLevel

`func (o *RunConnectorParamsLevelMessage) SetLevel(v string)`

SetLevel sets Level field to given value.

### HasLevel

`func (o *RunConnectorParamsLevelMessage) HasLevel() bool`

HasLevel returns a boolean if a field has been set.

### GetMessage

`func (o *RunConnectorParamsLevelMessage) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *RunConnectorParamsLevelMessage) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *RunConnectorParamsLevelMessage) SetMessage(v string)`

SetMessage sets Message field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


