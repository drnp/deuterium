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
 * @file handler.go
 * @package engine
 * author Dr.NP <conan.np@gmail.com>
 * @since 07/10/2020
 */

package engine

import "strings"

// UniformHandlerFunc : RPC / Task / Notify handler
type UniformHandlerFunc func(*UniformMessage) (*ResultMessage, error)

// UniformMsgHandler : Uniform message handler container
type UniformMsgHandler struct {
	method      string
	hdr         UniformHandlerFunc
	concurrency int64
}

var handlers = make(map[string]*UniformMsgHandler)

// RegisterHandler : Add handler to pool
func RegisterHandler(method string, handler UniformHandlerFunc, cv ...int64) {
	var (
		c int64
	)

	if method == "" {
		return
	}

	if len(cv) > 0 {
		c = cv[0]
	}

	handlers[strings.ToLower(method)] = &UniformMsgHandler{
		method:      strings.ToLower(method),
		hdr:         handler,
		concurrency: c,
	}

	return
}

// GetHandler : Get handler by given name (method)
func GetHandler(method string) *UniformMsgHandler {
	if method == "" {
		return nil
	}

	h := handlers[strings.ToLower(method)]

	return h
}

// SetConcurrency : Set concurrency of message handler
func SetConcurrency(method string, concurrency int64) {
	if method == "" {
		return
	}

	h := handlers[strings.ToLower(method)]
	if h != nil {
		h.concurrency = concurrency
	}

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
