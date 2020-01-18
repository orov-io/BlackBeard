# Go API Client

> A client to make easy REST API calls from go. provides convenience methods for working with services that expose a JSON interface.

## Installation

To use the api-client:

``` bash
go get github.com/orov-io/BlackBeard
```

## How to use

1. Import the library

    ```go
    import "github.com/orov-io/BlackBeard"
    ```

2. Initialize the client

    ```go
    client := api.MakeNewClient()
    ```

3. You can use convenience methods in order to set some parameters

    ```go
    // setting the auth header
    bearer := "bearer myJWTToken"
    client.WithAuthHeader(bearer)

    // setting the path to make calls
    path := "http://localhost:3000"
    client.WithBasePath(path)

    // setting custom headers
    headers := http.Header{}
    headers.Set("truman", "capote")
    headers.Set("capote", "truman")
    client.WithHeaders(headers)
    ```

    or, all in one:

    ```go
    // setting the auth header
    bearer := "bearer myJWTToken"
    path := "http://localhost:3000"
    headers := http.Header{}
    headers.Set("truman", "capote")
    headers.Set("capote", "truman")

    // setting the path to make calls
    path := "http://localhost:3000"
    client.WithAuthHeader(bearer).WithBasePath(path).WithHeaders(headers)
    ```

## Running test

This package relies on a basic [json-server](https://github.com/typicode/json-server) to make the api call test. You need to install it before running test:

```bash
npm install -g json-server
```
