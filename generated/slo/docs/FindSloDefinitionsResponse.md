# FindSloDefinitionsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Page** | Pointer to **float64** | for backward compability | [optional] [default to 1]
**PerPage** | Pointer to **float64** | for backward compability | [optional] 
**Total** | Pointer to **float64** |  | [optional] 
**Results** | Pointer to [**[]SloWithSummaryResponse**](SloWithSummaryResponse.md) |  | [optional] 
**Size** | Pointer to **float64** |  | [optional] 
**SearchAfter** | Pointer to **[]string** | the cursor to provide to get the next paged results | [optional] 

## Methods

### NewFindSloDefinitionsResponse

`func NewFindSloDefinitionsResponse() *FindSloDefinitionsResponse`

NewFindSloDefinitionsResponse instantiates a new FindSloDefinitionsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFindSloDefinitionsResponseWithDefaults

`func NewFindSloDefinitionsResponseWithDefaults() *FindSloDefinitionsResponse`

NewFindSloDefinitionsResponseWithDefaults instantiates a new FindSloDefinitionsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPage

`func (o *FindSloDefinitionsResponse) GetPage() float64`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *FindSloDefinitionsResponse) GetPageOk() (*float64, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *FindSloDefinitionsResponse) SetPage(v float64)`

SetPage sets Page field to given value.

### HasPage

`func (o *FindSloDefinitionsResponse) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetPerPage

`func (o *FindSloDefinitionsResponse) GetPerPage() float64`

GetPerPage returns the PerPage field if non-nil, zero value otherwise.

### GetPerPageOk

`func (o *FindSloDefinitionsResponse) GetPerPageOk() (*float64, bool)`

GetPerPageOk returns a tuple with the PerPage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPerPage

`func (o *FindSloDefinitionsResponse) SetPerPage(v float64)`

SetPerPage sets PerPage field to given value.

### HasPerPage

`func (o *FindSloDefinitionsResponse) HasPerPage() bool`

HasPerPage returns a boolean if a field has been set.

### GetTotal

`func (o *FindSloDefinitionsResponse) GetTotal() float64`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *FindSloDefinitionsResponse) GetTotalOk() (*float64, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *FindSloDefinitionsResponse) SetTotal(v float64)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *FindSloDefinitionsResponse) HasTotal() bool`

HasTotal returns a boolean if a field has been set.

### GetResults

`func (o *FindSloDefinitionsResponse) GetResults() []SloWithSummaryResponse`

GetResults returns the Results field if non-nil, zero value otherwise.

### GetResultsOk

`func (o *FindSloDefinitionsResponse) GetResultsOk() (*[]SloWithSummaryResponse, bool)`

GetResultsOk returns a tuple with the Results field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResults

`func (o *FindSloDefinitionsResponse) SetResults(v []SloWithSummaryResponse)`

SetResults sets Results field to given value.

### HasResults

`func (o *FindSloDefinitionsResponse) HasResults() bool`

HasResults returns a boolean if a field has been set.

### GetSize

`func (o *FindSloDefinitionsResponse) GetSize() float64`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *FindSloDefinitionsResponse) GetSizeOk() (*float64, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *FindSloDefinitionsResponse) SetSize(v float64)`

SetSize sets Size field to given value.

### HasSize

`func (o *FindSloDefinitionsResponse) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetSearchAfter

`func (o *FindSloDefinitionsResponse) GetSearchAfter() []string`

GetSearchAfter returns the SearchAfter field if non-nil, zero value otherwise.

### GetSearchAfterOk

`func (o *FindSloDefinitionsResponse) GetSearchAfterOk() (*[]string, bool)`

GetSearchAfterOk returns a tuple with the SearchAfter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSearchAfter

`func (o *FindSloDefinitionsResponse) SetSearchAfter(v []string)`

SetSearchAfter sets SearchAfter field to given value.

### HasSearchAfter

`func (o *FindSloDefinitionsResponse) HasSearchAfter() bool`

HasSearchAfter returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


