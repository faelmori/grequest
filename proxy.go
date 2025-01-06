package grequest

import (
	"net/http"
)

type Proxy struct {
	Client *HTTPClient
}

func (c *HTTPClient) Proxy() *Proxy {
	return &Proxy{Client: c}
}

func (c *Proxy) SetProxy(uri string) *HTTPClient {
	u, err := checkURL(uri)
	if err != nil {
		c.Client.errs = append(c.Client.errs, err)
		return c.Client
	}
	c.Client.transport.Proxy = http.ProxyURL(u)
	c.Client.client.Transport = c.Client.transport
	return c.Client
}
