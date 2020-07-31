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
 * @since 07/29/2020
 */

package engine

import (
	"sync"
	"sync/atomic"
)

var (
	scheduleChans    = make(map[string]chan *UniformMessage)
	scheduleRoutines = make(map[string]bool)
)

func _rtSchedule(h *UniformMsgHandler, ch chan *UniformMessage) {
	var (
		running int64
		wg      sync.WaitGroup
	)

	for msg := range ch {
		if msg == nil {
			break
		}

		// Task
		atomic.AddInt64(&running, 1)
		if running >= h.concurrency {
			wg.Add(1)
		}

		go func(msg *UniformMessage) {
			_, err := h.hdr(msg)
			if err != nil {
				Logger().Error(err)
			}

			if running >= h.concurrency {
				wg.Done()
			}

			atomic.AddInt64(&running, -1)
		}(msg)

		wg.Wait()
	}

	return
}

// runHandler : Run handler with suitable method
func runHandler(h *UniformMsgHandler, msg *UniformMessage) {
	if h == nil {
		return
	}

	if h.concurrency > 0 {
		// Buffered
		/*
			scheduler := fmt.Sprintf("%s@%d", h.method, h.concurrency)
			ch := scheduleChans[scheduler]
			if ch == nil {
				ch = make(chan *UniformMessage, h.concurrency)
				scheduleChans[scheduler] = ch
			}

			// Try routine
			if !scheduleRoutines[scheduler] {
				go _rtSchedule(h, ch)
			}

			ch <- msg
		*/
		ch := scheduleChans[h.method]
		if ch == nil {
			ch = make(chan *UniformMessage)
			scheduleChans[h.method] = ch
			go _rtSchedule(h, ch)
		}

		ch <- msg
	} else if h.concurrency == 0 {
		// Blocking
		_, err := h.hdr(msg)
		if err != nil {
			Logger().Error(err)
		}
	} else {
		// All passthru
		go func() {
			_, err := h.hdr(msg)
			if err != nil {
				Logger().Error(err)
			}
		}()
	}
}

/*
 * Local variables:
 * tab-width: 4
 * c-basic-offset: 4
 * End:
 * vim600: sw=4 ts=4 fdm=marker
 * vim<600: sw=4 ts=4
 */
