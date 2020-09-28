package sirius

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

const ErrUnauthorized ClientError = "unauthorized"

type ClientError string

func (e ClientError) Error() string {
	return string(e)
}

type ValidationErrors map[string]map[string]string

type ValidationError struct {
	Message string
	Errors  ValidationErrors
}

func (ve ValidationError) Error() string {
	return ve.Message
}

func NewClient(httpClient *http.Client, baseURL string) (*Client, error) {
	return &Client{
		http:    httpClient,
		baseURL: baseURL,
	}, nil
}

type Client struct {
	http    *http.Client
	baseURL string
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader, cookies []*http.Cookie) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	var xsrfToken string
	for _, c := range cookies {
		req.AddCookie(c)
		if c.Name == "XSRF-TOKEN" {
			xsrfToken = c.Value
		}
	}

	headerToken, err := url.QueryUnescape(xsrfToken)

	if err != nil {
		return nil, ErrUnauthorized
	}

	req.Header.Add("OPG-Bypass-Membrane", "1")
	req.Header.Add("X-XSRF-TOKEN", headerToken)

	return req, err
}
