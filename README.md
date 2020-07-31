# Deuterium - Codebase skeleton of Fusion (Go version)

Go版本代码骨架

## 基于

* fasthttp
* fasthttprouter
* viper
* logrus
* upper_io

## 支持

* 自定义HTTP路由
* Prometheus自定义监控指标
* 自定义日志格式（文本/JSON）
* 环境变量 -> 配置文件 -> 默认值
* HTTP2 (h2c) + Msgpack RPC调用（支持snappy压缩）
* 基于NSQ的任务队列，自定义并行
* 基于Nats的消息通知
* 秒级计划任务
