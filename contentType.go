package grequest

const (
	contentType       = "Content-Type"
	formUrlencoded    = "application/x-www-form-urlencoded"
	multipartFormData = "multipart/form-data"
)

type ContentType struct {
	Client *HTTPClient
}

func (c *HTTPClient) ContentType() *ContentType {
	return &ContentType{Client: c}
}

// Gets http client object
func (c *ContentType) client() *HTTPClient {
	return c.Client
}

// Sets content type in request header
func (c *ContentType) Set(value string) *HTTPClient {
	c.client().Header().Set(contentType, value)
	return c.client()
}

// Sets content type in request header
func (c *ContentType) SetFormUrlencoded() *HTTPClient {
	c.client().Header().Set(contentType, formUrlencoded)
	return c.client()
}

// Sets content type in request header
func (c *ContentType) SetMultipartFormData() *HTTPClient {
	c.client().Header().Set(contentType, multipartFormData)
	return c.client()
}

// Gets content type fron response headers and return as string
func (c *ContentType) Get() string {
	return c.client().Header().Get(contentType)
}
