
## serve

本项目使用 Go + cobra + logrus 实现启动一个静态服务 + 代理服务。


### 服务
- 服务支持 http 和 https
- 支持配置 ssl 证书 path
- 支持配置日志等级

### 静态服务
- 支持配置 static 路径


### 代理服务
代理机制：
实际请求的域名要作为 请求 的 第一个path item, 后续 item 依次排列，请求 Method 以及其他参数、请求头和请求体、响应参数、响应头、响应体不变。请求的 scheme 要根据配置进行转换。

比如：https://192.168.5.2:8080/www.qi.com/a/b/c?x=y
这个请求会请求到本服务，本服务根据第一个 path 判断是否命中代理配置，如果命中，则根据配置进行转换，配置中可以配置 use_https、insecure,其中 use_https 代表是否使用 https 协议，insecure 配置在使用 https 是生效，表示代理请求时是否不验证证书。
- 上述的链接在 use_https 为 false 时会转为 http://www.qi.com/a/b/c?x=y
- 上述的链接在 use_https 为 true 时会转为 https://www.qi.com/a/b/c?x=y， insecure 为 true 则在请求 https://www.qi.com/a/b/c?x=y 是跳过 ssl 验证

