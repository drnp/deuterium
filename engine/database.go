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
 * @file database.go
 * @package engine
 * author Dr.NP <conan.np@gmail.com>
 * @since 05/12/2020
 */

package engine

import (
	"strings"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
	"github.com/upper/db/v4/adapter/ql"
	"github.com/upper/db/v4/adapter/sqlite"
)

// NewDatabase : Create upper instance
/* {{{ [NewDatabase] */
func NewDatabase(dbtype, host, database, user, pass string, options ...map[string]string) (db.Session, error) {
	var (
		conn db.Session
		err  error
	)

	switch dbtype {
	case "postgresql":
		settings := postgresql.ConnectionURL{
			Database: database,
			Host:     host,
			User:     user,
			Password: pass,
		}
		if options != nil && options[0] != nil {
			settings.Options = options[0]
		}

		conn, err = postgresql.Open(settings)
	case "mysql":
		settings := mysql.ConnectionURL{
			Database: database,
			Host:     host,
			User:     user,
			Password: pass,
		}
		if options != nil && options[0] != nil {
			settings.Options = options[0]
		}

		conn, err = mysql.Open(settings)
	case "sqlite":
		settings := sqlite.ConnectionURL{
			Database: database,
		}
		if options != nil && options[0] != nil {
			settings.Options = options[0]
		}

		conn, err = sqlite.Open(settings)
	case "ql":
		settings := ql.ConnectionURL{
			Database: database,
		}
		if options != nil && options[0] != nil {
			settings.Options = options[0]
		}

		conn, err = ql.Open(settings)
	}

	return conn, err
}

/* }}} */

// SetDatabaseDebugLevel : Set debug level for upper globally
/* {{{ [SetDatabaseDebugLevel] */
func SetDatabaseDebugLevel(level string) {
	switch strings.ToLower(level) {
	case "trace":
		db.LC().SetLevel(db.LogLevelTrace)
	case "debug":
		db.LC().SetLevel(db.LogLevelDebug)
	case "info":
		db.LC().SetLevel(db.LogLevelInfo)
	case "warn":
		db.LC().SetLevel(db.LogLevelWarn)
	case "error":
		db.LC().SetLevel(db.LogLevelError)
	case "fatal":
		db.LC().SetLevel(db.LogLevelFatal)
	case "panic":
		db.LC().SetLevel(db.LogLevelPanic)
	default:
		// Do nothing
	}
}

/* }}} */

// SetDatabaseLogger : Set logger for upper globally
/* {{{ [SetDatabaseLogger] */
func SetDatabaseLogger(logger db.Logger) {
	db.LC().SetLogger(logger)
}

/* }}} */

/*
// dbLogger : Database logger implemetation
type dbLogger struct {
	logger *logrus.Entry
}

func (l *dbLogger) Log(q *db.QueryStatus) {
	l.logger.Print(q.Query)
}
*/

/*
 * Local variables:
 * tab-width: 4
 * c-basic-offset: 4
 * End:
 * vim600: sw=4 ts=4 fdm=marker
 * vim<600: sw=4 ts=4
 */
