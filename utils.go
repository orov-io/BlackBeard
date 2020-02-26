package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"

	"github.com/gin-gonic/gin"
)

// ErrorResponse models the common error response. Also implement the Error interface.
type ErrorResponse struct {
	Name      string            `json:"name,omitempty"`
	Message   string            `json:"message,omitempty"`
	Code      int               `json:"code,omitempty"`
	ClassName string            `json:"class_name,omitempty"`
	Data      map[string]string `json:"data,omitempty"`
	Errors    map[string]string `json:"errors,omitempty"`
}

func (e *ErrorResponse) Error() string {
	err, _ := json.Marshal(e)
	return fmt.Sprintf("ERROR:  %v", string(err))
}

// IsErrorResponse checks if the error is a ErrorResponse error
func IsErrorResponse(err error) bool {
	_, ok := err.(*ErrorResponse)
	return ok
}

// PaginatedResponse models a paginate response from services.
type PaginatedResponse struct {
	Total int           `json:"total,omitempty"`
	Limit int           `json:"limit,omitempty"`
	Skip  int           `json:"skip,omitempty"`
	Data  []interface{} `json:"data,omitempty"`
}

const testBearerTokenKey = "TEST_BEARER_TOKEN"
const authHeader = "authorization"

// GetNewGinContextWithAuthBearer returns a new gin context with a auth header
// filled with the token found in the TEST_BEARER_TOKEN env variable
func GetNewGinContextWithAuthBearer() (*gin.Context, string) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	request, err := http.NewRequest(http.MethodGet, "url", nil)
	if err != nil {
		panic(err)
	}
	testAuthBearer := os.Getenv(testBearerTokenKey)
	request.Header.Set(authHeader, testAuthBearer)
	ctx.Request = request

	return ctx, testAuthBearer
}

// ParseAllPaginated parses all occurrences of a paginated response to the
// receiver.
func ParseAllPaginated(resp *http.Response, receiver interface{}) error {
	paginatedData, err := getPaginatedData(resp)
	if err != nil {
		return err
	}

	return ParseTo(paginatedData, receiver)
}

func getPaginatedData(resp *http.Response) (*PaginatedResponse, error) {
	if !isValidResponse(resp) {
		return nil, parseError(resp)
	}

	paginatedData := new(PaginatedResponse)
	body, err := Body2Interface(resp)
	if err != nil {
		return nil, err
	}

	err = ParseTo(body, paginatedData)
	if err != nil {
		return nil, err
	}

	return paginatedData, nil
}

func parseError(resp *http.Response) error {
	errorResponse := new(ErrorResponse)
	body, err := Body2Interface(resp)
	if err != nil {
		return &ErrorResponse{
			Name: "No standar error found",
			Code: resp.StatusCode,
			Errors: map[string]string{
				"parsed error": err.Error(),
			},
		}
	}

	err = ParseTo(body, errorResponse)
	if err != nil {
		return &ErrorResponse{
			Name: "No standar error found",
			Code: resp.StatusCode,
			Errors: map[string]string{
				"parsed error": err.Error(),
			},
		}
	}

	return errorResponse
}

// ParseOnePaginated parses first item of the response data
func ParseOnePaginated(resp *http.Response, receiver interface{}) error {
	paginatedData, err := getPaginatedData(resp)
	if err != nil {
		return err
	}

	return ParseTo(paginatedData.Data[0], receiver)
}

// ParseTo parses generic interface to another. As this is a generic function
// that makes a high use of json marshaller, it has some additional computational
// cost.
func ParseTo(data, receiver interface{}) error {
	if !isAPointer(receiver) {
		return NewNotAPointerError()
	}

	ResponseBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Error: %v\nCan't marshal response data: %v", err, data)
	}
	err = json.Unmarshal(ResponseBytes, receiver)
	if err != nil {
		return fmt.Errorf("Error: %v\nCan't unmarshal response data: %v", err, data)
	}

	return nil
}

// Body2Interface parses a body of an http response to a empty interface
func Body2Interface(resp *http.Response) (interface{}, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data interface{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func isAPointer(i interface{}) bool {
	return reflect.ValueOf(i).Kind() == reflect.Ptr
}

// NotAPointerError is used to send a hidden error
type NotAPointerError struct{}

func (e *NotAPointerError) Error() string {
	return fmt.Sprintf("Receiver is not a pointer")
}

// NewNotAPointerError returns a new NotAPointerErrorError error
func NewNotAPointerError() error {
	return &NotAPointerError{}
}

// IsNotAPointerError checks if the error is a NotAPointerError error
func IsNotAPointerError(err error) bool {
	_, ok := err.(*NotAPointerError)
	return ok
}

func isValidResponse(response *http.Response) bool {
	return response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusBadRequest
}
