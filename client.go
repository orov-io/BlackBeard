package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	uriSeparator  = "/"
	portSeparator = ":"
	basePathKey   = "BASE_PATH"
	keyQuery      = "key"
)

// Client get basic support to make requests to the admin service.
type Client struct {
	parentCtx  context.Context
	ctx        context.Context
	basePath   string
	port       int
	version    string
	service    string
	httpClient *http.Client
	headers    http.Header
	apiKey     string
}

// MakeNewClient initializes and returns a new fresh service client.
func MakeNewClient() *Client {
	client := &Client{}
	client.httpClient = &http.Client{}
	client.ctx = context.Background()
	client.headers = http.Header{}

	return client
}

// WithBasePath set the client's base path.
func (client *Client) WithBasePath(path string) *Client {
	client.basePath = strings.TrimRight(path, uriSeparator)
	return client
}

// WithPort set the client's port to call.
func (client *Client) WithPort(port int) *Client {
	client.port = port
	return client
}

// ToService set the service destination
func (client *Client) ToService(service string) *Client {
	client.service = service
	return client
}

// WithVersion set the API version
func (client *Client) WithVersion(version string) *Client {
	client.version = version
	return client
}

// WithTimeout set a timeout to the api requests.
func (client *Client) WithTimeout(duration time.Duration) *Client {
	client.httpClient.Timeout = duration
	return client
}

// WithAPIKey adds a 'key' parameter to the call query
func (client *Client) WithAPIKey(key string) *Client {
	client.apiKey = key
	return client
}

// GET performs a secure GET petition. Final URI will be client base path + provided path
func (client *Client) GET(path string, body interface{}, query map[string]string) (*http.Response, error) {
	return client.executeCall(http.MethodGet, path, body, query)
}

// POST performs a secure POST petition. Final URI will be client base path + provided path
func (client *Client) POST(path string, data interface{}) (*http.Response, error) {
	return client.executeCall(http.MethodPost, path, data, nil)
}

// PUT performs a secure PUT petition. Final URI will be client base path + provided path
func (client *Client) PUT(path string, data interface{}) (*http.Response, error) {
	return client.executeCall(http.MethodPut, path, data, nil)
}

// DELETE performs a secure DELETE petition. Final URI will be client base path + provided path
func (client *Client) DELETE(path string, data interface{}) (*http.Response, error) {
	return client.executeCall(http.MethodDelete, path, data, nil)
}

func (client *Client) executeCall(method, path string, data interface{}, query map[string]string) (*http.Response, error) {
	body, err := client.interface2Body(data)
	if err != nil {
		return nil, err
	}

	endpoint, err := url.Parse(fmt.Sprintf("%v%v", client.getURI(), strings.TrimLeft(path, uriSeparator)))
	if err != nil {
		return nil, err
	}

	client.addQuery(endpoint, query)
	request, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	client.injectHeaders(request)
	return client.do(request)
}

func (client *Client) interface2Body(data interface{}) (io.Reader, error) {
	if data == nil {
		return nil, nil
	}

	requestBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(requestBody), nil
}

func (client *Client) getURI() string {
	URI := fmt.Sprintf("%v", client.basePath)

	if client.shouldAddPort() {
		URI = fmt.Sprintf("%v%v%v", URI, portSeparator, client.port)
	}

	URI = fmt.Sprintf("%v%v", URI, uriSeparator)

	if client.shouldAddVersion() {
		URI = fmt.Sprintf("%v%v%v", URI, client.version, uriSeparator)
	}

	if client.shouldAddService() {
		URI = fmt.Sprintf("%v%v%v", URI, client.service, uriSeparator)
	}
	return URI
}

func (client *Client) shouldAddPort() bool {
	return client.port != 0
}

func (client *Client) shouldAddVersion() bool {
	return client.version != ""
}

func (client *Client) shouldAddAPIKey() bool {
	return client.apiKey != ""
}

func (client *Client) shouldAddService() bool {
	return client.service != ""
}

func (client *Client) do(request *http.Request) (*http.Response, error) {
	return client.httpClient.Do(request)
}

// ------ Generic Getters ------\\

// GetHeaders returns the client actual header
func (client *Client) GetHeaders() http.Header {
	return client.headers
}

// GetBasePath returns the client actual header
func (client *Client) GetBasePath() string {
	return client.basePath
}

// GetService returns the client actual header
func (client *Client) GetService() string {
	return client.service
}

// GetVersion returns the client actual header
func (client *Client) GetVersion() string {
	return client.version
}

// GetTimeout returns the client actual header
func (client *Client) GetTimeout() time.Duration {
	return client.httpClient.Timeout
}

// GetPort returns the client port
func (client *Client) GetPort() int {
	return client.port
}

func (client *Client) addQuery(endpoint *url.URL, query map[string]string) {
	if query == nil {
		return
	}

	queryValues, _ := url.ParseQuery(endpoint.RawQuery)

	for key, value := range query {
		queryValues.Add(key, value)
	}

	if client.shouldAddAPIKey() {
		queryValues.Add(keyQuery, client.apiKey)
	}

	endpoint.RawQuery = queryValues.Encode()
	return
}
