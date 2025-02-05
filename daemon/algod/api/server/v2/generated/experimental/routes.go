// Package experimental provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/algorand/oapi-codegen DO NOT EDIT.
package experimental

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	. "github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated/model"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Returns OK if experimental API is enabled.
	// (GET /v2/experimental)
	ExperimentalCheck(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// ExperimentalCheck converts echo context to params.
func (w *ServerInterfaceWrapper) ExperimentalCheck(ctx echo.Context) error {
	var err error

	ctx.Set(Api_keyScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ExperimentalCheck(ctx)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface, m ...echo.MiddlewareFunc) {
	RegisterHandlersWithBaseURL(router, si, "", m...)
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string, m ...echo.MiddlewareFunc) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/v2/experimental", wrapper.ExperimentalCheck, m...)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"ZDeYzPHh0/8LAAD//+8NMlAJugAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
