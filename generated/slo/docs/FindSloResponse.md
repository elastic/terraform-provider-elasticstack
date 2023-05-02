# FindSloResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Page** | Pointer to **float32** |  | [optional] 
**PerPage** | Pointer to **float32** |  | [optional] 
**Total** | Pointer to **float32** |  | [optional] 
**Results** | Pointer to [**[]SloResponse**](SloResponse.md) |  | [optional] 

## Methods

### NewFindSloResponse

`func NewFindSloResponse() *FindSloResponse`

NewFindSloResponse instantiates a new FindSloResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFindSloResponseWithDefaults

`func NewFindSloResponseWithDefaults() *FindSloResponse`

NewFindSloResponseWithDefaults instantiates a new FindSloResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPage

`func (o *FindSloResponse) GetPage() float32`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *FindSloResponse) GetPageOk() (*float32, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *FindSloResponse) SetPage(v float32)`

SetPage sets Page field to given value.

### HasPage

`func (o *FindSloResponse) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetPerPage

`func (o *FindSloResponse) GetPerPage() float32`

GetPerPage returns the PerPage field if non-nil, zero value otherwise.

### GetPerPageOk

`func (o *FindSloResponse) GetPerPageOk() (*float32, bool)`

GetPerPageOk returns a tuple with the PerPage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPerPage

`func (o *FindSloResponse) SetPerPage(v float32)`

SetPerPage sets PerPage field to given value.

### HasPerPage

`func (o *FindSloResponse) HasPerPage() bool`

HasPerPage returns a boolean if a field has been set.

### GetTotal

`func (o *FindSloResponse) GetTotal() float32`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *FindSloResponse) GetTotalOk() (*float32, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *FindSloResponse) SetTotal(v float32)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *FindSloResponse) HasTotal() bool`

HasTotal returns a boolean if a field has been set.

### GetResults

`func (o *FindSloResponse) GetResults() []SloResponse`

GetResults returns the Results field if non-nil, zero value otherwise.

### GetResultsOk

`func (o *FindSloResponse) GetResultsOk() (*[]SloResponse, bool)`

GetResultsOk returns a tuple with the Results field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResults

`func (o *FindSloResponse) SetResults(v []SloResponse)`

SetResults sets Results field to given value.

### HasResults

`func (o *FindSloResponse) HasResults() bool`

HasResults returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


