# BulkDeleteStatusResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IsDone** | Pointer to **bool** | Indicates if the bulk deletion operation is completed | [optional] 
**Error** | Pointer to **string** | The error message if the bulk deletion operation failed | [optional] 
**Results** | Pointer to [**[]BulkDeleteStatusResponseResultsInner**](BulkDeleteStatusResponseResultsInner.md) | The results of the bulk deletion operation, including the success status and any errors for each SLO | [optional] 

## Methods

### NewBulkDeleteStatusResponse

`func NewBulkDeleteStatusResponse() *BulkDeleteStatusResponse`

NewBulkDeleteStatusResponse instantiates a new BulkDeleteStatusResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBulkDeleteStatusResponseWithDefaults

`func NewBulkDeleteStatusResponseWithDefaults() *BulkDeleteStatusResponse`

NewBulkDeleteStatusResponseWithDefaults instantiates a new BulkDeleteStatusResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIsDone

`func (o *BulkDeleteStatusResponse) GetIsDone() bool`

GetIsDone returns the IsDone field if non-nil, zero value otherwise.

### GetIsDoneOk

`func (o *BulkDeleteStatusResponse) GetIsDoneOk() (*bool, bool)`

GetIsDoneOk returns a tuple with the IsDone field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDone

`func (o *BulkDeleteStatusResponse) SetIsDone(v bool)`

SetIsDone sets IsDone field to given value.

### HasIsDone

`func (o *BulkDeleteStatusResponse) HasIsDone() bool`

HasIsDone returns a boolean if a field has been set.

### GetError

`func (o *BulkDeleteStatusResponse) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *BulkDeleteStatusResponse) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *BulkDeleteStatusResponse) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *BulkDeleteStatusResponse) HasError() bool`

HasError returns a boolean if a field has been set.

### GetResults

`func (o *BulkDeleteStatusResponse) GetResults() []BulkDeleteStatusResponseResultsInner`

GetResults returns the Results field if non-nil, zero value otherwise.

### GetResultsOk

`func (o *BulkDeleteStatusResponse) GetResultsOk() (*[]BulkDeleteStatusResponseResultsInner, bool)`

GetResultsOk returns a tuple with the Results field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResults

`func (o *BulkDeleteStatusResponse) SetResults(v []BulkDeleteStatusResponseResultsInner)`

SetResults sets Results field to given value.

### HasResults

`func (o *BulkDeleteStatusResponse) HasResults() bool`

HasResults returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


