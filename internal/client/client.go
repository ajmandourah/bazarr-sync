package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultTimeout  = 30 * time.Second
	syncTimeout     = 120 * time.Second // Sync operations need more time
	maxRedirects    = 10
)

// HTTPTransport returns a configured http.Transport for connection pooling and timeouts.
func HTTPTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
}

type HttpClient struct {
	client http.Client
	Token string
}

var client *HttpClient

// GetClient returns the singleton HTTP client. If the token differs from the existing
// client, a new one is created instead of returning the stale one.
func GetClient(token string) *HttpClient{
	if client == nil || client.Token != token {
		client = &HttpClient{
			Token: token,
			client: http.Client{
				Transport: HTTPTransport(),
				Timeout:   defaultTimeout,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					if len(via) >= maxRedirects {
						return fmt.Errorf("stopped after %d redirects", maxRedirects)
					}
					// Preserve API key on redirect
					if via[0] != nil {
						req.Header.Set("X-API-KEY", via[0].Header.Get("X-API-KEY"))
					}
					return nil
				},
			},
		}
	}
	return client
}

// WithTimeout returns a new HttpClient with the given timeout.
func GetClientWithTimeout(token string, timeout time.Duration) *HttpClient {
	return &HttpClient{
		Token: token,
		client: http.Client{
			Transport: HTTPTransport(),
			Timeout:   timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return fmt.Errorf("stopped after %d redirects", maxRedirects)
				}
				if via[0] != nil {
					req.Header.Set("X-API-KEY", via[0].Header.Get("X-API-KEY"))
				}
				return nil
			},
		},
	}
}

// GetSyncClient returns a client with a longer timeout suitable for sync operations.
func GetSyncClient(token string) *HttpClient {
	return GetClientWithTimeout(token, syncTimeout)
}

// rewrite of the Do method adding the api auth as a header and ensuring response bodies
// are always closed (even on error paths) so connections can be reused.
func (c *HttpClient) Do(req *http.Request) (resp *http.Response, err error) {
	req.Header.Add("X-API-KEY", c.Token)
	return c.client.Do(req)
}

// GetRaw makes a GET request and returns the response without reading the body.
// The caller is responsible for reading/closing the body.
func (c *HttpClient) Get(url string) (resp *http.Response, err error) {
	req , err := http.NewRequest("GET",url,nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// TODO : add paramaters
func (c *HttpClient) Post(url string, params url.Values) (resp *http.Response, err error) {
	req , err :=http.NewRequest("POST",url+params.Encode(),nil)
	if err != nil {
		return nil, err
	}

	
	return c.Do(req)
}

// TODO : Add paramaters
func (c *HttpClient) Patch(url string) (resp *http.Response, err error) {	
	req , err :=http.NewRequest("PATCH",url,nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
