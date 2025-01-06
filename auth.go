package grequest

type BasicAuth struct {
	User string
	Pass string
}

const authorization = "Authorization"

type Auth struct {
	Client *HTTPClient
}

// Init status
func (c *HTTPClient) Auth() *Auth {
	return &Auth{Client: c}
}

// Sets basic auth in request
func (c *Auth) SetBasic(user, pass string) *HTTPClient {
	c.Client.request.basic = &BasicAuth{
		User: user,
		Pass: pass,
	}
	return c.Client
}

// Sets custom auth token in header request
func (c *Auth) SetToken(token string) *HTTPClient {
	c.Client.Header().Set(authorization, token)
	return c.Client
}

// Sets bearer token in header request
func (c *Auth) SetBearer(token string) *HTTPClient {
	c.Client.Header().Set(authorization, "Bearer "+token)
	return c.Client
}
