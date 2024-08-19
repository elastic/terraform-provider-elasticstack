# ActionsInnerAlertsFilterQuery

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Kql** | Pointer to **string** | A filter written in Kibana Query Language (KQL). | [optional] 
**Filters** | Pointer to [**[]Filter**](Filter.md) |  | [optional] 

## Methods

### NewActionsInnerAlertsFilterQuery

`func NewActionsInnerAlertsFilterQuery() *ActionsInnerAlertsFilterQuery`

NewActionsInnerAlertsFilterQuery instantiates a new ActionsInnerAlertsFilterQuery object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewActionsInnerAlertsFilterQueryWithDefaults

`func NewActionsInnerAlertsFilterQueryWithDefaults() *ActionsInnerAlertsFilterQuery`

NewActionsInnerAlertsFilterQueryWithDefaults instantiates a new ActionsInnerAlertsFilterQuery object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKql

`func (o *ActionsInnerAlertsFilterQuery) GetKql() string`

GetKql returns the Kql field if non-nil, zero value otherwise.

### GetKqlOk

`func (o *ActionsInnerAlertsFilterQuery) GetKqlOk() (*string, bool)`

GetKqlOk returns a tuple with the Kql field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKql

`func (o *ActionsInnerAlertsFilterQuery) SetKql(v string)`

SetKql sets Kql field to given value.

### HasKql

`func (o *ActionsInnerAlertsFilterQuery) HasKql() bool`

HasKql returns a boolean if a field has been set.

### GetFilters

`func (o *ActionsInnerAlertsFilterQuery) GetFilters() []Filter`

GetFilters returns the Filters field if non-nil, zero value otherwise.

### GetFiltersOk

`func (o *ActionsInnerAlertsFilterQuery) GetFiltersOk() (*[]Filter, bool)`

GetFiltersOk returns a tuple with the Filters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilters

`func (o *ActionsInnerAlertsFilterQuery) SetFilters(v []Filter)`

SetFilters sets Filters field to given value.

### HasFilters

`func (o *ActionsInnerAlertsFilterQuery) HasFilters() bool`

HasFilters returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


