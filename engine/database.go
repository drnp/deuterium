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
	"github.com/sirupsen/logrus"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/mysql"
	"upper.io/db.v3/postgresql"
	"upper.io/db.v3/sqlite"
)

// NewDatabase : Create upper instance
/* {{{ [NewDatabase] */
func NewDatabase(dbtype, host, db, user, pass string) (sqlbuilder.Database, error) {
	var (
		conn sqlbuilder.Database
		err  error
	)

	switch dbtype {
	case "postgresql":
		settings := postgresql.ConnectionURL{
			Database: db,
			Host:     host,
			User:     user,
			Password: pass,
		}

		conn, err = postgresql.Open(settings)
	case "mysql":
		settings := mysql.ConnectionURL{
			Database: db,
			Host:     host,
			User:     user,
			Password: pass,
		}

		conn, err = mysql.Open(settings)
	case "sqlite":
		settings := sqlite.ConnectionURL{
			Database: db,
		}

		conn, err = sqlite.Open(settings)
	}
	return conn, err
}

/* }}} */

// dbLogger : Database logger implemetation
type dbLogger struct {
	logger *logrus.Entry
}

func (l *dbLogger) Log(q *db.QueryStatus) {
	l.logger.Print(q.Query)
}

/*
 * Local variables:
 * tab-width: 4
 * c-basic-offset: 4
 * End:
 * vim600: sw=4 ts=4 fdm=marker
 * vim<600: sw=4 ts=4
 */
