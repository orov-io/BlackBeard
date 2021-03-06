package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v2"
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
	cacheDB    *badger.DB
	logger     Logger
}

// MakeNewClient initializes and returns a new fresh service client.
func MakeNewClient() *Client {
	client := &Client{}
	client.httpClient = &http.Client{}
	client.ctx = context.Background()
	client.headers = http.Header{}
	client.logger = &noLogger{}

	return client
}

// WithLogger attach a logger to the client
func (client *Client) WithLogger(logger Logger) *Client {
	client.logger = logger
	return client
}

// WithCache enables caching results for this client object.
func (client *Client) WithCache() *Client {
	options := badger.DefaultOptions("").WithInMemory(true)
	client.cacheDB, _ = badger.Open(options)
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

// GetFullPath returns the full path to the service base URL
func (client *Client) GetFullPath() string {
	return client.getURI()
}

// GET performs a secure GET petition. Final URI will be client base path + provided path
func (client *Client) GET(path string, body interface{}, query map[string][]string) (*http.Response, error) {
	return client.executeCall(http.MethodGet, path, body, query)
}

// POST performs a secure POST petition. Final URI will be client base path + provided path
func (client *Client) POST(path string, body interface{}, query map[string][]string) (*http.Response, error) {
	return client.executeCall(http.MethodPost, path, body, query)
}

// MultipartBody models the body of a multipart POST call, where:
// files: a map in with the key represent the form key, and the value represents the path to the file.
// params: A map with the key-values to be send in the body with the files.
type MultipartBody struct {
	Params map[string]string
	Files  map[string]string
}

// NewMultipartBody returns a new struct with desired values attached.
func NewMultipartBody(params map[string]string, files map[string]string) MultipartBody {
	return MultipartBody{
		Params: params,
		Files:  files,
	}
}

// MULTIPART performs a secure POST petition setting content type to be multipart/form-data.
// Final URI will be client base path + provided path
// You will need to provide the content type with boundary in formDataContentType.
func (client *Client) MULTIPART(
	path string,
	bodyData MultipartBody,
	query map[string][]string,
) (*http.Response, error) {

	body, formDataContentType, err := client.getMultipartBody(bodyData)
	if err != nil {
		return nil, err
	}

	headers := client.headers.Clone()
	client.headers.Set(contentTypeHeader, formDataContentType)
	resp, err := client.executeCall(http.MethodPost, path, body, query)
	client.headers = headers
	return resp, err
}

func (client *Client) getMultipartBody(data MultipartBody) (body *bytes.Buffer, contentType string, err error) {
	body = &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, path := range data.Files {
		var file *os.File
		file, err = os.Open(path)
		if err != nil {
			return
		}

		var part io.Writer
		part, err = writer.CreateFormFile(key, filepath.Base(path))
		if err != nil {
			return
		}
		_, err = io.Copy(part, file)
		file.Close()
	}

	for key, val := range data.Params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return
	}

	contentType = writer.FormDataContentType()
	return
}

// PUT performs a secure PUT petition. Final URI will be client base path + provided path
func (client *Client) PUT(path string, body interface{}, query map[string][]string) (*http.Response, error) {
	return client.executeCall(http.MethodPut, path, body, query)
}

// DELETE performs a secure DELETE petition. Final URI will be client base path + provided path
func (client *Client) DELETE(path string, body interface{}, query map[string][]string) (*http.Response, error) {
	return client.executeCall(http.MethodDelete, path, body, query)
}

func (client *Client) executeCall(method, path string, body interface{}, query map[string][]string) (*http.Response, error) {
	if response, isCached := client.callCached(method, path, body, query); isCached {
		client.logger.Debugf("Cached response for [%s] %s\n", method, path)
		return response, nil
	}

	bodyReader, err := client.interface2Reader(body)
	if err != nil {
		return nil, err
	}

	endpoint, err := url.Parse(fmt.Sprintf("%v%v", client.getURI(), strings.TrimLeft(path, uriSeparator)))
	if err != nil {
		return nil, err
	}

	client.addQuery(endpoint, query)
	request, err := http.NewRequest(method, endpoint.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	client.injectHeaders(request)
	response, err := client.do(request)
	if err != nil {
		return nil, err
	}

	client.cache(method, path, body, query, response)
	return response, nil
}

func (client *Client) callCached(method, path string, body interface{}, query map[string][]string) (*http.Response, bool) {
	if client.cacheDB == nil {
		return nil, false
	}
	key := getCacheKey(method, path, body, query)
	response := new(http.Response)
	err := client.cacheDB.View(getResponseFromCache(response, key))
	return response, err != nil
}

func getCacheKey(method, path string, body interface{}, query map[string][]string) []byte {
	key := make([]byte, 0)

	key = appendBytes(key, method)
	key = appendBytes(key, path)
	key = appendBytes(key, body)
	key = appendBytes(key, query)

	return key
}

func appendBytes(key []byte, value interface{}) []byte {
	b, _ := json.Marshal(value)
	return append(key, b...)
}

func getResponseFromCache(response *http.Response, key []byte) func(txn *badger.Txn) error {
	return func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			response = nil
			return nil
		}

		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &response)
		})

		return err
	}
}

func (client *Client) cache(method, path string, body interface{}, query map[string][]string, response *http.Response) {
	if client.cacheDB == nil {
		return
	}

	key := getCacheKey(method, path, body, query)
	value, _ := json.Marshal(response)
	client.cacheDB.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

func (client *Client) interface2Reader(data interface{}) (io.Reader, error) {
	if data == nil {
		return nil, nil
	}

	reader, ok := data.(io.Reader)
	if ok {
		return reader, nil
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

func (client *Client) addQuery(endpoint *url.URL, query map[string][]string) {
	if query == nil {
		return
	}

	queryValues, _ := url.ParseQuery(endpoint.RawQuery)

	for key, values := range query {
		for _, value := range values {
			queryValues.Add(key, value)
		}
	}

	if client.shouldAddAPIKey() {
		queryValues.Add(keyQuery, client.apiKey)
	}

	endpoint.RawQuery = queryValues.Encode()
	return
}
