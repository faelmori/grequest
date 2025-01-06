package grequest

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

// Init get http method
func Get(u string) *HTTPClient {
	return New().Get(u)
}

// Init Post http method
func Post(u string) *HTTPClient {
	return New().Post(u)
}

// Init Put http method
func Put(u string) *HTTPClient {
	return New().Put(u)
}

// Init Path http method
func Patch(u string) *HTTPClient {
	return New().Patch(u)
}

// Init Delete http method
func Delete(u string) *HTTPClient {
	return New().Delete(u)
}

// Init Head http method
func Head(u string) *HTTPClient {
	return New().Head(u)
}

func (c *HTTPClient) SetMethod(method string) *HTTPClient {
	c.request.method = method
	return c
}

func (c *HTTPClient) Get(u string) *HTTPClient {
	c.request.method = MethodGet
	c.request.url = u
	return c
}

func (c *HTTPClient) Post(u string) *HTTPClient {
	c.request.method = MethodPost
	c.request.url = u
	return c
}

func (c *HTTPClient) Put(u string) *HTTPClient {
	c.request.method = MethodPut
	c.request.url = u
	return c
}

func (c *HTTPClient) Patch(u string) *HTTPClient {
	c.request.method = MethodPatch
	c.request.url = u
	return c
}

func (c *HTTPClient) Delete(u string) *HTTPClient {
	c.request.method = MethodDelete
	c.request.url = u
	return c
}

func (c *HTTPClient) Head(u string) *HTTPClient {
	c.request.method = MethodHead
	c.request.url = u
	return c
}
