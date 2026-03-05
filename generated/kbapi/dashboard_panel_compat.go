// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kbapi

import "encoding/json"

// DashboardPanelItem_Config is a compatibility wrapper retained for
// dashboard provider code that previously worked with generic panel config
// unions in the generated spec.
type DashboardPanelItem_Config struct {
	union json.RawMessage
}

func (t DashboardPanelItem_Config) AsDashboardPanelItemConfig8() (map[string]any, error) {
	var body map[string]any
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DashboardPanelItem_Config) FromDashboardPanelItemConfig8(v map[string]any) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DashboardPanelItem_Config) AsDashboardPanelItemConfig4() (DashboardPanelItemConfig4, error) {
	var body DashboardPanelItemConfig4
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DashboardPanelItem_Config) FromDashboardPanelItemConfig4(v DashboardPanelItemConfig4) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DashboardPanelItem_Config) AsDashboardPanelItemConfig7() (DashboardPanelItemConfig7, error) {
	var body DashboardPanelItemConfig7
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DashboardPanelItem_Config) FromDashboardPanelItemConfig7(v DashboardPanelItemConfig7) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DashboardPanelItem_Config) MarshalJSON() ([]byte, error) {
	return t.union.MarshalJSON()
}

func (t *DashboardPanelItem_Config) UnmarshalJSON(b []byte) error {
	return t.union.UnmarshalJSON(b)
}

type DashboardPanelItemConfig4 struct {
	union json.RawMessage
}

func (t DashboardPanelItemConfig4) AsDashboardPanelItemConfig40() (DashboardPanelItemConfig40, error) {
	var body DashboardPanelItemConfig40
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DashboardPanelItemConfig4) FromDashboardPanelItemConfig40(v DashboardPanelItemConfig40) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DashboardPanelItemConfig4) MarshalJSON() ([]byte, error) {
	return t.union.MarshalJSON()
}

func (t *DashboardPanelItemConfig4) UnmarshalJSON(b []byte) error {
	return t.union.UnmarshalJSON(b)
}

// DashboardPanelItemConfig40 is retained for markdown panel conversion tests.
type DashboardPanelItemConfig40 struct {
	Content     string  `json:"content"`
	Description *string `json:"description,omitempty"`
	HideTitle   *bool   `json:"hide_title,omitempty"`
	Title       *string `json:"title,omitempty"`
}

type DashboardPanelItemConfig70Attributes0 = KbnDashboardPanelLensConfig0Attributes0

type DashboardPanelItem_Config_7_0_Attributes struct {
	union json.RawMessage
}

func (t DashboardPanelItem_Config_7_0_Attributes) AsDashboardPanelItemConfig70Attributes0() (DashboardPanelItemConfig70Attributes0, error) {
	var body DashboardPanelItemConfig70Attributes0
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DashboardPanelItem_Config_7_0_Attributes) FromDashboardPanelItemConfig70Attributes0(v DashboardPanelItemConfig70Attributes0) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DashboardPanelItem_Config_7_0_Attributes) MarshalJSON() ([]byte, error) {
	return t.union.MarshalJSON()
}

func (t *DashboardPanelItem_Config_7_0_Attributes) UnmarshalJSON(b []byte) error {
	return t.union.UnmarshalJSON(b)
}

type DashboardPanelItemConfig70 struct {
	Attributes DashboardPanelItem_Config_7_0_Attributes `json:"attributes"`
}

type DashboardPanelItemConfig7 struct {
	union json.RawMessage
}

func (t DashboardPanelItemConfig7) AsDashboardPanelItemConfig70() (DashboardPanelItemConfig70, error) {
	var body DashboardPanelItemConfig70
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *DashboardPanelItemConfig7) FromDashboardPanelItemConfig70(v DashboardPanelItemConfig70) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t DashboardPanelItemConfig7) MarshalJSON() ([]byte, error) {
	return t.union.MarshalJSON()
}

func (t *DashboardPanelItemConfig7) UnmarshalJSON(b []byte) error {
	return t.union.UnmarshalJSON(b)
}
