package client

import (
	"net/http"
	"net/url"
)

type HttpClient struct {
	client http.Client
	Token string
}

var client *HttpClient

//Singelton
func GetClient(token string) *HttpClient{
	if client == nil{
		client = &HttpClient{Token:token}
	}
	return client
}

//rewrite of the Do method adding the api auth as a header.
func (c *HttpClient) Do(req *http.Request) (resp *http.Response, err error) {
	req.Header.Add("X-API-KEY", c.Token)
	if  req.Method == "PATCH" {
		c.client.Timeout = 0
	}
	return c.client.Do(req)
}

//the get request
func (c *HttpClient) Get(url string) (resp *http.Response, err error) {

	req , err :=http.NewRequest("GET",url,nil)
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
