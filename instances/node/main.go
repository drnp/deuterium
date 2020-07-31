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
 * @since 06/28/2020
 */

package main

import (
	"fmt"
	"time"

	"github.com/drnp/deuterium/engine"
)

const (
	// AppName : Application name
	AppName = "deuterium.skel.node"
)

func myName(msg *engine.UniformMessage) (*engine.ResultMessage, error) {
	type _data struct {
		ID int64 `msgpack:"id"`
	}

	type _resp struct {
		Name string `msgpack:"name"`
	}

	input := new(_data)
	resp := new(_resp)
	err := msg.Unmarshal(input)
	if err == nil {
		resp.Name = fmt.Sprintf("Name_%04d", input.ID)
	} else {
		resp.Name = "Unknown"
	}

	r := engine.NewResultMessage(resp, false)

	return r, nil
}

func smile(msg *engine.UniformMessage) (*engine.ResultMessage, error) {
	type _data struct {
		ID int64 `msgpack:"id"`
	}

	input := new(_data)
	msg.Unmarshal(input)
	fmt.Println(msg.Sender)
	time.Sleep(1 * time.Second)
	fmt.Println("Hehehe ...", input.ID)
	time.Sleep(1 * time.Second)
	fmt.Println("Hahaha ...", input.ID)
	time.Sleep(1 * time.Second)
	fmt.Println("Xixixi ...", input.ID)

	return nil, nil
}

func notify(msg *engine.UniformMessage) (*engine.ResultMessage, error) {
	fmt.Println(msg.Sender)
	fmt.Println("收到")

	return nil, nil
}

func main() {
	app := engine.NewApp(AppName)
	app.SetDefaultConfig(map[string]interface{}{
		"nsq.nsqd.addr": "localhost:4150",
		"nats.url":      "nats://localhost:4222",
		//"rpc.server.access_log": true,
	})

	engine.RegisterHandler("myName", myName)
	engine.RegisterHandler("smile", smile, 3)
	engine.RegisterHandler("notify", notify)
	rpc := engine.NewRPCServer()
	app.SetRPC(rpc)

	nsq := engine.NewNsqClient(app.Config().GetString("nsq.nsqd.addr"))
	app.SetNsq(nsq)

	nats := engine.NewNatsClient(app.Config().GetString("nats.url"))
	app.SetNats(nats)

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
