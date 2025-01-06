package grequest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
)

type Cookie struct {
	savePath string
	cookies  *http.Cookie
	Client   *HTTPClient
}
type SerializableCookies struct {
	URL     string         `json:"url"`
	Cookies []*http.Cookie `json:"cookies"`
}

const cookiesPath = "cookies"

// This method init client with cookie
func (c *HTTPClient) Cookie() *Cookie {
	return &Cookie{Client: c}
}

// Gets http client object
func (c *Cookie) client() *HTTPClient {
	return c.Client
}

func (c *Cookie) loadPath() string {
	toPath := cookiesPath
	if c.savePath != "" {
		return c.savePath
	}
	currentHost := c.client().GetCurrentUrl()
	fromFilePath := fmt.Sprintf("%s/%s/cookies.json", toPath, currentHost.Host)
	return fromFilePath
}

// Sets cookies as string ex: name=xxxx; count=x
func (c *Cookie) SetString(cookie string) *HTTPClient {
	c.Client.Header().Set("Cookie", cookie)
	return c.Client
}

// Sets cookies as *http.Cookie
func (c *Cookie) SetCookies(cookies []*http.Cookie) *HTTPClient {
	c.Client.request.Cookie = cookies
	return c.Client
}

func (c *Cookie) SetCookieJar(jar *cookiejar.Jar) *HTTPClient {
	c.Client.cjar = jar
	c.Client.client.Jar = c.Client.cjar
	return c.Client
}

// Gets cookies as *http.Cookie
func (c *Cookie) Get() []*http.Cookie {
	return c.Client.response().Cookies()
}

// Sets path for save
func (c *Cookie) Path(path string) *Cookie {
	c.savePath = path
	return c
}

// Save cookies to file for next requests
// by default saves to the cookies/domainname directory
func (c *Cookie) Save() *Cookie {
	cookies := c.Get()
	data := SerializableCookies{URL: c.Client.request.url, Cookies: cookies}

	cookieBytes, err := json.Marshal(data)
	if err != nil {
		return c
	}
	cookieReader := io.NopCloser(bytes.NewReader(cookieBytes))
	toFilePath := c.loadPath()
	saveToFile(toFilePath, cookieReader)
	return c
}

func (c *Cookie) Load() *Cookie {
	fromFilePath := c.loadPath()
	file, err := os.ReadFile(fromFilePath)
	if err != nil {
		return c
	}

	var data SerializableCookies
	if err := json.Unmarshal(file, &data); err != nil {
		return c
	}
	c.SetCookies(data.Cookies)
	return c
}
