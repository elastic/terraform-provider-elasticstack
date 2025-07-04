# FindSloResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Size** | Pointer to **float64** | Size provided for cursor based pagination | [optional] 
**SearchAfter** | Pointer to **string** |  | [optional] 
**Page** | Pointer to **float64** |  | [optional] 
**PerPage** | Pointer to **float64** |  | [optional] 
**Total** | Pointer to **float64** |  | [optional] 
**Results** | Pointer to [**[]SloWithSummaryResponse**](SloWithSummaryResponse.md) |  | [optional] 

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

### GetSize

`func (o *FindSloResponse) GetSize() float64`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *FindSloResponse) GetSizeOk() (*float64, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *FindSloResponse) SetSize(v float64)`

SetSize sets Size field to given value.

### HasSize

`func (o *FindSloResponse) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetSearchAfter

`func (o *FindSloResponse) GetSearchAfter() string`

GetSearchAfter returns the SearchAfter field if non-nil, zero value otherwise.

### GetSearchAfterOk

`func (o *FindSloResponse) GetSearchAfterOk() (*string, bool)`

GetSearchAfterOk returns a tuple with the SearchAfter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSearchAfter

`func (o *FindSloResponse) SetSearchAfter(v string)`

SetSearchAfter sets SearchAfter field to given value.

### HasSearchAfter

`func (o *FindSloResponse) HasSearchAfter() bool`

HasSearchAfter returns a boolean if a field has been set.

### GetPage

`func (o *FindSloResponse) GetPage() float64`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *FindSloResponse) GetPageOk() (*float64, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *FindSloResponse) SetPage(v float64)`

SetPage sets Page field to given value.

### HasPage

`func (o *FindSloResponse) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetPerPage

`func (o *FindSloResponse) GetPerPage() float64`

GetPerPage returns the PerPage field if non-nil, zero value otherwise.

### GetPerPageOk

`func (o *FindSloResponse) GetPerPageOk() (*float64, bool)`

GetPerPageOk returns a tuple with the PerPage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPerPage

`func (o *FindSloResponse) SetPerPage(v float64)`

SetPerPage sets PerPage field to given value.

### HasPerPage

`func (o *FindSloResponse) HasPerPage() bool`

HasPerPage returns a boolean if a field has been set.

### GetTotal

`func (o *FindSloResponse) GetTotal() float64`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *FindSloResponse) GetTotalOk() (*float64, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *FindSloResponse) SetTotal(v float64)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *FindSloResponse) HasTotal() bool`

HasTotal returns a boolean if a field has been set.

### GetResults

`func (o *FindSloResponse) GetResults() []SloWithSummaryResponse`

GetResults returns the Results field if non-nil, zero value otherwise.

### GetResultsOk

`func (o *FindSloResponse) GetResultsOk() (*[]SloWithSummaryResponse, bool)`

GetResultsOk returns a tuple with the Results field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResults

`func (o *FindSloResponse) SetResults(v []SloWithSummaryResponse)`

SetResults sets Results field to given value.

### HasResults

`func (o *FindSloResponse) HasResults() bool`

HasResults returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


