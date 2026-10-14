package kbapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Label struct {
	Key   string  `json:"key"`
	Value *string `json:"value,omitempty"`
}

type Settings struct {
	Mode    *string `json:"mode,omitempty"`
	Retries *int    `json:"retries,omitempty"`
}

type SourceA struct {
	A *string `json:"a,omitempty"`
}

type SourceB struct {
	B *string `json:"b,omitempty"`
}

type Source struct {
	union json.RawMessage
}

func (t Source) AsSourceA() (SourceA, error) {
	var body SourceA
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *Source) FromSourceA(v SourceA) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

func (t Source) AsSourceB() (SourceB, error) {
	var body SourceB
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t *Source) FromSourceB(v SourceB) error {
	b, err := json.Marshal(v)
	t.union = b
	return err
}

type Widget struct {
	Config      *map[string]any `json:"config,omitempty"`
	Description *string         `json:"description,omitempty"`
	Id          *string         `json:"id,omitempty"`
	Labels      *[]Label        `json:"labels,omitempty"`
	Legacy      *string         `json:"legacy,omitempty"`
	Name        string          `json:"name"`
	Settings    *Settings       `json:"settings,omitempty"`
	Source      *Source         `json:"source,omitempty"`
}

type CreateWidgetParams struct {
	DryRun  *bool `form:"dry_run,omitempty" json:"dry_run,omitempty"`
	Verbose *bool `form:"verbose,omitempty" json:"verbose,omitempty"`
}

type CreateWidgetJSONRequestBody = Widget

type BlobText struct {
	Text *string `json:"text,omitempty"`
}

type BlobNumber struct {
	Number *int `json:"number,omitempty"`
}

type CreateBlobJSONBody struct {
	union json.RawMessage
}

func (t CreateBlobJSONBody) AsBlobText() (BlobText, error) {
	var body BlobText
	err := json.Unmarshal(t.union, &body)
	return body, err
}

func (t CreateBlobJSONBody) AsBlobNumber() (BlobNumber, error) {
	var body BlobNumber
	err := json.Unmarshal(t.union, &body)
	return body, err
}

type CreateBlobJSONRequestBody CreateBlobJSONBody

type CreateBlobResponse struct {
	Body []byte
}

type GetWidgetResponse struct {
	Body []byte
}

type CreateWidgetResponse struct {
	Body []byte
}

type Client struct {
	Server string
	Client *http.Client
}

type ClientInterface interface {
	CreateWidgetWithBody(ctx context.Context, params *CreateWidgetParams, contentType string, body io.Reader) (*http.Response, error)
	CreateBlob(ctx context.Context, body CreateBlobJSONRequestBody) (*http.Response, error)
	CreateWidget(ctx context.Context, params *CreateWidgetParams, body CreateWidgetJSONRequestBody) (*http.Response, error)
	GetWidget(ctx context.Context, id string) (*http.Response, error)
}

func (c *Client) CreateWidget(ctx context.Context, params *CreateWidgetParams, body CreateWidgetJSONRequestBody) (*http.Response, error) {
	req, err := NewCreateWidgetRequest(c.Server, params, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

func (c *Client) GetWidget(ctx context.Context, id string) (*http.Response, error) {
	req, err := NewGetWidgetRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

func (c *Client) CreateBlob(ctx context.Context, body CreateBlobJSONRequestBody) (*http.Response, error) {
	req, err := NewCreateBlobRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

func (c *Client) CreateWidgetWithBody(ctx context.Context, params *CreateWidgetParams, contentType string, body io.Reader) (*http.Response, error) {
	req, err := NewCreateWidgetRequestWithBody(c.Server, params, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Client.Do(req)
}

type ClientWithResponses struct {
	ClientInterface
}

func (c *ClientWithResponses) CreateWidgetWithResponse(ctx context.Context, params *CreateWidgetParams, body CreateWidgetJSONRequestBody) (*CreateWidgetResponse, error) {
	rsp, err := c.CreateWidget(ctx, params, body)
	if err != nil {
		return nil, err
	}
	return &CreateWidgetResponse{Body: nil}, rsp.Body.Close()
}

func (c *ClientWithResponses) GetWidgetWithResponse(ctx context.Context, id string) (*GetWidgetResponse, error) {
	rsp, err := c.GetWidget(ctx, id)
	if err != nil {
		return nil, err
	}
	return &GetWidgetResponse{Body: nil}, rsp.Body.Close()
}

func NewCreateWidgetRequest(server string, params *CreateWidgetParams, body CreateWidgetJSONRequestBody) (*http.Request, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return NewCreateWidgetRequestWithBody(server, params, "application/json", bytes.NewReader(buf))
}

func NewCreateWidgetRequestWithBody(server string, _ *CreateWidgetParams, _ string, body io.Reader) (*http.Request, error) {
	operationPath := fmt.Sprintf("/api/widgets")
	return http.NewRequest(http.MethodPost, server+operationPath, body)
}

func (c *ClientWithResponses) CreateBlobWithResponse(ctx context.Context, body CreateBlobJSONRequestBody) (*CreateBlobResponse, error) {
	rsp, err := c.CreateBlob(ctx, body)
	if err != nil {
		return nil, err
	}
	return &CreateBlobResponse{Body: nil}, rsp.Body.Close()
}

func (c *ClientWithResponses) CreateWidgetWithBodyWithResponse(ctx context.Context, params *CreateWidgetParams, contentType string, body io.Reader) (*CreateWidgetResponse, error) {
	rsp, err := c.CreateWidgetWithBody(ctx, params, contentType, body)
	if err != nil {
		return nil, err
	}
	return &CreateWidgetResponse{Body: nil}, rsp.Body.Close()
}

func NewCreateBlobRequest(server string, body CreateBlobJSONRequestBody) (*http.Request, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return NewCreateBlobRequestWithBody(server, "application/json", bytes.NewReader(buf))
}

func NewCreateBlobRequestWithBody(server string, _ string, body io.Reader) (*http.Request, error) {
	operationPath := fmt.Sprintf("/api/blobs")
	return http.NewRequest(http.MethodPost, server+operationPath, body)
}

func NewGetWidgetRequest(server string, id string) (*http.Request, error) {
	operationPath := fmt.Sprintf("/api/widgets/%s", id)
	return http.NewRequest(http.MethodGet, server+operationPath, nil)
}
