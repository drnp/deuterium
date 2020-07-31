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
 * @file metrics.go
 * @package engine
 * author Dr.NP <conan.np@gmail.com>
 * @since 05/12/2020
 */

package engine

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const (
	// MetricsRoute : Route path of metrics
	MetricsRoute = "/metrics"
)

// Metric types
const (
	MetricTypeCounter   = 1
	MetricTypeGauge     = 2
	MetricTypeHistogram = 3
	MetricTypeSummary   = 4
)

// MetricsIns : Instance (HTTP server) of prometheus exporter
type MetricsIns struct {
	addr        string
	tls         bool
	sslCertFile string
	sslKeyFile  string
	server      *http.Server

	list       map[string]*Metric
	counters   map[string]prometheus.Counter
	gauges     map[string]prometheus.Gauge
	histograms map[string]prometheus.Histogram
	summaries  map[string]prometheus.Summary
}

// Metric : Metric value
type Metric struct {
	name  string
	help  string
	vType int
}

// NewMetrics : Create prometheus exporter node by given parameters
/* {{{ [NewMetrics] */
func NewMetrics(addr string) *MetricsIns {
	metrics := &MetricsIns{
		addr: addr,
		server: &http.Server{
			Addr: addr,
		},

		list:       make(map[string]*Metric),
		counters:   make(map[string]prometheus.Counter),
		gauges:     make(map[string]prometheus.Gauge),
		histograms: make(map[string]prometheus.Histogram),
		summaries:  make(map[string]prometheus.Summary),
	}

	return metrics
}

/* }}} */

// SetSSL : Set SSL cert & key for metrics node
/* {{{ [MetricsIns::SetSSL] */
func (metrics *MetricsIns) SetSSL(sslCertFile, sslKeyFile string) {
	if sslCertFile != "" && sslKeyFile != "" {
		metrics.tls = true
		metrics.sslCertFile = sslCertFile
		metrics.sslKeyFile = sslKeyFile
	}

	return
}

/* }}} */

// Startup : Start and serve
/* {{{ [MetricsIns::Startup] */
func (metrics *MetricsIns) Startup(logger *logrus.Entry) {
	mux := http.NewServeMux()
	mux.Handle(MetricsRoute, promhttp.Handler())
	go func() {
		var failed error
		metrics.server.Handler = mux
		if metrics.tls {
			// HTTPS
			logger.Printf("Prometheus exporter node initialized at [%s] with SSL", metrics.addr)
			failed = metrics.server.ListenAndServeTLS(metrics.sslCertFile, metrics.sslKeyFile)
		} else {
			logger.Printf("Prometheus exporter node initialized at [%s]", metrics.addr)
			failed = metrics.server.ListenAndServe()
		}

		if failed != nil {
			// Server closed
			// logger.Printf("HTTP server listen and serve failed : %s\n", failed.Error())
		}
	}()

	// Do not fly
	time.Sleep(100 * time.Microsecond)

	return
}

/* }}} */

// Shutdown : Graceful shutdown HTTP (node) server
/* {{{ [MetricsIns::Shutdown] */
func (metrics *MetricsIns) Shutdown() {
	metrics.server.Close()

	return
}

/* }}} */

// SetMetrics : Set metrics norm
/* {{{ [Metrics::SetMetrics] */
func (metrics *MetricsIns) SetMetrics(m []*Metric) error {
	if m == nil {
		return fmt.Errorf("Empty metric list")
	}

	for _, metric := range m {
		if metric.name == "" {
			continue
		}

		if metrics.list[metric.name] != nil {
			continue
		}

		switch metric.vType {
		case MetricTypeCounter:
			v := promauto.NewCounter(prometheus.CounterOpts{
				Name: metric.name,
				Help: metric.help,
			})
			metrics.counters[metric.name] = v
		case MetricTypeGauge:
			v := promauto.NewGauge(prometheus.GaugeOpts{
				Name: metric.name,
				Help: metric.help,
			})
			metrics.gauges[metric.name] = v
		case MetricTypeHistogram:
			v := promauto.NewHistogram(prometheus.HistogramOpts{
				Name: metric.name,
				Help: metric.help,
			})
			metrics.histograms[metric.name] = v
		case MetricTypeSummary:
			v := promauto.NewSummary(prometheus.SummaryOpts{
				Name: metric.name,
				Help: metric.help,
			})
			metrics.summaries[metric.name] = v
		default:
			// Unknown type
		}

		metrics.list[metric.name] = metric
	}

	return nil
}

/* }}} */

/* {{{ Getters */

// Counter : Get counter
func (metrics *MetricsIns) Counter(name string) prometheus.Counter {
	v, ok := metrics.counters[name]
	if !ok {
		v = promauto.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: "Casual counter",
		})
		metrics.counters[name] = v
		metrics.list[name] = &Metric{
			name:  name,
			vType: MetricTypeCounter,
		}
	}

	return v
}

// Gauge : Get gauge
func (metrics *MetricsIns) Gauge(name string) prometheus.Gauge {
	v, ok := metrics.gauges[name]
	if !ok {
		v = promauto.NewGauge(prometheus.GaugeOpts{
			Name: name,
			Help: "Casual gauge",
		})
		metrics.gauges[name] = v
		metrics.list[name] = &Metric{
			name:  name,
			vType: MetricTypeGauge,
		}
	}

	return v
}

// Histogram : Get histogram
func (metrics *MetricsIns) Histogram(name string) prometheus.Histogram {
	v, ok := metrics.histograms[name]
	if !ok {
		v = promauto.NewHistogram(prometheus.HistogramOpts{
			Name: name,
			Help: "Casual histogram",
		})
		metrics.histograms[name] = v
		metrics.list[name] = &Metric{
			name:  name,
			vType: MetricTypeHistogram,
		}
	}

	return v
}

// Summary : Get summary
func (metrics *MetricsIns) Summary(name string) prometheus.Summary {
	v, ok := metrics.summaries[name]
	if !ok {
		v = promauto.NewSummary(prometheus.SummaryOpts{
			Name: name,
			Help: "Casual summary",
		})
		metrics.summaries[name] = v
		metrics.list[name] = &Metric{
			name:  name,
			vType: MetricTypeSummary,
		}
	}

	return v
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
