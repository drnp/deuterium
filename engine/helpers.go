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
 * @file helpers.go
 * @package engine
 * author Dr.NP <conan.np@gmail.com>
 * @since 05/12/2020
 */

package engine

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// appModeName : Stringify application mode
/* {{{ [appModeName] */
func appModeName(mode int) string {
	name := ""
	switch mode {
	case AppModeService:
		name = "Service"
	case AppModeCli:
		name = "Command line interface"
	case AppModeOther:
		name = "Other"
	case AppModeTest:
		name = "Test mode"
	default:
		name = "Unknown"
	}

	return name
}

/* }}} */

// logFormatter : Get logrus formatter
/* {{{ [logFormatter] */
func logFormatter(fmt string) logrus.Formatter {
	switch fmt {
	case LogFormatJSON:
		return &logrus.JSONFormatter{}
	default:
		return &logrus.TextFormatter{
			FullTimestamp: true,
		}
	}
}

/* }}} */

// ReleaseVersion : version info
type ReleaseVersion struct {
	Major   int
	Minor   int
	Release int
	Path    int
	Tag     string
}

// String : Stringify version
/* {{{ [ReleaseVersion::String] */
func (v ReleaseVersion) String() string {
	var output string
	output = fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Release)
	if v.Path > 0 {
		output = fmt.Sprintf("%s.%d", output, v.Path)
	}

	if v.Tag != "" {
		output = fmt.Sprintf("%s-%s", output, v.Tag)
	}

	return output
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
