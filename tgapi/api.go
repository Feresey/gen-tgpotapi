package tgapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

const APIEndpoint = "https://api.telegram.org"
const FileEndpoint = "https://api.telegram.org/file"

type Response struct {
	OK          bool               `json:"ok"`
	Result      json.RawMessage    `json:"result"`
	Description string             `json:"description,omitempty"`
	Parameters  ResponseParameters `json:"parameters,omitempty"`
	ErrorCode   int                `json:"error_code,omitempty"`
}

type Error struct {
	Code    int
	Message string
	ResponseParameters
}

func (e Error) Error() string {
	return fmt.Sprintf(
		"api response. code: %d, message: %s, parameters: %v",
		e.Code, e.Message, e.ResponseParameters)
}

type API struct {
	cli          *http.Client
	endpoint     string
	fileEndpoint string
}

func NewWithEndpointAndClient(token, endpoint, fileEndpoint string, cli *http.Client) *API {
	return &API{
		cli:          cli,
		endpoint:     fmt.Sprintf("%s/bot%s", endpoint, token),
		fileEndpoint: fmt.Sprintf("%s/bot%s", fileEndpoint, token),
	}
}

func New(token string) *API {
	return NewWithEndpointAndClient(token, APIEndpoint, FileEndpoint, http.DefaultClient)
}

func decodeApiResponse(r io.ReadCloser) (*Response, error) {
	defer r.Close()
	var apiResp Response
	if err := json.NewDecoder(r).Decode(&apiResp); err != nil {
		return &apiResp, err
	}

	if !apiResp.OK {
		return &apiResp, Error{
			Code:               apiResp.ErrorCode,
			Message:            apiResp.Description,
			ResponseParameters: apiResp.Parameters,
		}
	}

	return &apiResp, nil
}

// MakeRequest makes a request to a specific endpoint with our token.
func (api *API) MakeRequest(ctx context.Context, method string, data interface{}) (*Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", api.endpoint, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.cli.Do(req)
	if err != nil {
		return nil, err
	}
	return decodeApiResponse(resp.Body)
}

func (api *API) UploadFile(
	ctx context.Context,
	values url.Values,
	filetype string,
	file *InputFile,
) (*Response, error) {
	b := new(bytes.Buffer)
	w := multipart.NewWriter(b)
	defer w.Close()

	wr, err := w.CreateFormFile(filetype, file.Name)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(wr, file.Reader)
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		api.endpoint,
		b,
	)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = values.Encode()
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := api.cli.Do(req)
	if err != nil {
		return nil, err
	}

	return decodeApiResponse(resp.Body)
}
