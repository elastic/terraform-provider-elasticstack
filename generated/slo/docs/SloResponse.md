# SloResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | The identifier of the SLO. | [optional] 
**Name** | Pointer to **string** | The name of the SLO. | [optional] 
**Description** | Pointer to **string** | The description of the SLO. | [optional] 
**Indicator** | Pointer to [**SloResponseIndicator**](SloResponseIndicator.md) |  | [optional] 
**TimeWindow** | Pointer to [**SloResponseTimeWindow**](SloResponseTimeWindow.md) |  | [optional] 
**BudgetingMethod** | Pointer to [**BudgetingMethod**](BudgetingMethod.md) |  | [optional] 
**Objective** | Pointer to [**Objective**](Objective.md) |  | [optional] 
**Settings** | Pointer to [**Settings**](Settings.md) |  | [optional] 
**Revision** | Pointer to **float32** | The SLO revision | [optional] 
**Summary** | Pointer to [**Summary**](Summary.md) |  | [optional] 
**Enabled** | Pointer to **bool** | Indicate if the SLO is enabled | [optional] 
**CreatedAt** | Pointer to **string** | The creation date | [optional] 
**UpdatedAt** | Pointer to **string** | The last update date | [optional] 

## Methods

### NewSloResponse

`func NewSloResponse() *SloResponse`

NewSloResponse instantiates a new SloResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSloResponseWithDefaults

`func NewSloResponseWithDefaults() *SloResponse`

NewSloResponseWithDefaults instantiates a new SloResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *SloResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SloResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SloResponse) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *SloResponse) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *SloResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SloResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SloResponse) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *SloResponse) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDescription

`func (o *SloResponse) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *SloResponse) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *SloResponse) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *SloResponse) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetIndicator

`func (o *SloResponse) GetIndicator() SloResponseIndicator`

GetIndicator returns the Indicator field if non-nil, zero value otherwise.

### GetIndicatorOk

`func (o *SloResponse) GetIndicatorOk() (*SloResponseIndicator, bool)`

GetIndicatorOk returns a tuple with the Indicator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndicator

`func (o *SloResponse) SetIndicator(v SloResponseIndicator)`

SetIndicator sets Indicator field to given value.

### HasIndicator

`func (o *SloResponse) HasIndicator() bool`

HasIndicator returns a boolean if a field has been set.

### GetTimeWindow

`func (o *SloResponse) GetTimeWindow() SloResponseTimeWindow`

GetTimeWindow returns the TimeWindow field if non-nil, zero value otherwise.

### GetTimeWindowOk

`func (o *SloResponse) GetTimeWindowOk() (*SloResponseTimeWindow, bool)`

GetTimeWindowOk returns a tuple with the TimeWindow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeWindow

`func (o *SloResponse) SetTimeWindow(v SloResponseTimeWindow)`

SetTimeWindow sets TimeWindow field to given value.

### HasTimeWindow

`func (o *SloResponse) HasTimeWindow() bool`

HasTimeWindow returns a boolean if a field has been set.

### GetBudgetingMethod

`func (o *SloResponse) GetBudgetingMethod() BudgetingMethod`

GetBudgetingMethod returns the BudgetingMethod field if non-nil, zero value otherwise.

### GetBudgetingMethodOk

`func (o *SloResponse) GetBudgetingMethodOk() (*BudgetingMethod, bool)`

GetBudgetingMethodOk returns a tuple with the BudgetingMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBudgetingMethod

`func (o *SloResponse) SetBudgetingMethod(v BudgetingMethod)`

SetBudgetingMethod sets BudgetingMethod field to given value.

### HasBudgetingMethod

`func (o *SloResponse) HasBudgetingMethod() bool`

HasBudgetingMethod returns a boolean if a field has been set.

### GetObjective

`func (o *SloResponse) GetObjective() Objective`

GetObjective returns the Objective field if non-nil, zero value otherwise.

### GetObjectiveOk

`func (o *SloResponse) GetObjectiveOk() (*Objective, bool)`

GetObjectiveOk returns a tuple with the Objective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjective

`func (o *SloResponse) SetObjective(v Objective)`

SetObjective sets Objective field to given value.

### HasObjective

`func (o *SloResponse) HasObjective() bool`

HasObjective returns a boolean if a field has been set.

### GetSettings

`func (o *SloResponse) GetSettings() Settings`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *SloResponse) GetSettingsOk() (*Settings, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *SloResponse) SetSettings(v Settings)`

SetSettings sets Settings field to given value.

### HasSettings

`func (o *SloResponse) HasSettings() bool`

HasSettings returns a boolean if a field has been set.

### GetRevision

`func (o *SloResponse) GetRevision() float32`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *SloResponse) GetRevisionOk() (*float32, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *SloResponse) SetRevision(v float32)`

SetRevision sets Revision field to given value.

### HasRevision

`func (o *SloResponse) HasRevision() bool`

HasRevision returns a boolean if a field has been set.

### GetSummary

`func (o *SloResponse) GetSummary() Summary`

GetSummary returns the Summary field if non-nil, zero value otherwise.

### GetSummaryOk

`func (o *SloResponse) GetSummaryOk() (*Summary, bool)`

GetSummaryOk returns a tuple with the Summary field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSummary

`func (o *SloResponse) SetSummary(v Summary)`

SetSummary sets Summary field to given value.

### HasSummary

`func (o *SloResponse) HasSummary() bool`

HasSummary returns a boolean if a field has been set.

### GetEnabled

`func (o *SloResponse) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *SloResponse) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *SloResponse) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *SloResponse) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetCreatedAt

`func (o *SloResponse) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *SloResponse) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *SloResponse) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *SloResponse) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *SloResponse) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *SloResponse) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *SloResponse) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *SloResponse) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


