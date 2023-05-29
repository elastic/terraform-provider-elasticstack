# UpdateSloRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | A name for the SLO. | [optional] 
**Description** | Pointer to **string** | A description for the SLO. | [optional] 
**Indicator** | Pointer to [**CreateSloRequestIndicator**](CreateSloRequestIndicator.md) |  | [optional] 
**TimeWindow** | Pointer to [**SloResponseTimeWindow**](SloResponseTimeWindow.md) |  | [optional] 
**BudgetingMethod** | Pointer to [**BudgetingMethod**](BudgetingMethod.md) |  | [optional] 
**Objective** | Pointer to [**Objective**](Objective.md) |  | [optional] 
**Settings** | Pointer to [**Settings**](Settings.md) |  | [optional] 

## Methods

### NewUpdateSloRequest

`func NewUpdateSloRequest() *UpdateSloRequest`

NewUpdateSloRequest instantiates a new UpdateSloRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateSloRequestWithDefaults

`func NewUpdateSloRequestWithDefaults() *UpdateSloRequest`

NewUpdateSloRequestWithDefaults instantiates a new UpdateSloRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *UpdateSloRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateSloRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateSloRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *UpdateSloRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDescription

`func (o *UpdateSloRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *UpdateSloRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *UpdateSloRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *UpdateSloRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetIndicator

`func (o *UpdateSloRequest) GetIndicator() CreateSloRequestIndicator`

GetIndicator returns the Indicator field if non-nil, zero value otherwise.

### GetIndicatorOk

`func (o *UpdateSloRequest) GetIndicatorOk() (*CreateSloRequestIndicator, bool)`

GetIndicatorOk returns a tuple with the Indicator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndicator

`func (o *UpdateSloRequest) SetIndicator(v CreateSloRequestIndicator)`

SetIndicator sets Indicator field to given value.

### HasIndicator

`func (o *UpdateSloRequest) HasIndicator() bool`

HasIndicator returns a boolean if a field has been set.

### GetTimeWindow

`func (o *UpdateSloRequest) GetTimeWindow() SloResponseTimeWindow`

GetTimeWindow returns the TimeWindow field if non-nil, zero value otherwise.

### GetTimeWindowOk

`func (o *UpdateSloRequest) GetTimeWindowOk() (*SloResponseTimeWindow, bool)`

GetTimeWindowOk returns a tuple with the TimeWindow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeWindow

`func (o *UpdateSloRequest) SetTimeWindow(v SloResponseTimeWindow)`

SetTimeWindow sets TimeWindow field to given value.

### HasTimeWindow

`func (o *UpdateSloRequest) HasTimeWindow() bool`

HasTimeWindow returns a boolean if a field has been set.

### GetBudgetingMethod

`func (o *UpdateSloRequest) GetBudgetingMethod() BudgetingMethod`

GetBudgetingMethod returns the BudgetingMethod field if non-nil, zero value otherwise.

### GetBudgetingMethodOk

`func (o *UpdateSloRequest) GetBudgetingMethodOk() (*BudgetingMethod, bool)`

GetBudgetingMethodOk returns a tuple with the BudgetingMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBudgetingMethod

`func (o *UpdateSloRequest) SetBudgetingMethod(v BudgetingMethod)`

SetBudgetingMethod sets BudgetingMethod field to given value.

### HasBudgetingMethod

`func (o *UpdateSloRequest) HasBudgetingMethod() bool`

HasBudgetingMethod returns a boolean if a field has been set.

### GetObjective

`func (o *UpdateSloRequest) GetObjective() Objective`

GetObjective returns the Objective field if non-nil, zero value otherwise.

### GetObjectiveOk

`func (o *UpdateSloRequest) GetObjectiveOk() (*Objective, bool)`

GetObjectiveOk returns a tuple with the Objective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjective

`func (o *UpdateSloRequest) SetObjective(v Objective)`

SetObjective sets Objective field to given value.

### HasObjective

`func (o *UpdateSloRequest) HasObjective() bool`

HasObjective returns a boolean if a field has been set.

### GetSettings

`func (o *UpdateSloRequest) GetSettings() Settings`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *UpdateSloRequest) GetSettingsOk() (*Settings, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *UpdateSloRequest) SetSettings(v Settings)`

SetSettings sets Settings field to given value.

### HasSettings

`func (o *UpdateSloRequest) HasSettings() bool`

HasSettings returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


