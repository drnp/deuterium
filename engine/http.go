/*
 * MIT License
 *
 * Copyright (c) [year] [fullname]
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

/**
 * @file http.go
 * @package engine
 * author Dr.NP <conan.np@gmail.com>
 * @since 05/12/2020
 */

package engine

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

const (
	// HTTPServerMaxRequestBodySize : 1GB here
	HTTPServerMaxRequestBodySize = 1024 * 1024 * 1024
	// HTTPServerReadTimeout : 30 seconds
	HTTPServerReadTimeout = 30 * time.Second
)

/* {{{ [HTTPServer] */

// HTTPServer : Fasthttp server
type HTTPServer struct {
	addr        string
	tls         bool
	sslCertFile string
	sslKeyFile  string
	server      *fasthttp.Server
	router      *router.Router
	routes      []*HTTPRoute
}

// HTTPRoute : Route for fasthttprouter
type HTTPRoute struct {
	Name        string
	Description string
	Method      string
	Path        string
	Aliases     []string
	Handler     fasthttp.RequestHandler
	Permissions uint64
	Middlewares []string
}

// NewHTTPServer : Create fasthttp server by given parameters
/* {{{ [NewHTTPServer] */
func NewHTTPServer(addr string) *HTTPServer {
	f := &fasthttp.Server{
		MaxRequestBodySize: HTTPServerMaxRequestBodySize,
		ReadTimeout:        HTTPServerReadTimeout,
	}
	r := router.New()
	r.OPTIONS("/*_all", mwCors)
	server := &HTTPServer{
		addr:   addr,
		server: f,
		router: r,
	}

	return server
}

/* }}} */

// SetSSL : Set SSL cert & key for HTTP server
/* {{{ [HTTPServer::SetSSL] */
func (s *HTTPServer) SetSSL(sslCertFile, sslKeyFile string) {
	if sslCertFile != "" && sslKeyFile != "" {
		s.tls = true
		s.sslCertFile = sslCertFile
		s.sslKeyFile = sslKeyFile
	}
}

/* }}} */

// Startup : Start and serve
/* {{{ [HTTPServer::Startup] */
func (s *HTTPServer) Startup(logger fasthttp.Logger) {
	if logger != nil && s.server.Logger == nil {
		s.server.Logger = logger
	}

	s.server.Handler = s.router.Handler
	go func() {
		var failed error
		if s.tls == true {
			// HTTPS
			if s.server.Logger != nil {
				s.server.Logger.Printf("HTTP server initialized at [%s] with SSL", s.addr)
			}

			failed = s.server.ListenAndServeTLS(s.addr, s.sslCertFile, s.sslKeyFile)
		} else {
			// Normal HTTP
			if s.server.Logger != nil {
				s.server.Logger.Printf("HTTP server initialized at [%s]", s.addr)
			}

			failed = s.server.ListenAndServe(s.addr)
		}

		if s.server.Logger != nil {
			if failed != nil {
				s.server.Logger.Printf("HTTP server listen and serve failed : %s", failed.Error())
			}
		}

		//waiter.Done()
	}()

	// Do not fly
	time.Sleep(100 * time.Microsecond)

	return
}

/* }}} */

// Shutdown : Graceful stop HTTP server
/* {{{ [HTTPServer::Shutdown] */
func (s *HTTPServer) Shutdown() {
	s.server.Shutdown()
}

/* }}} */

// SetName : Set name (either in response header) of HTTP server
/* {{{ [HTTPServer::SetName] */
func (s *HTTPServer) SetName(name string) {
	if name != "" {
		s.server.Name = name
	}
}

/* }}} */

// SetRoutes : Set HTTP routes to server
/* {{{ [HTTPServer::SetRoutes] */
func (s *HTTPServer) SetRoutes(routes ...*HTTPRoute) error {
	s.routes = append(s.routes, routes...)

	return nil
}

/* }}} */

// loadRoutes : Load routes into router
/* {{{ [HTTPServer::loadRoutes] */
func (s *HTTPServer) loadRoutes() {
	for _, route := range s.routes {
		if route.Path == "" || route.Handler == nil {
			continue
		}

		h := route.Handler
		// AccessLog
		if App().Config().GetBool("http.server.access_log") {
			h = mwAccessLog(h)
		}

		uris := []string{route.Path}
		uris = append(uris, route.Aliases...)
		for _, uri := range uris {
			switch strings.ToLower(route.Method) {
			case "post":
				s.router.POST(uri, h)
			case "delete":
				s.router.DELETE(uri, h)
			case "put":
				s.router.PUT(uri, h)
			case "options":
				// BUGS(np) : CORS only
				s.router.OPTIONS(uri, h)
			case "patch":
				s.router.PATCH(uri, h)
			case "head":
				s.router.HEAD(uri, h)
			default:
				// Get for default
				s.router.GET(uri, h)
			}

			App().Logger().Debugf("Load route <%s> as path <%s> with method %s", route.Name, uri, route.Method)
		}
	}

	return
}

/* }}} */

/* {{{ [HTTPMiddlewares] */

// mwCors : CORS
func mwCors(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Accept, X-Version")
	ctx.Response.Header.Set("Access-Control-Max-Age", "86400")

	return
}

// mwAccessLog : Access log
func mwAccessLog(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		App().Logger().Debugf("HTTP Access : %s, Method: %s, Request body : %d bytes from <%s>", ctx.URI().String(), ctx.Method(), len(ctx.Request.Body()), ctx.RemoteAddr().String())
		if App().Debug {
			fmt.Println("====== Debug : Request body ======")
			fmt.Println(string(ctx.Request.Body()))
			fmt.Println("====== Debug ======")
		}

		now0 := time.Now().UnixNano()
		h(ctx)
		now1 := time.Now().UnixNano()
		App().Logger().Debugf("HTTP Access : %s, Method: %s, Response body : %d bytes, Status code : %d, Elapsed time (nano seconds) : %d", ctx.URI().String(), ctx.Method(), len(ctx.Response.Body()), ctx.Response.StatusCode(), now1-now0)
		if App().Debug {
			fmt.Println("====== Debug : Response body ======")
			fmt.Println(string(ctx.Response.Body()))
			fmt.Println("====== Debug ======")
		}

		return
	})
}

/* }}} */

// HTTPResponseEnvelope : Response body envelope
type HTTPResponseEnvelope struct {
	Code           int             `json:"code" xml:"code"`
	HTTPStatus     int             `json:"http_status" xml:"http_status"`
	StartTimestamp int64           `json:"start_timestamp" xml:"start_timestamp"`
	EndTimestamp   int64           `json:"end_timestamp" xml:"end_timestamp"`
	ElapsedTime    int64           `json:"elapsed_time" xml:"elapsed_time"`
	Message        string          `json:"message,omitempty" xml:"message"`
	ErrorPrompt    string          `json:"error_prompt,omitempty" xml:"error_prompt,omitempty"`
	Links          []string        `json:"linkes,omitempty" xml:"links"`
	Pagination     *HTTPPagination `json:"pagination,omitempty" xml:"pagination,omitempty"`
	Data           interface{}     `json:"data" xml:"data"`
}

// HTTPPagination : Pagination variables
type HTTPPagination struct {
	TotalEntries   int64      `json:"total_entries" xml:"total_entries"`
	Current        int        `json:"current" xml:"current"`
	EntriesPerPage int        `json:"entries_per_page" xml:"entries_per_page"`
	Start          int        `json:"start" xml:"start"`
	OrderBy        string     `json:"order_by,omitemtpy" xml:"order_by,omitempty"`
	Conditions     [][]string `json:"-" xml:"-"`
	WhereStr       string     `json:"-" xml:"-"`
	Desc           bool       `json:"-" xml:"-"`
	All            bool       `json:"-" xml:"-"`
}

const (
	paginationStart   = "start"
	paginationEnd     = "end"
	paginationPerPage = "n"
	paginationPage    = "p"
	paginationAll     = "_all"
	defaultPerPage    = 20
)

func parseRequestBodyField(midField map[string]interface{}, field []string) []string {
	for k := range midField {
		field = append(field, k)
	}

	return field
}

// HTTPParseRequestBody : Parse body by content-type in request header
/* {{{ [HTTPParseRequestBody] */
func HTTPParseRequestBody(ctx *fasthttp.RequestCtx, obj interface{}) ([]string, error) {
	var err error
	var field []string
	midField := make(map[string]interface{})
	contentType := strings.ToLower(string(ctx.Request.Header.ContentType()))
	if strings.HasPrefix(contentType, "application/xml") {
		// XML
		err = xml.Unmarshal(ctx.Request.Body(), obj)
		if err == nil {
			err = xml.Unmarshal(ctx.Request.Body(), &midField)
			field = parseRequestBodyField(midField, field)
		}
	} else if strings.HasPrefix(contentType, "application/json") {
		// JSON
		err = json.Unmarshal(ctx.Request.Body(), obj)
		if err == nil {
			err = json.Unmarshal(ctx.Request.Body(), &midField)
			field = parseRequestBodyField(midField, field)
		}
	} else {
		// Form data
		args := ctx.Request.PostArgs()
		t := reflect.TypeOf(obj)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if t.Kind() == reflect.Struct {
			ins := reflect.ValueOf(obj).Elem()
			for i := 0; i < t.NumField(); i++ {
				tt := t.Field(i)
				f := ins.FieldByName(tt.Name)
				if f.IsValid() && f.CanSet() {
					pname := tt.Tag.Get("post")
					if pname == "" {
						pname = tt.Name
					}

					if args.Has(pname) {
						field = append(field, pname)
						switch f.Kind() {
						case reflect.Bool:
							f.SetBool(args.GetBool(pname))
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							f.SetInt(int64(args.GetUintOrZero(pname)))
						case reflect.Float32, reflect.Float64:
							f.SetFloat(args.GetUfloatOrZero(pname))
						case reflect.String:
							f.SetString(string(args.Peek(pname)))
						}
					}
				}
			}
		}
	}

	return field, err
}

/* }}} */

// AcquireHTTPEnvelope : Generate HTTP envelope
/* {{{ [AcquireHTTPEnvelope] */
func AcquireHTTPEnvelope() *HTTPResponseEnvelope {
	return &HTTPResponseEnvelope{
		Message:        "OK",
		HTTPStatus:     fasthttp.StatusOK,
		StartTimestamp: time.Now().UnixNano(),
	}
}

/* }}} */

// HTTPEnvelope : Build envelope to response
/* {{{ [HTTPEnvelope] */
func HTTPEnvelope(ctx *fasthttp.RequestCtx, e *HTTPResponseEnvelope) error {
	var (
		accept string
		body   []byte
		err    error
	)

	if e.EndTimestamp == 0 {
		e.EndTimestamp = time.Now().UnixNano()
	}

	if e.HTTPStatus == 0 {
		e.HTTPStatus = fasthttp.StatusOK
	}

	/*
		if e.Code != 0 && e.Message != "" {
			Logger().Error(e.Message)
		}
	*/

	e.ElapsedTime = e.EndTimestamp - e.StartTimestamp

	ctx.SetStatusCode(e.HTTPStatus)
	ctx.ResetBody()

	// Parse Accept header
	accepts := make(map[string]bool)
	parts := strings.Split(string(ctx.Request.Header.Peek("Accept")), ",")
	if len(parts) > 0 {
		for _, t0 := range parts {
			ts := strings.Split(t0, ";")
			if ts != nil && len(t0) > 0 {
				accept = strings.ToLower(strings.TrimSpace(ts[0]))
				accepts[accept] = true
				if accept == "*/*" {
					accepts[Config().GetString("http.default_accept")] = true
				}
			}
		}
	}

	// BUG : Hack here ~ zhanghao05 12/03/2019
	//delete(accepts, "application/xml")
	accepts["application/json"] = true
	if accepts["application/xml"] == true {
		ctx.Response.Header.Set("Content-Type", "application/xml")
		body, err = xml.Marshal(e)
		if err == nil {
			ctx.Write(body)
		}
	} else if accepts["application/json"] == true {
		ctx.Response.Header.Set("Content-Type", "application/json")
		body, err = json.Marshal(e)
		if err == nil {
			ctx.Write(body)
		}
	} else {
		ctx.Response.Header.Set("Content-Type", "text/plain")
		ctx.WriteString("Envelope")
		ctx.Write(body)
	}

	return err
}

/* }}} */

// HTTPUnEnvelope : Unmarshal envelope from context
/* {{{ [HTTPUnEnvelope] */
func HTTPUnEnvelope(ctx *fasthttp.RequestCtx, e *HTTPResponseEnvelope) error {
	if e == nil {
		return fmt.Errorf("Empty envelope")
	}

	var (
		err error
	)

	switch string(ctx.Response.Header.Peek("Content-Type")) {
	case "application/xml":
		err = xml.Unmarshal(ctx.Response.Body(), e)
	case "application/json":
		err = json.Unmarshal(ctx.Response.Body(), e)
	default:
	}

	return err
}

/* }}} */

/* {{{ [HTTPArg*] */

// HTTPArgString : Get string value from arguments by given key
func HTTPArgString(ctx *fasthttp.RequestCtx, key string) string {
	// UserValue
	uv := ctx.UserValue(key)
	uvStr, ok := uv.(string)
	if ok {
		return uvStr
	}

	// QueryString
	args := ctx.QueryArgs()
	qb := args.Peek(key)
	if qb != nil {
		return string(qb)
	}

	// Query body
	var (
		body = map[string]interface{}{
			key: nil,
		}
		bv interface{}
	)

	switch strings.ToLower(string(ctx.Request.Header.ContentType())) {
	case "application/json":
		json.Unmarshal(ctx.Request.Body(), &body)
		bv = body[key]
	case "application/xml":
		xml.Unmarshal(ctx.Request.Body(), &body)
		bv = body[key]
	default:
		args = ctx.PostArgs()
		bvb := args.Peek(key)
		if bvb != nil {
			return string(bvb)
		}
	}

	bvStr, ok := bv.(string)
	if ok {
		return bvStr
	}

	return ""
}

// HTTPArgInt : Get int value from arguments by given key
func HTTPArgInt(ctx *fasthttp.RequestCtx, key string) int64 {
	// UserValue
	uv := ctx.UserValue(key)
	uvStr, ok := uv.(string)
	if ok {
		uvInt, _ := strconv.ParseInt(uvStr, 10, 0)
		return uvInt
	}

	// QueryString
	args := ctx.QueryArgs()
	qb := args.Peek(key)
	if qb != nil {
		uvInt, _ := strconv.ParseInt(string(qb), 10, 0)
		return uvInt
	}

	// Query body
	var (
		body = map[string]interface{}{
			key: nil,
		}
		bv interface{}
	)

	switch strings.ToLower(string(ctx.Request.Header.ContentType())) {
	case "application/json":
		json.Unmarshal(ctx.Request.Body(), &body)
		bv = body[key]
	case "application/xml":
		xml.Unmarshal(ctx.Request.Body(), &body)
		bv = body[key]
	default:
		args = ctx.PostArgs()
		bvb := args.Peek(key)
		if bvb != nil {
			bvInt, _ := strconv.ParseInt(string(bvb), 10, 0)
			return bvInt
		}
	}

	switch bv.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return bv.(int64)
	case float32, float64:
		return int64(bv.(float64))
	}

	return 0
}

// HTTPArgBool : Get bool value from arguments by given key
func HTTPArgBool(ctx *fasthttp.RequestCtx, key string) bool {
	// UserValue
	uv := ctx.UserValue(key)
	uvStr, ok := uv.(string)
	if ok {
		uvBool, _ := strconv.ParseBool(uvStr)
		return uvBool
	}

	// QueryString
	args := ctx.QueryArgs()
	qb := args.Peek(key)
	if qb != nil {
		uvBool, _ := strconv.ParseBool(string(qb))
		return uvBool
	}

	// Query body
	var (
		body = map[string]interface{}{
			key: nil,
		}
		bv interface{}
	)

	switch strings.ToLower(string(ctx.Request.Header.ContentType())) {
	case "application/json":
		json.Unmarshal(ctx.Request.Body(), &body)
		bv = body[key]
	case "application/xml":
		xml.Unmarshal(ctx.Request.Body(), &body)
		bv = body[key]
	default:
		args = ctx.PostArgs()
		bvb := args.Peek(key)
		if bvb != nil {
			bvBool, _ := strconv.ParseBool(string(bvb))
			return bvBool
		}
	}

	bvBool, ok := bv.(bool)
	if ok {
		return bvBool
	}

	return false
}

/* }}} */

/*
 * Local variables:
 * tab-width: 4
 * c-basic-offset: 4
 * End:
 * vim600: sw=4 ts=4 fdm=marker
 * vim<600: sw=4 ts=4
 */
