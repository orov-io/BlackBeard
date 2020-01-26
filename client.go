package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "authorization"
	uriSeparator        = "/"
	portSeparator       = ":"
	basePathKey         = "BASE_PATH"
	contentTypeHeader   = "Content-type"
	jsonContent         = "application/json"
	keyQuery            = "key"
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

// WithAuthHeader attaches provided bearer to the client.
func (client *Client) WithAuthHeader(bearer string) *Client {
	client.headers.Set(authorizationHeader, bearer)
	return client
}

// InheritFromParentContext set the client's bearer token to the authorization header
// in the provided context.
func (client *Client) InheritFromParentContext(ctx *gin.Context) *Client {
	client.parentCtx = ctx
	return client.WithAuthHeader(ctx.GetHeader(authorizationHeader))

}

// WithDefaultBasePath tries to infer the server path from the env variable
// BASE_PATH.
func (client *Client) WithDefaultBasePath() *Client {
	path := os.Getenv(basePathKey)
	return client.WithBasePath(path)
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

// WithHeaders adds the key-value pair to client request headers.
// It internally uses the http.CanonicalHeaderKey to format the key.
//
// Please, note that this client is build to provide JSON exchange, so
// the content type header will be overwrite by application/json
func (client *Client) WithHeaders(headers http.Header) *Client {
	client.headers = headers
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
func (client *Client) GET(path string, data interface{}) (*http.Response, error) {
	return client.executeCall(http.MethodGet, path, data)
}

// POST performs a secure POST petition. Final URI will be client base path + provided path
func (client *Client) POST(path string, data interface{}) (*http.Response, error) {
	return client.executeCall(http.MethodPost, path, data)
}

// PUT performs a secure PUT petition. Final URI will be client base path + provided path
func (client *Client) PUT(path string, data interface{}) (*http.Response, error) {
	return client.executeCall(http.MethodPut, path, data)
}

// DELETE performs a secure DELETE petition. Final URI will be client base path + provided path
func (client *Client) DELETE(path string, data interface{}) (*http.Response, error) {
	return client.executeCall(http.MethodDelete, path, data)
}

func (client *Client) executeCall(method, path string, data interface{}) (*http.Response, error) {
	body, err := client.interface2Body(data)
	if err != nil {
		return nil, err
	}

	path = strings.TrimLeft(path, uriSeparator)
	URI := fmt.Sprintf("%v%v", client.getURI(), path)
	request, err := http.NewRequest(method, URI, body)
	if err != nil {
		return nil, err
	}

	if client.shouldAddAPIKey() {
		query := request.URL.Query()
		query.Add(keyQuery, client.apiKey)
		request.URL.RawQuery = query.Encode()
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

func (client *Client) injectHeaders(request *http.Request) {
	if client.isUsingCustomHeaders() {
		client.injectCustomHeaders(request)
		return
	}

	client.injectDefaultHeaders(request)
}

func (client *Client) isUsingCustomHeaders() bool {
	return len(client.headers) > 0
}

func (client *Client) injectCustomHeaders(request *http.Request) {
	request.Header = client.headers
}

func (client *Client) injectDefaultHeaders(request *http.Request) {
	client.injectContentTypeHeader(request)
}

func (client *Client) injectContentTypeHeader(request *http.Request) {
	request.Header.Set(contentTypeHeader, jsonContent)
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

// GetPortFromEnvKey returns the port form the env variables if any.
func GetPortFromEnvKey(key string) int {
	envPort := os.Getenv(key)
	port, err := strconv.ParseInt(envPort, 10, 32)
	if err != nil {
		return 0
	}
	return int(port)
}

// CheckResponse checks for a valid response.
//
// One response will be valid if err is nil and response code is 200 ≤ code < 400
func CheckResponse(err error, resp *http.Response) error {
	if err != nil {
		return err
	}

	if !isCallOK(resp) {
		return BadRequestError()
	}

	return nil
}

func isCallOK(resp *http.Response) bool {
	return resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest
}

// BadRequest is used when user try to initialize the database and
//it is already initialized
type BadRequest struct {
	code int
}

func (e *BadRequest) Error() string {
	return fmt.Sprintf("Bad response code: %v", e.code)
}

// BadRequestError returns a new BadRequest error
func BadRequestError() error {
	return &BadRequest{}
}

// IsBadRequestError checks if the error is a BadRequest error
func IsBadRequestError(err error) bool {
	_, ok := err.(*BadRequest)
	return ok
}
