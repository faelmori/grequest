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
	"slices"
	"sync"
	"time"
)

type HTTPClient struct {
	cacheEnabled bool
	errors       []error
	maxRedirect  int
	retryCodes   []int
	maxRetries   int
	Timeout      time.Duration
	userAgent    string
	client       *http.Client
	transport    *http.Transport
	cjar         *cookiejar.Jar
	request      *Request
	res          *http.Response
	BodyBytes    []byte
	errs         []error
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

const (
	maxRetriesDefault  = 1
	maxRedirectDefault = 5
	userAgentField     = "User-Agent"
	userAgentName      = "grequest"
)

var (
	client         *HTTPClient
	once           sync.Once
	timeoutDefault = 30 * time.Second
	// Errors
	ErrInvalidHost             = errors.New("Invalid Host Request")
	ErrInvalidURL              = errors.New("Invalid URL")
	ErrInvalidRedirectLocation = errors.New("Invalid Redirect Location")
	ErrTooManyRedirection      = errors.New("Too many Redirect")
	ErrTooManyRetry            = errors.New("Too many Retry")
)

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
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		transport: transport,
		cjar:      jar,
		request: &Request{
			method:         http.MethodGet,
			isRequestReady: false,
			isRequested:    false,
			header:         make(http.Header),
		},
		maxRedirect:  maxRedirectDefault,
		cacheEnabled: false,
	}

	if err != nil {
		client.errs[0] = err
	}
	client.request.header[userAgentField] = []string{userAgentName}
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
	c.request.header.Set(userAgentField, agent)
	return c
}

// Set max redirect count
// 0 = disable redirects
func (c *HTTPClient) MaxRedirect(maxRedirect int) *HTTPClient {
	c.maxRedirect = maxRedirect
	return c
}

func (c *HTTPClient) SetRequest(req *http.Request) *HTTPClient {
	c.request.req = req
	c.request.isRequestReady = true
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

// Sets the conditions for retrying a http request
func (c *HTTPClient) RetryIf(statusCodes ...int) *HTTPClient {
	c.retryCodes = statusCodes
	c.maxRetries = maxRetriesDefault
	return c
}

// Sets the counts for retrying a http request
func (c *HTTPClient) RetryMax(maxRetries int) *HTTPClient {
	c.maxRetries = maxRetries
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

func (c *HTTPClient) DoWithContext(ctx context.Context) *HTTPClient {
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
	defer c.Close()
	status := res.StatusCode
	if c.maxRetries > 0 && slices.Contains(c.retryCodes, status) {
		res, err = c.retryRequest(c.request.req, res, 0)
		if err != nil {
			c.res = res
			c.errs = append(c.errs, err)
		}
	}

	if c.maxRedirect > 0 && status != 300 && status/100 == 3 {
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
	bodyBytes, _ := io.ReadAll(res.Body)
	c.BodyBytes = bodyBytes
	c.res = res
	c.request.isRequested = true

	return c
}

func (c *HTTPClient) retryRequest(req *http.Request, res *http.Response, count int) (*http.Response, error) {
	if count <= c.maxRetries {
		retryRes, err := c.client.Transport.RoundTrip(req)
		if err != nil {
			c.res = retryRes
			c.errs = append(c.errs, err)
		}
		if slices.Contains(c.retryCodes, res.StatusCode) {
			return c.retryRequest(req, retryRes, count+1)
		}
		return retryRes, nil
	}
	return res, ErrTooManyRetry
}

func (c *HTTPClient) redirectRequest(req *http.Request, res *http.Response, count int) (redirectRes *http.Response, err error) {

	if count > c.maxRedirect {
		return res, ErrTooManyRedirection
	}
	redirectReq := req
	loc := res.Header.Get("Location")
	if len(loc) == 0 {
		return res, ErrInvalidRedirectLocation
	}
	var redirectToUrl *url.URL
	urlLoc, err := checkURL(loc)
	redirectToUrl = urlLoc
	if err != nil {
		toLoc, _ := url.ParseRequestURI(req.URL.String() + loc)
		redirectToUrl = toLoc
	}

	redirectReq.URL = redirectToUrl
	redirectRes, err = c.client.Transport.RoundTrip(redirectReq)
	if err == nil {
		switch redirectRes.StatusCode / 100 {
		case 2:
			return redirectRes, nil
		case 3:
			return c.redirectRequest(redirectReq, redirectRes, count+1)
		case 4, 5:
			return redirectRes, errors.New(http.StatusText(redirectRes.StatusCode))
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
