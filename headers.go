package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	traceIDHeader       = "X-trace-id"
	contentTypeHeader   = "Content-type"
)

const jsonContent = "application/json"

// WithTraceID sets the X-trace-id header to provided trace id.
func (client *Client) WithTraceID(id string) *Client {
	client.headers.Set(traceIDHeader, id)
	return client
}

// WithContentType sets the Content-type header to provided content type.
func (client *Client) WithContentType(content string) *Client {
	client.headers.Set(contentTypeHeader, content)
	return client
}

// WithJSONContent sets the Content-type header to application/json
func (client *Client) WithJSONContent() *Client {
	client.headers.Set(contentTypeHeader, jsonContent)
	return client
}

// WithAuthHeader sets the Authorization header to provided token.
func (client *Client) WithAuthHeader(token string) *Client {
	client.headers.Set(authorizationHeader, token)
	return client
}

func (client *Client) injectHeaders(request *http.Request) {
	request.Header = client.headers
}

// InheritFromParentContext set the client's headers to headers founded in the
// provided context
func (client *Client) InheritFromParentContext(ctx *gin.Context) *Client {
	if ctx == nil || ctx.Request == nil {
		return client
	}
	if len(ctx.Request.Header) == 0 {
		return client
	}

	client.headers.Set(authorizationHeader, ctx.GetHeader(authorizationHeader))
	return client
}

// SetHeader sets provided key - value in the headers
func (client *Client) SetHeader(header, value string) {
	client.headers.Set(header, value)
}

// AddHeader adds provided key - value to the headers
func (client *Client) AddHeader(header, value string) {
	client.headers.Set(header, value)
}
