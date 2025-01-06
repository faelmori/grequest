package grequest

import (
	"fmt"
	"net/http"
)

type Header struct {
	Client *HTTPClient
}

// Sets the header with io.Reader
func (c *HTTPClient) Header() *Header {
	return &Header{Client: c}
}

// Gets http client object
func (c *Header) client() *HTTPClient {
	return c.Client
}

// Gets raw http.Header from body
func (c *Header) get() *http.Header {
	return &c.client().response().Header
}

// Gets header and convert in string map
func (c *Header) GetWithStringMap() map[string]string {
	var headersMap = make(map[string]string)
	headers := *c.get()
	for key, values := range headers {
		for _, value := range values {
			headersMap[fmt.Sprintf("%s", key)] = fmt.Sprintf("%s", value)
		}
	}
	return headersMap
}

// Gets header string by key
func (c *Header) Get(key string) string {
	return c.client().res.Header.Get(key)
}

// Delete header by key
func (c *Header) Del(key string) *HTTPClient {
	c.Client.request.header.Del(key)
	return c.client()
}

// Sets header key with string value
func (c *Header) Set(key string, value string) *HTTPClient {
	c.Client.request.header.Add(key, value)
	return c.client()
}

// Sets headers from string map object
func (c *Header) Add(header *map[string]string) *HTTPClient {
	for k, val := range *header {
		c.Client.request.header.Add(k, val)
	}
	return c.client()
}
