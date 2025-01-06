package grequest

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"
)

type HTTPClient struct {
	cacheEnabled    bool
	errors          []error
	maxRedirect     int
	redirectEnabled bool
	Timeout         time.Duration
	userAgent       string
	client          *http.Client
	transport       *http.Transport
	cjar            *cookiejar.Jar
	request         *Request
	res             *http.Response
	BodyBytes       []byte
	errs            []error
}

type Request struct {
	req            *http.Request
	header         http.Header
	body           io.Reader
	method         string
	url            string
	basic          *BasicAuth
	isRequestReady bool
	isRequested    bool
	Cookie         []*http.Cookie
}

var (
	client         *HTTPClient
	once           sync.Once
	userAgent      = "grequest"
	timeoutDefault = 30 * time.Second
	// Errors
	ErrInvalidHost             = errors.New("Invalid Host Request")
	ErrInvalidURL              = errors.New("Invalid URL")
	ErrInvalidRedirectLocation = errors.New("Invalid Redirect Location")
	ErrTooManyRedirection      = errors.New("Too many Redirect")
)

func GetHTTPClient() *HTTPClient {
	once.Do(func() {
		client = New()
	})
	client.request = new(Request)
	client.maxRedirect = 0
	return client
}

func New() *HTTPClient {
	jar, err := cookiejar.New(&cookiejar.Options{})

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 32,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &HTTPClient{
		client: &http.Client{
			Jar:       jar,
			Transport: transport,
		},
		transport: transport,
		cjar:      jar,
		request: &Request{
			method:         http.MethodGet,
			isRequestReady: false,
			isRequested:    false,
			header:         make(http.Header),
		},
		maxRedirect:  0,
		cacheEnabled: false,
	}

	if err != nil {
		client.errs[0] = err
	}
	client.request.header["User-Agent"] = []string{userAgent}
	return client
}

func (c *HTTPClient) SetURL(u string) *HTTPClient {
	c.request.url = u
	return c
}

func (c *HTTPClient) Request() {
	if !c.request.isRequested {
		c.Do()
	}
}

func (c *HTTPClient) SetUserAgent(agent string) *HTTPClient {
	c.request.header.Set("User-Agent", agent)
	return c
}

func (c *HTTPClient) EnableRedirct() *HTTPClient {
	c.maxRedirect = 2
	c.redirectEnabled = true
	return c
}

func (c *HTTPClient) SetRequest(req *http.Request) *HTTPClient {
	c.request.req = req
	c.request.isRequestReady = true
	return c
}

func (c *HTTPClient) SetRedirectCount(count int) *HTTPClient {
	c.maxRedirect = count
	c.redirectEnabled = true
	return c
}

func (c *HTTPClient) SetTLSConfig(config *tls.Config) *HTTPClient {
	c.transport.TLSClientConfig = config
	c.client.Transport = c.transport
	return c
}

func (c *HTTPClient) newRequest(ctx context.Context) *HTTPClient {

	parsedURL, err := checkURL(c.request.url)

	if err != nil {
		c.errs = append(c.errs, err)
		return c
	}

	c.request.url = parsedURL.String()
	c.request.req, err = http.NewRequestWithContext(ctx, c.request.method, c.request.url, c.request.body)

	if err != nil {
		c.errs = append(c.errs, err)
		return c
	}

	c.request.req.Header = c.request.header

	if c.request.basic != nil {
		c.request.req.SetBasicAuth(c.request.basic.User, c.request.basic.Pass)
	}
	if c.request.Cookie != nil {
		for _, cookie := range c.request.Cookie {
			c.request.req.AddCookie(cookie)
		}
	}
	c.request.isRequestReady = true

	return c
}

func (c *HTTPClient) SetTimeout(timeout time.Duration) *HTTPClient {
	c.Timeout = timeout
	return c
}

func (c *HTTPClient) SetTimeoutSecons(second time.Duration) *HTTPClient {
	c.Timeout = second * time.Second
	return c
}

func (c *HTTPClient) WithRetry(maxRetries int, code int) *HTTPClient {
	for i := 0; i <= maxRetries; i++ {
		if c.Status().GetCode() == code {
			return c.Do()
		}
	}
	return c
}

// Makes a request to the http server
func (c *HTTPClient) Do() *HTTPClient {
	timeoutRequest := c.Timeout
	if timeoutRequest == 0 {
		timeoutRequest = timeoutDefault
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeoutRequest)
	defer cancel()
	return c.newRequest(ctx).do()
}

func (c *HTTPClient) do() *HTTPClient {

	var res *http.Response
	var err error
	res, err = c.client.Do(c.request.req)
	if err != nil {
		c.errs = append(c.errs, err)
		return c
	}
	//close main body reader
	defer c.Close()
	status := res.StatusCode

	if c.redirectEnabled && c.maxRedirect > 0 && status != 300 && status/100 == 3 {
		res, err = c.redirectRequest(c.request.req, res, 0)
		if err != nil {
			c.res = res
			c.errs = append(c.errs, err)
		}
	}
	if res.Header.Get("Content-Encoding") == "gzip" {
		var gres io.ReadCloser
		gres, err = gzip.NewReader(res.Body)
		if err != nil {
			c.res = res
			c.errs = append(c.errs, err)
			c.request.isRequested = true
			return c
		}
		res.Body = gres
	}
	//get bytes
	bodyBytes, _ := io.ReadAll(res.Body)
	//copy bytes in BodyBytes for reusability
	c.BodyBytes = bodyBytes
	c.res = res
	c.request.isRequested = true

	return c
}

func (c *HTTPClient) redirectRequest(req *http.Request, res *http.Response, count int) (rres *http.Response, err error) {

	if count > c.maxRedirect {
		return res, ErrTooManyRedirection
	}

	rreq := req

	loc := res.Header.Get("Location")

	if len(loc) == 0 {
		return res, ErrInvalidRedirectLocation
	}

	rreq.URL, err = url.ParseRequestURI(loc)
	if err == nil {
		rres, err = c.client.Transport.RoundTrip(rreq)
		if err == nil {
			switch rres.StatusCode / 100 {
			case 2:
				return rres, nil
			case 3:
				return c.redirectRequest(rreq, rres, count+1)
			case 4, 5:
				return rres, errors.New(http.StatusText(rres.StatusCode))
			}
		}
	}
	return res, err
}

func (c *HTTPClient) response() *http.Response {
	return c.res
}

func (c *HTTPClient) requests() *http.Request {
	return c.request.req
}

func (c *HTTPClient) GetResponse() (*http.Response, []error) {
	if !c.request.isRequested {
		c.Do()
	}
	return c.res, c.errs
}

func (c *HTTPClient) GetErrors() []error {
	return c.errs
}

func (c *HTTPClient) ResetClient() *HTTPClient {
	return New()
}

func (c *HTTPClient) Close() []error {
	err := c.res.Body.Close()
	if err != nil {
		c.errs = append(c.errs, err)
	}
	errs := c.errs
	c = nil
	return errs
}

func (c *HTTPClient) GetCurrentUrl() *url.URL {
	parsedURL, _ := url.Parse(c.request.url)
	return parsedURL
}

func checkURL(u string) (*url.URL, error) {
	parsedURL, err := url.Parse(u)

	if err != nil {
		return nil, ErrInvalidURL
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}

	if parsedURL.Host == "" {
		return nil, ErrInvalidHost
	}

	if parsedURL.String() == "" {
		return nil, ErrInvalidURL
	}

	return parsedURL, nil
}
