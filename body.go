package grequest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Body struct {
	savePath string
	Client   *HTTPClient
}

const filesPath = "files"

// Sets the request body with io.Reader
func (c *HTTPClient) Body() *Body {
	return &Body{Client: c}
}

// Gets http client object
func (c *Body) client() *HTTPClient {
	return c.Client
}

func (c *Body) loadPath() string {
	toPath := filesPath
	if c.savePath != "" {
		return c.savePath
	}
	currentHost := c.client().GetCurrentUrl()
	path := fmt.Sprintf("%s/%s", toPath, currentHost.Host)
	return path
}

// Gets raw io.ReadCloser from body
func (c *Body) get() io.Reader {
	return io.NopCloser(bytes.NewReader(c.client().BodyBytes))
}

func (c *Body) getBytes() []byte {
	return c.client().BodyBytes
}

// Sets the request body with io.Reader
func (c *Body) Set(body io.Reader) *HTTPClient {
	c.client().request.body = body
	return c.client()
}

// Sets the request body with only string
func (c *Body) SetString(body string) *HTTPClient {
	c.client().request.body = strings.NewReader(body)
	return c.client()
}

// Sets the request body with bytes
func (c *Body) SetByte(body []byte) *HTTPClient {
	c.client().request.body = bytes.NewBuffer(body)
	return c.client()
}

// Sets the request body with json
func (c *Body) SetJson(data interface{}) *HTTPClient {
	jsonData, _ := json.Marshal(data)
	c.client().request.body = bytes.NewBuffer(jsonData)
	return c.client()
}

// Gets raw io.ReadCloser from body
func (c *Body) GetRaw() io.Reader {
	return c.get()
}

// Gets response body and return in bytes
func (c *Body) GetBytes() []byte {
	b := c.getBytes()
	return b
}

// Gets response body and return as string
func (c *Body) GetStrings() string {
	b := c.getBytes()
	return string(b)
}

// Gets response body and return as interface map array
func (c *Body) GetWithJson() (interface{}, error) {
	content := c.getBytes()
	var WithJson interface{}
	err := json.Unmarshal(content, &WithJson)
	return WithJson, err
}

// Gets response body and make to struct
func (c *Body) GetWithJsonStruct(target interface{}) error {
	b := c.getBytes()

	err := json.Unmarshal(b, &target)
	if err != nil {
		return err
	}
	return nil
}

// Sets path for save
func (c *Body) Path(path string) *Body {
	c.savePath = path
	return c
}

// Save response body to file
func (c *Body) ToFile(fileName string) {
	saveToFile(fileName, c.get())
}

// Save response body to file
func (c *Body) SaveFile() *Body {
	extension := getFileExtensionByContentType(c.client().ContentType().Get())
	currentHost := c.client().GetCurrentUrl()
	name := getFileNameByPath(currentHost.Path)
	fileName := fmt.Sprintf("%s/%s%s", c.loadPath(), name, extension)
	saveToFile(fileName, c.get())
	return c
}
