package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// Usefull constants
const (
	TestAuthPrefix = "Bearer"
	AuthHeader     = "authorization"
)

// GetNewGinContextWithAuthBearer returns a new gin test context with the
// provided bearer attached to the authorization header
func GetNewGinContextWithAuthBearer(bearer string) (*gin.Context, string) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctxRequest, err := http.NewRequest(http.MethodGet, "url", nil)
	if err != nil {
		panic(err)
	}
	authHeaderContent := fmt.Sprintf("%v %v", TestAuthPrefix, bearer)

	ctxRequest.Header.Set(AuthHeader, authHeaderContent)
	ctx.Request = ctxRequest

	return ctx, authHeaderContent
}
