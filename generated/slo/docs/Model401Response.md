# Model401Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StatusCode** | **float32** |  | 
**Error** | **string** |  | 
**Message** | **string** |  | 

## Methods

### NewModel401Response

`func NewModel401Response(statusCode float32, error_ string, message string, ) *Model401Response`

NewModel401Response instantiates a new Model401Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModel401ResponseWithDefaults

`func NewModel401ResponseWithDefaults() *Model401Response`

NewModel401ResponseWithDefaults instantiates a new Model401Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatusCode

`func (o *Model401Response) GetStatusCode() float32`

GetStatusCode returns the StatusCode field if non-nil, zero value otherwise.

### GetStatusCodeOk

`func (o *Model401Response) GetStatusCodeOk() (*float32, bool)`

GetStatusCodeOk returns a tuple with the StatusCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusCode

`func (o *Model401Response) SetStatusCode(v float32)`

SetStatusCode sets StatusCode field to given value.


### GetError

`func (o *Model401Response) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *Model401Response) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *Model401Response) SetError(v string)`

SetError sets Error field to given value.


### GetMessage

`func (o *Model401Response) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *Model401Response) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *Model401Response) SetMessage(v string)`

SetMessage sets Message field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


