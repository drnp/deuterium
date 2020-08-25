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
 * @file task.go
 * @package engine
 * author Dr.NP <conan.np@gmail.com>
 * @since 07/10/2020
 */

package engine

import (
	"fmt"

	"github.com/nsqio/go-nsq"
)

type _taskNsqConsumerHandler struct{}

func (th *_taskNsqConsumerHandler) HandleMessage(message *nsq.Message) error {
	msg := NewUniformMessage(nil, false)
	err := msg.Decode(message.Body)
	if err != nil {
		Logger().Error(err)

		return err
	}

	defer message.Finish()

	self := _msgTarget(App().Name)
	if msg.Reciever != self {
		// Not you?
		err = fmt.Errorf("Task : Wrong message reciever : Self <%s> / Reciever <%s>", self, msg.Reciever)
		Logger().Error(err)

		return err
	}

	h := GetHandler(msg.Method)
	if h == nil {
		err = fmt.Errorf("Task method handler <%s> not found", msg.Method)
		Logger().Error(err)

		return err
	}

	runHandler(h, msg)

	return nil
}

/*
 * Local variables:
 * tab-width: 4
 * c-basic-offset: 4
 * End:
 * vim600: sw=4 ts=4 fdm=marker
 * vim<600: sw=4 ts=4
 */
