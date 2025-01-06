package grequest

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/url"
)

type FormData struct {
	formDataContentType string
	urlValues           url.Values
	isMultipartForm     bool
	body                *bytes.Buffer
	writer              *multipart.Writer
	Client              *HTTPClient
}

// Init status
func (c *HTTPClient) FormData() *FormData {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	urlValues := url.Values{}
	return &FormData{Client: c, writer: writer, body: body, urlValues: urlValues}
}

// Gets http client object
func (c *FormData) client() *HTTPClient {
	return c.Client
}

// SetMultipart
func (c *FormData) WithMultipart() *FormData {
	c.isMultipartForm = true
	return c
}

// Sets form data field
func (c *FormData) AddField(key, value string) *FormData {
	if c.isMultipartForm {
		return c.setFieldMultipart(key, value)
	}
	return c.SetFieldEncode(key, value)
}

func (c *FormData) SetFieldEncode(key, value string) *FormData {
	c.urlValues.Add(key, value)
	return c
}

// Sets form data field
func (c *FormData) setFieldMultipart(fieldname, value string) *FormData {
	_ = c.writer.WriteField(fieldname, value)
	return c
}

// Sets form data fields
func (c *FormData) SetFields(fields *map[string]string) *HTTPClient {
	for k, val := range *fields {
		if c.isMultipartForm {
			c.setFieldMultipart(k, val)
			continue
		}
		c.SetFieldEncode(k, val)
	}
	return c.Push()
}

// Attach file to request
func (c *FormData) AddFile(key, path string) *FormData {
	c.WithMultipart()
	file, err := readFileByPath(path)
	defer file.Close()
	if err != nil {
		return c
	}
	fileName := getFileNameByPath(path)
	part, err := c.writer.CreateFormFile(key, fileName)
	if err != nil {
		return c
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return c
	}

	c.formDataContentType = c.writer.FormDataContentType()
	return c
}

func (c *FormData) Push() *HTTPClient {
	c.writer.Close()
	if !c.isMultipartForm {
		c.client().ContentType().SetFormUrlencoded()
		c.client().Body().SetString(c.urlValues.Encode())
	}
	if c.isMultipartForm {
		if c.formDataContentType != "" {
			c.client().ContentType().Set(c.formDataContentType)
		}
		if c.formDataContentType == "" {
			c.client().ContentType().SetMultipartFormData()
		}
		c.client().Body().Set(c.body)
	}

	return c.client()
}
