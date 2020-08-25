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
 * @file nsq.go
 * @package engine
 * author Dr.NP <conan.np@gmail.com>
 * @since 06/22/2020
 */

package engine

import (
	"fmt"

	"github.com/nsqio/go-nsq"
)

// NsqClient : NSQ client
type NsqClient struct {
	addr      string
	config    *nsq.Config
	consumers []*nsq.Consumer
	producer  *nsq.Producer
}

type _defaultNsqConsumerHandler struct{}

func (h *_defaultNsqConsumerHandler) HandleMessage(message *nsq.Message) error {
	fmt.Println(message.ID)
	fmt.Printf("%v\n", message.Body)

	return nil
}

// NewNsqClient : Create NSQ client
/* {{{ [NewNsq] */
func NewNsqClient(addr string) *NsqClient {
	nc := &NsqClient{
		addr:      addr,
		config:    nsq.NewConfig(),
		consumers: make([]*nsq.Consumer, 0),
	}

	return nc
}

/* }}} */

// Publish : Publish data via producer
/* {{{ [NsqClient::Publish] */
func (n *NsqClient) Publish(topic string, msg []byte) error {
	var err error
	if n.producer == nil {
		n.producer, err = nsq.NewProducer(n.addr, n.config)
		if err != nil {
			return err
		}
	}

	err = n.producer.Publish(topic, msg)

	return err
}

/* }}} */

// Subscribe : Subscribe from nsq
/* {{{ [NsqClient::Subscribe] */
func (n *NsqClient) Subscribe(topic, channel string, handler nsq.Handler, concurrency int) error {
	var (
		err      error
		consumer *nsq.Consumer
	)

	consumer, err = nsq.NewConsumer(topic, channel, n.config)
	if err != nil {
		return err
	}

	if handler == nil {
		handler = &_defaultNsqConsumerHandler{}
	}

	if concurrency > 1 {
		//Logger().Debugf("Add consumer with concurrency %d", concurrency)
		consumer.AddConcurrentHandlers(handler, concurrency)
	} else {
		//Logger().Debug("Add single consumer")
		consumer.AddHandler(handler)
	}

	err = consumer.ConnectToNSQD(n.addr)
	if err != nil {
		return err
	}

	n.consumers = append(n.consumers, consumer)

	return err
}

/* }}} */

// Shutdown : Shutdown nsq client
/* {{{ [NsqClient::Shutdown] */
func (n *NsqClient) Shutdown() {
	if n.producer != nil {
		n.producer.Stop()
	}

	for _, c := range n.consumers {
		if c != nil {
			c.Stop()
		}
	}

	return
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
