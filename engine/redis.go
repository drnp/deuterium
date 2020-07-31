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
 * @since 05/07/2020
 */

package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// RedisStreamReadBlockTime : Duration blocked
	RedisStreamReadBlockTime = 10 * time.Second
)

// NewRedis : Create redis client
/* {{{ [NewRedis] */
func NewRedis(addr, auth string, db int) (*redis.Client, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: auth,
		DB:       db,
	})
	if c != nil {
		return c, nil
	}

	return nil, fmt.Errorf("Create redis client to %s failed", addr)
}

/* }}} */

// RedisClose : Redis client disconnect
// TODO : Removal, no need close redis instance
/* {{{ [RedisClose] */
func RedisClose(r *redis.Client) error {
	return r.Close()
}

/* }}} */

// Helpers **

// QAdd : Add item to stream (XAdd)
/* {{{ [QAdd] */
func QAdd(r *redis.Client, key string, msg map[string]interface{}) string {
	if r == nil {
		return ""
	}

	args := &redis.XAddArgs{
		Stream: key,
		Values: msg,
	}

	strCMD := r.XAdd((context.Background()), args)
	qID, err := strCMD.Result()
	if err == nil {
		return qID
	}

	fmt.Println(err)

	return ""
}

/* }}} */

// QRead : Read item from stream (XRead)
/* {{{ [QRead] */
func QRead(r *redis.Client, key, last string) (string, map[string]interface{}) {
	if r == nil {
		return "", nil
	}

	if last == "" {
		last = "$"
	}

	args := &redis.XReadArgs{
		Streams: []string{key, last},
		Count:   1,
		Block:   RedisStreamReadBlockTime,
	}

	sliceCmd := r.XRead(context.Background(), args)
	x, err := sliceCmd.Result()
	if err == nil {
		if len(x) == 1 {
			xs := x[0]
			if len(xs.Messages) == 1 && xs.Stream == key {
				xm := xs.Messages[0]

				return xm.ID, xm.Values
			}
		}
	}

	return "", nil
}

/* }}} */

// QReadGroup : Read item from stream grouply (XReadGroup)
/* {{{ [QReadGroup] */
func QReadGroup(r *redis.Client, group, consumer, key, last string) (string, map[string]interface{}) {
	if r == nil {
		return "", nil
	}

	if last == "" {
		last = ">"
	}

	args := &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{key, last},
		Count:    1,
		Block:    RedisStreamReadBlockTime,
	}

	sliceCmd := r.XReadGroup(context.Background(), args)
	x, err := sliceCmd.Result()
	if err == nil {
		if len(x) == 1 {
			xs := x[0]
			if xs.Stream == key {
				if len(xs.Messages) >= 1 {
					xm := xs.Messages[0]

					return xm.ID, xm.Values
				}

				return ">", nil
			}
		}
	}

	return "", nil
}

/* }}} */

// QGroup : Create group (XGroup Create MKStream)
/* {{{ [Qgroup] */
func QGroup(r *redis.Client, group, key, id string) error {
	if r == nil {
		return fmt.Errorf("Null redis instance")
	}

	statusCmd := r.XGroupCreateMkStream(context.Background(), key, group, id)
	_, err := statusCmd.Result()
	if err != nil {
		return err
	}

	return nil
}

/* }}} */

// QDone : Send ACK back to queue
/* {{{ [QDone] */
func QDone(r *redis.Client, key, id string) int64 {
	if r == nil {
		return 0
	}

	intCmd := r.XDel(context.Background(), key, id)
	n, err := intCmd.Result()
	if err == nil {
		return n
	}

	return 0
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
