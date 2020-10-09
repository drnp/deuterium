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
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/go-redis/redis/v8"
	nats "github.com/nats-io/nats.go"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/upper/db/v4"
)

// Application modes
const (
	AppModeService = 1
	AppModeCli     = 2
	AppModeRunner  = 3
	AppModeTest    = 4
	AppModeOther   = 255
)

// Log formats
const (
	LogFormatText = "text"
	LogFormatJSON = "json"
)

// Keys
const (
	logFieldAppName = "_app"
)

var (
	apps    = make(map[string]*AppIns)
	appName string
)

// RunnerWorker : Worker function
type RunnerWorker func(int)

// Interaction : Interactive cli properties
type Interaction struct {
	prompt  string
	color   string
	quitCmd string
}

// AppIns : Application defination
type AppIns struct {
	Name    string
	config  *viper.Viper
	logger  *logrus.Logger
	waiter  *sync.WaitGroup
	cronner *cron.Cron
	http    *HTTPServer
	metrics *MetricsIns
	rpc     *RPCServer
	db      db.Session
	redis   *redis.Client
	nsq     *NsqClient
	nats    *nats.Conn

	//mode    int
	goProcs int
	running bool

	//workerFunc RunnerWorker
	nWorkers int

	LogLevel  logrus.Level
	LogFormat string
	Debug     bool
}

// NewApp : Create new application instance
/* {{{ [NewApp] - Create new application instance */
func NewApp(name string) *AppIns {
	var (
		w sync.WaitGroup
	)

	appName = strings.ToLower(name)
	_defaultCronnerInstance = cron.New(cron.WithParser(cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))
	_defaultLoggerInstance = logrus.New()
	_defaultConfigInstance = viper.New()

	_defaultConfigInstance.SetConfigName(appName)
	_defaultConfigInstance.SetEnvPrefix(appName)
	_defaultConfigInstance.AutomaticEnv()
	_defaultConfigInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	_defaultConfigInstance.AddConfigPath(fmt.Sprintf("$HOME/.%s", appName))
	_defaultConfigInstance.AddConfigPath(fmt.Sprintf("/etc/%s", appName))
	_defaultConfigInstance.AddConfigPath(".")

	app := &AppIns{
		Name:      appName,
		cronner:   _defaultCronnerInstance,
		logger:    _defaultLoggerInstance,
		config:    _defaultConfigInstance,
		waiter:    &w,
		goProcs:   runtime.NumCPU() * 4,
		running:   true,
		LogLevel:  logrus.DebugLevel,
		LogFormat: LogFormatText,
		nWorkers:  1,
	}
	_defaultAppInstance = app
	apps[appName] = app
	app.logger.SetLevel(app.LogLevel)

	return app
}

/* }}} */

// Silence : Disable logging
/* {{{ [AppIns::Silence] */
func (app *AppIns) Silence() {
	app.logger.SetOutput(ioutil.Discard)

	return
}

/* }}} */

// Startup : Startup app
/* {{{ [AppIns::Startup] */
func (app *AppIns) Startup() {
	runtime.GOMAXPROCS(app.goProcs)

	sigQuit := make(chan os.Signal, 1)
	sigReload := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)
	/*
		// TODO : POSIX only. Windows does not support 'SIGSUR2'
		signal.Notify(sigReload, syscall.SIGUSR2)
	*/

	go func() {
		for {
			// Waiting for signals
			select {
			case <-sigQuit:
				// Quit application
				fmt.Println()
				app.Shutdown()
			case <-sigReload:
				// Reload (configurations)
				fmt.Println()
				app.ReloadConfig()
			}
		}
	}()

	app.logger.SetFormatter(logFormatter(app.LogFormat))
	app.Logger().Infof("Application <%s> startup", app.Name)
	if app.redis != nil {
		_, err := app.redis.Ping(context.Background()).Result()
		if err != nil {
			app.Logger().Errorf("Redis instance connection error : %s", err.Error())
		} else {
			app.Logger().Info("Redis instance connected")
		}
	}

	if app.db != nil {
		//dl := &dbLogger{logger: app.Logger()}
		//app.db.SetLogger(dl)
		err := app.db.Ping()
		if err != nil {
			app.Logger().Errorf("Database instance connection error : %s", err.Error())
		} else {
			app.Logger().Info("Database instance connected")
		}
	}

	if app.cronner != nil {
		app.cronner.Start()
	}

	if app.nsq != nil {
		// Subscribe
		topic := fmt.Sprintf("%s%s", TaskTopicPrefix, _msgTarget(app.Name))
		err := app.nsq.Subscribe(topic, TaskTopicPrefix, &_taskNsqConsumerHandler{}, app.nWorkers)
		if err != nil {
			app.logger.Error(err)
		} else {
			app.logger.Debugf("NSQ subscribed to <%s>", topic)
		}
	}

	if app.nats != nil {
		// Subscribe
		topic := fmt.Sprintf("%s%s", NotifyTopicPrefix, _msgTarget(app.Name))
		_, err := app.nats.Subscribe(topic, _notifyNatsConsumerHandler)
		if err != nil {
			app.logger.Error(err)
		} else {
			app.logger.Debugf("NATS subscribed to <%s>", topic)
		}
	}

	if app.rpc != nil {
		// RPC
		app.rpc.Startup(app.Logger())
		app.waiter.Add(1)
	}

	if app.http != nil {
		// HTTP
		app.http.loadRoutes()
		app.http.Startup(app.Logger())
		app.waiter.Add(1)
	}

	if app.metrics != nil {
		// Metrics node
		app.metrics.Startup(app.Logger())
		app.waiter.Add(1)
	}

	app.waiter.Wait()

	return
}

/* }}} */

// Shutdown : Close application
/* {{{ [AppIns::Shutdown] - Shutdown */
func (app *AppIns) Shutdown() {
	defer func() {
		recover()
	}()

	for i := 0; i < app.nWorkers; i++ {
		// Close worker
	}

	if app.metrics != nil {
		app.Logger().Debug("Metric node shutting down ...")
		app.metrics.Shutdown()
		app.Logger().Info("Metric node shutted")
		app.waiter.Done()
	}

	if app.http != nil {
		app.Logger().Debug("HTTP server shutting down ...")
		app.http.Shutdown()
		app.Logger().Info("HTTP server shutted")
		app.waiter.Done()
	}

	if app.rpc != nil {
		app.Logger().Debug("RPC server shutting down ...")
		app.rpc.Shutdown()
		app.Logger().Info("RPC server shutted")
		app.waiter.Done()
	}

	if app.nats != nil {
		app.Logger().Debug("NATS disconnecting ...")
		app.nats.Close()
		app.Logger().Info("NATS connection closed")
	}

	if app.nsq != nil {
		app.Logger().Debug("NSQ disconnecting ...")
		app.nsq.Shutdown()
		app.Logger().Info("NSQ connection closed")
	}

	if app.cronner != nil {
		app.cronner.Stop()
	}

	if app.db != nil {
		app.Logger().Debug("Database disconnecting ...")
		app.db.Close()
		app.Logger().Info("Database connection closed")
	}

	if app.redis != nil {
		app.Logger().Debug("Redis disconnecting ...")
		app.redis.Close()
		app.Logger().Info("Redis connection closed")
	}

	app.running = false

	return
}

/* }}} */

// LoadConfig : Load configuration
/* {{{ [AppIns::LoadConfig] */
func (app *AppIns) LoadConfig() {
	err := app.Config().ReadInConfig()
	if err != nil {
		app.Logger().Error(err)
	}

	return
}

/* }}} */

// ReloadConfig : Reload configuration
/* {{{ [AppIns::ReloadConfig] */
func (app *AppIns) ReloadConfig() {
	return
}

/* }}} */

// SetDefaultConfig : Set default configuration into default instance
/* {{{ [AppIns::SetDefaultConfig] */
func (app *AppIns) SetDefaultConfig(defaults map[string]interface{}) error {
	if defaults == nil {
		return fmt.Errorf("Null default configurations")
	}

	for k, v := range defaults {
		if v != nil {
			app.config.SetDefault(k, v)
		}
	}

	return nil
}

/* }}} */

// SetConfigs : Set configuration values, default values with same keys will be overridded
/* {{{ [AppIns::SetConfigs] */
func (app *AppIns) SetConfigs(c map[string]interface{}) error {
	if c == nil {
		return fmt.Errorf("Null configurations")
	}

	for k, v := range c {
		if v != nil {
			app.config.Set(k, v)
		}
	}

	return nil
}

/* }}} */

// SetHTTP : Set HTTP server instance
/* {{{ [AppIns::SetHTTP] */
func (app *AppIns) SetHTTP(srv *HTTPServer) {
	app.http = srv

	return
}

/* }}} */

// SetMetrics : Set metrics service instance
/* {{{ [AppIns::SetMetrics] */
func (app *AppIns) SetMetrics(metrics *MetricsIns) {
	app.metrics = metrics

	return
}

/* }}} */

// SetRPC : Set RPC server
/* {{{ [AppIns::SetRPC] */
func (app *AppIns) SetRPC(rpc *RPCServer) {
	app.rpc = rpc

	return
}

/* }}} */

// SetDB : Set database instance
/* {{{ [AppIns::SetDB] */
func (app *AppIns) SetDB(db db.Session) {
	app.db = db
	if _defaultDatabaseInstance == nil {
		_defaultDatabaseInstance = db
	}

	return
}

/* }}} */

// SetRedis : Set redis instance
/* {{{ [AppIns::SetRedis] */
func (app *AppIns) SetRedis(r *redis.Client) {
	app.redis = r
	if _defaultRedisInstance == nil {
		_defaultRedisInstance = r
	}

	return
}

/* }}} */

// SetNsq : Set Nsq client
/* {{{ [AppIns::SetNsq] */
func (app *AppIns) SetNsq(nsq *NsqClient) {
	app.nsq = nsq
	if _defaultNsqInstance == nil {
		_defaultNsqInstance = nsq
	}

	return
}

/* }}} */

// SetNats : Set NATS client
/* {{{ [AppIns::SetNats] */
func (app *AppIns) SetNats(nats *nats.Conn) {
	app.nats = nats
	if _defaultNatsInstance == nil {
		_defaultNatsInstance = nats
	}

	return
}

/* }}} */

// SetNWorker : Set number of workers
/* {{{ [AppIns::SetNWorker] */
func (app *AppIns) SetNWorker(n int) {
	app.nWorkers = n

	return
}

/* }}} */

/* {{{ [App::INSTANCES] */

// Cronner : Get cronner
func (app *AppIns) Cronner() *cron.Cron {
	return app.cronner
}

// Logger : Get logger
func (app *AppIns) Logger() *logrus.Entry {
	return app.logger.WithField(logFieldAppName, app.Name)
}

// Config : Get config
func (app *AppIns) Config() *viper.Viper {
	return app.config
}

// Metrics : Get metrics
func (app *AppIns) Metrics() *MetricsIns {
	return app.metrics
}

// DB : Get database
func (app *AppIns) DB() db.Session {
	return app.db
}

// Redis : Get redis
func (app *AppIns) Redis() *redis.Client {
	return app.redis
}

// Nsq : Get nsq
func (app *AppIns) Nsq() *NsqClient {
	return app.nsq
}

// Nats : Get nats
func (app *AppIns) Nats() *nats.Conn {
	return app.nats
}

// IsRunning : running status
func (app *AppIns) IsRunning() bool {
	return app.running
}

/* }}} */

/* {{{ [Instance] */

var (
	_defaultAppInstance      *AppIns
	_defaultCronnerInstance  *cron.Cron
	_defaultLoggerInstance   *logrus.Logger
	_defaultConfigInstance   *viper.Viper
	_defaultDatabaseInstance db.Session
	_defaultRedisInstance    *redis.Client
	_defaultNsqInstance      *NsqClient
	_defaultNatsInstance     *nats.Conn
)

// App : Get default app
func App() *AppIns {
	return _defaultAppInstance
}

// AppN : Get app by given name
func AppN(name string) *AppIns {
	return apps[name]
}

// Cronner : Get default cronner
func Cronner() *cron.Cron {
	return _defaultCronnerInstance
}

// Logger : Get default logger entry
func Logger() *logrus.Entry {
	return _defaultLoggerInstance.WithField(logFieldAppName, appName)
}

// Config : Get default viper config instance
func Config() *viper.Viper {
	return _defaultConfigInstance
}

// DB : Get default database instance
func DB() db.Session {
	return _defaultDatabaseInstance
}

// Redis : Get default redis instance
func Redis() *redis.Client {
	return _defaultRedisInstance
}

// Nsq : Get default nsq instance
func Nsq() *NsqClient {
	return _defaultNsqInstance
}

// Nats : Get default nats instance
func Nats() *nats.Conn {
	return _defaultNatsInstance
}

// Debug : Get debug status of default app instance
func Debug() bool {
	return _defaultAppInstance.Debug
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
