package grequest

type Status struct {
	Client *HTTPClient
}

// Init status
func (c *HTTPClient) Status() *Status {
	return &Status{Client: c}
}

// Gets http client object
func (c *Status) client() *HTTPClient {
	return c.Client
}

// Get status code with string like 200=OK
func (c *Status) Get() string {
	return c.client().response().Status
}

// Get status code with int
func (c *Status) GetCode() int {
	return c.client().response().StatusCode
}
