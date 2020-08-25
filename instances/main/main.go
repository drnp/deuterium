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
 * @file app.go
 * @package engine
 * author Dr.NP <conan.np@gmail.com>
 * @since 05/12/2020
 */

package main

import (
	"github.com/drnp/deuterium/engine"

	"github.com/valyala/fasthttp"
)

const (
	// AppName : Application name
	AppName = "deuterium.skel.main"
)

func serv(ctx *fasthttp.RequestCtx) {
	type _data struct {
		ID int64 `msgpack:"id"`
	}

	type _resp struct {
		Name string `json:"name" msgpack:"name"`
	}

	id := engine.HTTPArgInt(ctx, "id")
	e := engine.AcquireHTTPEnvelope()
	req := &_data{ID: id}
	resp := &_resp{}
	msg := engine.NewUniformMessage(req, false)
	for i := 0; i < 30; i++ {
		msg.Task("deuterium.skel.node", "smile")
	}

	msg.Notify("deuterium.skel.node", "notify")
	r, err := msg.Call("deuterium.skel.node", "myName")
	if err != nil {
		//engine.Logger().Error(err)
		e.Message = err.Error()
		e.HTTPStatus = fasthttp.StatusInternalServerError
		e.Code = 1023
	} else {
		r.Unmarshal(resp)
		e.Data = resp
	}

	engine.HTTPEnvelope(ctx, e)

	return
}

func main() {
	app := engine.NewApp(AppName)
	app.SetDefaultConfig(map[string]interface{}{
		"nsq.nsqd.addr": "localhost:4150",
		"nats.url":      "nats://localhost:4222",
		//"http.server.access_log": true,
	})

	http := engine.NewHTTPServer(":9080")
	http.SetRoutes(
		&engine.HTTPRoute{
			Name:    "TestServ",
			Method:  "GET",
			Path:    "/serv/{id}",
			Aliases: []string{"/serv"},
			Handler: serv,
		},
	)
	app.SetHTTP(http)

	metrics := engine.NewMetrics(":9081")
	app.SetMetrics(metrics)

	nsq := engine.NewNsqClient(app.Config().GetString("nsq.nsqd.addr"))
	app.SetNsq(nsq)

	nats := engine.NewNatsClient(app.Config().GetString("nats.url"))
	app.SetNats(nats)

	app.Metrics().Counter("test_counter").Add(100)
	engine.SetDistinguishBranch(true)

	app.Startup()

	return
}

/*
 * Local variables:
 * tab-width: 4
 * c-basic-offset: 4
 * End:
 * vim600: sw=4 ts=4 fdm=marker
 * vim<600: sw=4 ts=4
 */
