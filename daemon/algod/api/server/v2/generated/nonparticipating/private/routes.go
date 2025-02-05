// Package private provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/algorand/oapi-codegen DO NOT EDIT.
package private

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	. "github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated/model"
	"github.com/algorand/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Aborts a catchpoint catchup.
	// (DELETE /v2/catchup/{catchpoint})
	AbortCatchup(ctx echo.Context, catchpoint string) error
	// Starts a catchpoint catchup.
	// (POST /v2/catchup/{catchpoint})
	StartCatchup(ctx echo.Context, catchpoint string) error

	// (POST /v2/shutdown)
	ShutdownNode(ctx echo.Context, params ShutdownNodeParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// AbortCatchup converts echo context to params.
func (w *ServerInterfaceWrapper) AbortCatchup(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "catchpoint" -------------
	var catchpoint string

	err = runtime.BindStyledParameterWithLocation("simple", false, "catchpoint", runtime.ParamLocationPath, ctx.Param("catchpoint"), &catchpoint)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter catchpoint: %s", err))
	}

	ctx.Set(Api_keyScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.AbortCatchup(ctx, catchpoint)
	return err
}

// StartCatchup converts echo context to params.
func (w *ServerInterfaceWrapper) StartCatchup(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "catchpoint" -------------
	var catchpoint string

	err = runtime.BindStyledParameterWithLocation("simple", false, "catchpoint", runtime.ParamLocationPath, ctx.Param("catchpoint"), &catchpoint)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter catchpoint: %s", err))
	}

	ctx.Set(Api_keyScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.StartCatchup(ctx, catchpoint)
	return err
}

// ShutdownNode converts echo context to params.
func (w *ServerInterfaceWrapper) ShutdownNode(ctx echo.Context) error {
	var err error

	ctx.Set(Api_keyScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params ShutdownNodeParams
	// ------------- Optional query parameter "timeout" -------------

	err = runtime.BindQueryParameter("form", true, false, "timeout", ctx.QueryParams(), &params.Timeout)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter timeout: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ShutdownNode(ctx, params)
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

	router.DELETE(baseURL+"/v2/catchup/:catchpoint", wrapper.AbortCatchup, m...)
	router.POST(baseURL+"/v2/catchup/:catchpoint", wrapper.StartCatchup, m...)
	router.POST(baseURL+"/v2/shutdown", wrapper.ShutdownNode, m...)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
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
