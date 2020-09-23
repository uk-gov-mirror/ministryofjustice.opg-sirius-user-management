package sirius

import (
	"net/http"
	"net/url"
)

const ErrUnauthorized ClientError = "unauthorized"

type ClientError string

func (e ClientError) Error() string {
	return string(e)
}

func NewClient(httpClient *http.Client, baseURL string) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		http:    httpClient,
		baseURL: parsed,
	}, nil
}

type Client struct {
	http    *http.Client
	baseURL *url.URL
}

func (c *Client) url(path string) string {
	partial, _ := url.Parse(path)

	return c.baseURL.ResolveReference(partial).String()
}
