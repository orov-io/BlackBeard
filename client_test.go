package api_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"

	api "github.com/orov-io/BlackBeard"
)

const (
	testAuthBearer         = "Bearer testBearer"
	authHeader             = "authorization"
	testBasePathKey        = "BASE_PATH"
	testBasePath           = "http://localhost"
	testPort               = 3000
	testTargetService      = "truman"
	testTimeout            = 3
	testDurationMultiplier = time.Second
	postsEndpoint          = "/posts"
	serverDB               = "./db.json"
	serverDBSeed           = "./dbSeed.json"
	serverChangedDB        = "./~db.json"
	testVersion            = "vTest"
)

const (
	givenAClient  = "Given a client"
	validResponse = "Then we obtain a valid response"
)

var jsonServer *exec.Cmd

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	go startJSONServer()
	time.Sleep(2 * time.Second)
}

func startJSONServer() {
	jsonServer = exec.Command("json-server", "--watch", serverDB)
	err := jsonServer.Run()
	if err != nil {
		fmt.Printf("The error: %v", err)
		panic(err)
	}
}

func shutdown() {
	stopJSONServer()
	restoreDB()
}

func stopJSONServer() {
	jsonServer.Process.Kill()
}

func restoreDB() {
	removeChangedDB()
	copySeedDB()
}

func removeChangedDB() {
	os.Remove(serverChangedDB)
}

func copySeedDB() {
	seed, err := os.Open(serverDBSeed)
	if err != nil {
		panic(err)
	}
	defer seed.Close()

	err = os.Remove(serverDB)
	if err != nil {
		panic(err)
	}

	db, err := os.Create(serverDB)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = io.Copy(db, seed)
	if err != nil {
		panic(err)
	}
}

func TestMakeNewClient(t *testing.T) {
	client := api.MakeNewClient()

	Convey("Given a new client call", t, func() {
		Convey("When the struct is returned", func() {
			Convey("Then the client is not empty", func() {
				So(client, ShouldNotBeNil)
			})
			Convey("Then the client is a valid instance", func() {
				So(client, ShouldHaveSameTypeAs, &api.Client{})
			})
		})

	})
}

func TestWithAuthBearer(t *testing.T) {
	Convey("Given a bearer auth header", t, func() {
		bearer := testAuthBearer

		Convey("When the client has initialized with the bearer token", func() {
			client := api.MakeNewClient().WithAuthHeader(bearer)

			Convey("Then the client auth header must be set to the bearer token", func() {
				authHeader := client.GetHeaders().Get(authHeader)
				So(authHeader, ShouldEqual, bearer)
			})
		})
	})
}

func TestInheritFromParentContext(t *testing.T) {
	Convey("Given a parent gin.Context with an auth bearer", t, func() {
		context, bearer := getNewGinContextWithAuthBearer()

		Convey("When the client has initialized with the context", func() {
			client := api.MakeNewClient().InheritFromParentContext(context)

			Convey("Then the client inherits from the context auth header", func() {
				authHeader := client.GetHeaders().Get(authHeader)
				So(authHeader, ShouldEqual, bearer)
			})
		})
	})
}

func TestWithDefaultBasePath(t *testing.T) {
	Convey("Given a project id key-value on env variables", t, func() {
		os.Setenv(testBasePathKey, testBasePath)

		Convey("When it's initialized with the default base path", func() {
			client := api.MakeNewClient().WithDefaultBasePath()

			Convey("Then client base path is set to default base path", func() {
				So(client.GetBasePath(), ShouldEqual, testBasePath)
			})
		})
	})
}

func TestWithPort(t *testing.T) {
	Convey("Given a target service", t, func() {
		port := testPort

		Convey("When a client it's initialized with this service", func() {
			client := api.MakeNewClient().WithPort(port)

			Convey("Then service is sets on the client", func() {

				So(client.GetPort(), ShouldEqual, port)
			})
		})
	})
}

func TestToService(t *testing.T) {
	Convey("Given a target service", t, func() {
		service := testTargetService

		Convey("When a client it's initialized with this service", func() {
			client := api.MakeNewClient().ToService(service)

			Convey("Then service is sets on the client", func() {

				So(client.GetService(), ShouldEqual, service)
			})
		})
	})
}

func TestWithVersion(t *testing.T) {
	Convey("Given an API version", t, func() {
		version := testVersion

		Convey("When a client it's initialized with this service", func() {
			client := api.MakeNewClient().WithVersion(version)

			Convey("Then service is sets on the client", func() {

				So(client.GetVersion(), ShouldEqual, version)
			})
		})
	})
}

func TestWhitHeaders(t *testing.T) {
	Convey("Given a set of valid headers", t, func() {
		headers := getTestHeaders()

		Convey("When the client is initialized with custom headers", func() {
			client := api.MakeNewClient().WithHeaders(headers)

			Convey("Then headers is sets on the client", func() {

				So(client.GetHeaders(), ShouldResemble, headers)
			})
		})
	})
}

func TestWhitTimeout(t *testing.T) {
	Convey("Given a desired timeout to wait", t, func() {
		timeout := testTimeout * testDurationMultiplier

		Convey("When the client is initialized with custom headers", func() {
			client := api.MakeNewClient().WithTimeout(timeout)

			Convey("Then headers is sets on the client", func() {

				So(client.GetTimeout(), ShouldEqual, timeout)
			})
		})
	})
}

func TestGET(t *testing.T) {
	Convey(givenAClient, t, func() {
		client := getDefaultTestClient()

		Convey("When we make a valid GET call", func() {
			resp, err := client.GET(postsEndpoint, nil)

			Convey(validResponse, func() {
				checkResponseIsValid(resp, err)
			})
		})
	})
}

func TestGETSadPath(t *testing.T) {
	Convey(givenAClient, t, func() {
		client := getDefaultTestClient()

		Convey("When we make a invalid GET call", func() {
			resp, err := client.GET("/wrong", nil)

			Convey("Then we obtain a not found response", func() {
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}

func TestPOST(t *testing.T) {
	Convey(givenAClient, t, func() {
		client := getDefaultTestClient()

		Convey("When we make a valid POST call", func() {
			resp, err := client.POST(postsEndpoint, map[string]interface{}{
				"title":  "Desayuno con diamantes",
				"author": "Truman Capote",
			})

			Convey(validResponse, func() {
				checkResponseIsValid(resp, err)
			})
		})
	})
}

func TestPUT(t *testing.T) {
	Convey(givenAClient, t, func() {
		client := getDefaultTestClient()

		Convey("When we make a valid PUT call", func() {
			resp, err := client.PUT(postsEndpoint+"/1", map[string]interface{}{
				"title":  "Desayuno con Diamantes",
				"author": "Truman Capote",
			})

			Convey(validResponse, func() {
				checkResponseIsValid(resp, err)
			})
		})
	})
}

func TestDELETE(t *testing.T) {
	Convey(givenAClient, t, func() {
		client := getDefaultTestClient()

		Convey("When we make a valid DELETE call", func() {
			resp, err := client.DELETE(postsEndpoint+"/1", nil)

			Convey(validResponse, func() {
				checkResponseIsValid(resp, err)
			})
		})
	})
}

func getDefaultTestClient() *api.Client {
	return api.MakeNewClient().WithBasePath(testBasePath).WithPort(3000)
}

func getTestHeaders() http.Header {
	headers := http.Header{}
	headers.Set("truman", "capote")
	headers.Set("capote", "truman")
	return headers
}

func getNewGinContextWithAuthBearer() (*gin.Context, string) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	request, err := http.NewRequest(http.MethodGet, "url", nil)
	if err != nil {
		panic(err)
	}
	request.Header.Set(authHeader, testAuthBearer)
	ctx.Request = request

	return ctx, testAuthBearer
}

func checkResponseIsValid(resp *http.Response, err error) {
	So(err, ShouldBeNil)
	So(resp.StatusCode, ShouldBeGreaterThanOrEqualTo, http.StatusOK)
	So(resp.StatusCode, ShouldBeLessThan, http.StatusBadRequest)
}
