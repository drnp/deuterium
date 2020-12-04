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
 * @file rpc.go
 * @package engine
 * author Dr.NP <conan.np@gmail.com>
 * @since 06/22/2020
 */

package engine

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	// RPCTCPPort : TCP port of RPC
	RPCTCPPort = 19080
)

// RPC result codes
const (
	RPCCodeOK = 0
)

// RPCServer : HTTP2 (H2C) server
type RPCServer struct {
	addr        string
	tls         bool
	sslCertFile string
	sslKeyFile  string
	server      *http.Server
	mux         *http.ServeMux
}

func defaultRPCMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			// No data
			Logger().Error(err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		msg := NewUniformMessage(nil, false)
		err = msg.Decode(body)
		if err != nil {
			// Msgpack failed
			Logger().Error(err)
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		self := _msgTarget(App().Name)
		if msg.Reciever != self {
			// Not you?
			Logger().Errorf("RPC : Wrong message reciever : Self <%s> / Reciever <%s>", self, msg.Reciever)
			w.WriteHeader(http.StatusNotAcceptable)

			return
		}

		h := GetHandler(msg.Method)
		if h == nil {
			Logger().Errorf("RPC method handler <%s> not found", msg.Method)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if Config().GetBool("rpc.server.access_log") {
			Logger().Debugf("RPC Access : <%s> from [%s]", msg.Method, msg.Sender)
		}

		// Ignore concurrency
		now0 := time.Now().UnixNano()
		ret, err := h.hdr(msg)
		now1 := time.Now().UnixNano()
		if Config().GetBool("rpc.server.access_log") {
			Logger().Debugf("RPC Access : <%s> from [%s], Elapsed time (nano seconds) : %d", msg.Method, msg.Sender, now1-now0)
		}

		if ret == nil {
			ret = NewResultMessage(nil, false)
		}

		if err != nil {
			Logger().Error(err)
			if ret.Message == "OK" {
				ret.Message = err.Error()
			}
		}

		if ret.HTTPStatus == 0 {
			ret.HTTPStatus = http.StatusOK
		}

		rba, err := ret.Encode()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.WriteHeader(ret.HTTPStatus)
		w.Write(rba)

		return
	})

	return mux
}

// NewRPCServer : Create HTTP2 (h2c) instance by given parameters
/* {{{ [NewRPCServer] */
func NewRPCServer() *RPCServer {
	addr := fmt.Sprintf(":%d", RPCTCPPort)
	server := &RPCServer{
		addr: addr,
		server: &http.Server{
			Addr: addr,
		},
	}

	return server
}

/* }}} */

// Startup : Start and serve
/* {{{ [RPCServer::Startup] */
func (s *RPCServer) Startup(logger *logrus.Entry) {
	if s.mux == nil {
		s.mux = defaultRPCMux()
	}

	go func() {
		var failed error
		if s.tls {
			s.server.Handler = s.mux
			http2.ConfigureServer(s.server, &http2.Server{})
			logger.Printf("RPC server initialized at [%s] with SSL", s.addr)
			failed = s.server.ListenAndServeTLS(s.sslCertFile, s.sslKeyFile)
		} else {
			s.server.Handler = h2c.NewHandler(s.mux, &http2.Server{})
			logger.Printf("RPC server initialized at [%s]", s.addr)
			failed = s.server.ListenAndServe()
		}

		if failed != nil {
			// Server closed
		}
	}()

	time.Sleep(100 * time.Microsecond)

	return
}

/* }}} */

// Shutdown : Shutdown RCP server
/* {{{ [RPCServer::Shutdown] */
func (s *RPCServer) Shutdown() {
	s.server.Close()

	return
}

/* }}} */

// RPCClient : HTTP2 (h2c) client
type RPCClient struct {
	addr   string
	client http.Client
}

// NewRPCClient : Create RPC (HTTP2) client
func NewRPCClient(addr string) *RPCClient {
	if addr == "" {
		addr = "localhost"
	}

	addr = fmt.Sprintf("%s:%d", addr, RPCTCPPort)
	client := &RPCClient{
		addr: addr,
		client: http.Client{
			// Ignore SSL (h2c)
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		},
	}

	return client
}

// Call : Call RPC
/* {{{ [RPCClient::Call] */
func (c *RPCClient) Call(payload []byte) (*ResultMessage, error) {
	resp, err := c.client.Post("http://"+c.addr, "application/msgpack", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	defer func() {
		resp.Body.Close()
		c.client.CloseIdleConnections()
	}()

	if resp.StatusCode < 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		r := NewResultMessage(nil, false)
		r.Decode(body)

		return r, nil
	}

	//Logger().Debugf("RPC call response with status %d", resp.StatusCode)

	return nil, fmt.Errorf("Invalid RPC call, response with HTTP status %d", resp.StatusCode)
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
