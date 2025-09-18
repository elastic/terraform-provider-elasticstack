# BulkDeleteStatusResponseResultsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | The ID of the SLO that was deleted | [optional] 
**Success** | Pointer to **bool** | The result of the deletion operation for this SLO | [optional] 
**Error** | Pointer to **string** | The error message if the deletion operation failed for this SLO | [optional] 

## Methods

### NewBulkDeleteStatusResponseResultsInner

`func NewBulkDeleteStatusResponseResultsInner() *BulkDeleteStatusResponseResultsInner`

NewBulkDeleteStatusResponseResultsInner instantiates a new BulkDeleteStatusResponseResultsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBulkDeleteStatusResponseResultsInnerWithDefaults

`func NewBulkDeleteStatusResponseResultsInnerWithDefaults() *BulkDeleteStatusResponseResultsInner`

NewBulkDeleteStatusResponseResultsInnerWithDefaults instantiates a new BulkDeleteStatusResponseResultsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *BulkDeleteStatusResponseResultsInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *BulkDeleteStatusResponseResultsInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *BulkDeleteStatusResponseResultsInner) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *BulkDeleteStatusResponseResultsInner) HasId() bool`

HasId returns a boolean if a field has been set.

### GetSuccess

`func (o *BulkDeleteStatusResponseResultsInner) GetSuccess() bool`

GetSuccess returns the Success field if non-nil, zero value otherwise.

### GetSuccessOk

`func (o *BulkDeleteStatusResponseResultsInner) GetSuccessOk() (*bool, bool)`

GetSuccessOk returns a tuple with the Success field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuccess

`func (o *BulkDeleteStatusResponseResultsInner) SetSuccess(v bool)`

SetSuccess sets Success field to given value.

### HasSuccess

`func (o *BulkDeleteStatusResponseResultsInner) HasSuccess() bool`

HasSuccess returns a boolean if a field has been set.

### GetError

`func (o *BulkDeleteStatusResponseResultsInner) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *BulkDeleteStatusResponseResultsInner) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *BulkDeleteStatusResponseResultsInner) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *BulkDeleteStatusResponseResultsInner) HasError() bool`

HasError returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


