# CreateSloRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | A optional and unique identifier for the SLO. Must be between 8 and 36 chars | [optional] 
**Name** | **string** | A name for the SLO. | 
**Description** | **string** | A description for the SLO. | 
**Indicator** | [**CreateSloRequestIndicator**](CreateSloRequestIndicator.md) |  | 
**TimeWindow** | [**TimeWindow**](TimeWindow.md) |  | 
**BudgetingMethod** | [**BudgetingMethod**](BudgetingMethod.md) |  | 
**Objective** | [**Objective**](Objective.md) |  | 
**Settings** | Pointer to [**Settings**](Settings.md) |  | [optional] 

## Methods

### NewCreateSloRequest

`func NewCreateSloRequest(name string, description string, indicator CreateSloRequestIndicator, timeWindow TimeWindow, budgetingMethod BudgetingMethod, objective Objective, ) *CreateSloRequest`

NewCreateSloRequest instantiates a new CreateSloRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateSloRequestWithDefaults

`func NewCreateSloRequestWithDefaults() *CreateSloRequest`

NewCreateSloRequestWithDefaults instantiates a new CreateSloRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *CreateSloRequest) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *CreateSloRequest) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *CreateSloRequest) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *CreateSloRequest) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *CreateSloRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateSloRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateSloRequest) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *CreateSloRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *CreateSloRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *CreateSloRequest) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetIndicator

`func (o *CreateSloRequest) GetIndicator() CreateSloRequestIndicator`

GetIndicator returns the Indicator field if non-nil, zero value otherwise.

### GetIndicatorOk

`func (o *CreateSloRequest) GetIndicatorOk() (*CreateSloRequestIndicator, bool)`

GetIndicatorOk returns a tuple with the Indicator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndicator

`func (o *CreateSloRequest) SetIndicator(v CreateSloRequestIndicator)`

SetIndicator sets Indicator field to given value.


### GetTimeWindow

`func (o *CreateSloRequest) GetTimeWindow() TimeWindow`

GetTimeWindow returns the TimeWindow field if non-nil, zero value otherwise.

### GetTimeWindowOk

`func (o *CreateSloRequest) GetTimeWindowOk() (*TimeWindow, bool)`

GetTimeWindowOk returns a tuple with the TimeWindow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeWindow

`func (o *CreateSloRequest) SetTimeWindow(v TimeWindow)`

SetTimeWindow sets TimeWindow field to given value.


### GetBudgetingMethod

`func (o *CreateSloRequest) GetBudgetingMethod() BudgetingMethod`

GetBudgetingMethod returns the BudgetingMethod field if non-nil, zero value otherwise.

### GetBudgetingMethodOk

`func (o *CreateSloRequest) GetBudgetingMethodOk() (*BudgetingMethod, bool)`

GetBudgetingMethodOk returns a tuple with the BudgetingMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBudgetingMethod

`func (o *CreateSloRequest) SetBudgetingMethod(v BudgetingMethod)`

SetBudgetingMethod sets BudgetingMethod field to given value.


### GetObjective

`func (o *CreateSloRequest) GetObjective() Objective`

GetObjective returns the Objective field if non-nil, zero value otherwise.

### GetObjectiveOk

`func (o *CreateSloRequest) GetObjectiveOk() (*Objective, bool)`

GetObjectiveOk returns a tuple with the Objective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjective

`func (o *CreateSloRequest) SetObjective(v Objective)`

SetObjective sets Objective field to given value.


### GetSettings

`func (o *CreateSloRequest) GetSettings() Settings`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *CreateSloRequest) GetSettingsOk() (*Settings, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *CreateSloRequest) SetSettings(v Settings)`

SetSettings sets Settings field to given value.

### HasSettings

`func (o *CreateSloRequest) HasSettings() bool`

HasSettings returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


