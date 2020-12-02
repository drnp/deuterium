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
 * @file mesage.go
 * @package engine
 * author Dr.NP <conan.np@gmail.com>
 * @since 06/28/2020
 */

package engine

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang/snappy"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack"
)

// Topic prefix
const (
	TaskTopicPrefix   = "_.task_"
	NotifyTopicPrefix = "_.notify_"
)

// DistinguishBranch
var (
	distinguishBranch bool
)

// SetDistinguishBranch : Distinguish branch append branch to remote address
func SetDistinguishBranch(d bool) {
	distinguishBranch = d
}

// UniformMessage : Uniform message type for task / RPC / broadcast
type UniformMessage struct {
	ID       string
	Sender   string
	Reciever string
	Method   string
	Compress bool
	Time     time.Time
	Data     []byte
}

// NewUniformMessage : Create new uniform message
func NewUniformMessage(data interface{}, compress bool) *UniformMessage {
	msg := new(UniformMessage)
	ba, err := msgpack.Marshal(data)
	if err == nil {
		if compress {
			// Snappy
			bac := snappy.Encode(nil, ba)
			msg.Compress = true
			msg.Data = bac
		} else {
			msg.Data = ba
		}
	}

	uuid, err := uuid.NewRandom()
	if err == nil {
		msg.ID = uuid.String()
	}

	msg.Time = time.Now()

	return msg
}

// Encode : Stringify
func (msg *UniformMessage) Encode() ([]byte, error) {
	return msgpack.Marshal(msg)
}

// Decode : Stream to message
func (msg *UniformMessage) Decode(data []byte) error {
	return msgpack.Unmarshal(data, msg)
}

// Bytes : Get raw data
func (msg *UniformMessage) Bytes() ([]byte, error) {
	if msg.Compress {
		return snappy.Decode(nil, msg.Data)
	}

	return msg.Data, nil
}

// Unmarshal : Unmarshal data
func (msg *UniformMessage) Unmarshal(v interface{}) error {
	raw, err := msg.Bytes()
	if err != nil {
		return err
	}

	return msgpack.Unmarshal(raw, v)
}

// Length : Data size
func (msg *UniformMessage) Length() int {
	if msg.Compress {
		cl, _ := snappy.DecodedLen(msg.Data)

		return cl
	}

	return len(msg.Data)
}

// Service-to-service methods

func _msgTarget(target string) string {
	if enableBranch == true {
		branch := os.Getenv(BranchEnvName)
		if branch == "" {
			branch = DefaultBranchValue
		}

		return target + "_" + branch
	}

	return target
}

// Call : Synchronously RPC via HTTP2
func (msg *UniformMessage) Call(reciever, method string) (*ResultMessage, error) {
	msg.Reciever = _msgTarget(reciever)
	msg.Method = method
	msg.Sender = App().Name
	payload, err := msg.Encode()
	if err != nil {
		// Encode failed
		return nil, err
	}

	client := NewRPCClient(msg.Reciever)
	r, err := client.Call(payload)
	if err != nil {
		Logger().Errorf("RPC call to <%s>:[%s] failed : %s", reciever, method, err.Error())

		return nil, err
	}

	Logger().Debugf("RPC call to <%s>:[%s] finished", reciever, method)
	if r != nil {
		return r, nil
	}

	return nil, fmt.Errorf("Empty RPC response body")
}

// Task : Asynchronously queue via NSQ
func (msg *UniformMessage) Task(target, method string) error {
	msg.Reciever = _msgTarget(target)
	msg.Method = method
	msg.Sender = _msgTarget(App().Name)
	payload, err := msg.Encode()
	if err != nil {
		return err
	}

	client := App().nsq
	if client == nil {
		return fmt.Errorf("No NSQ connection")
	}

	topic := fmt.Sprintf("%s%s", TaskTopicPrefix, msg.Reciever)
	err = client.Publish(topic, payload)
	if err != nil {
		Logger().Errorf("NSQ publish to <%s>:[%s] failed : %s", msg.Reciever, method, err.Error())
	} else {
		Logger().Debugf("NSQ publish to <%s>:[%s] finished", msg.Reciever, method)
	}

	return err
}

// Notify : Shout to all instances of service via MQTT
func (msg *UniformMessage) Notify(target, method string) error {
	msg.Reciever = _msgTarget(target)
	msg.Method = method
	msg.Sender = App().Name
	payload, err := msg.Encode()
	if err != nil {
		return nil
	}

	client := App().nats
	if client == nil {
		return fmt.Errorf("No NATS connection")
	}

	topic := fmt.Sprintf("%s%s", NotifyTopicPrefix, msg.Reciever)
	err = client.Publish(topic, payload)
	if err != nil {
		Logger().Errorf("NATS publish to <%s>:[%s] failed : %s", msg.Reciever, method, err.Error())
	} else {
		Logger().Debugf("NATS publish to <%s>:[%s] finished", msg.Reciever, method)
	}

	return err
}

// ResultMessage : Result
type ResultMessage struct {
	ID         string
	Code       int
	HTTPStatus int
	Message    string
	Compress   bool
	Time       time.Time
	Data       []byte
}

// NewResultMessage : Create result message
func NewResultMessage(data interface{}, compress bool) *ResultMessage {
	msg := new(ResultMessage)
	ba, err := msgpack.Marshal(data)
	if err == nil {
		if compress {
			// Snappy
			bac := snappy.Encode(nil, ba)
			msg.Compress = true
			msg.Data = bac
		} else {
			msg.Data = ba
		}
	}

	msg.Time = time.Now()
	msg.HTTPStatus = http.StatusOK

	return msg
}

// Encode : Stringify
func (msg *ResultMessage) Encode() ([]byte, error) {
	return msgpack.Marshal(msg)
}

// Decode : Stream to message
func (msg *ResultMessage) Decode(data []byte) error {
	return msgpack.Unmarshal(data, msg)
}

// Bytes : Get raw data
func (msg *ResultMessage) Bytes() ([]byte, error) {
	if msg.Compress {
		return snappy.Decode(nil, msg.Data)
	}

	return msg.Data, nil
}

// Unmarshal : Unmarshal data
func (msg *ResultMessage) Unmarshal(v interface{}) error {
	raw, err := msg.Bytes()
	if err != nil {
		return err
	}

	return msgpack.Unmarshal(raw, v)
}

// Length : Data size
func (msg *ResultMessage) Length() int {
	if msg.Compress {
		cl, _ := snappy.DecodedLen(msg.Data)

		return cl
	}

	return len(msg.Data)
}

/*
 * Local variables:
 * tab-width: 4
 * c-basic-offset: 4
 * End:
 * vim600: sw=4 ts=4 fdm=marker
 * vim<600: sw=4 ts=4
 */
