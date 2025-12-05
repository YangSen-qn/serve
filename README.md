# serve

一个基于 Go + cobra + logrus 实现的 HTTP/HTTPS 服务器，集成了静态文件服务和反向代理功能。
提供一种在手机端浏览器无法访问 http 服务的解决方案。

## 功能特性

### 服务配置

- 支持 HTTP 和 HTTPS 协议
- 支持配置 SSL 证书路径
- 支持配置日志等级（debug, info, warn, error）

### 静态文件服务

- 支持配置静态文件目录路径
- 提供静态资源访问能力
- 自动处理路径清理，防止路径遍历攻击

### 反向代理服务

代理服务采用路径前缀匹配的方式，将请求路径的第一段作为目标域名，后续路径段依次拼接。

#### 工作原理

1. 请求路径格式：`/{目标域名}/{路径段1}/{路径段2}/...?{查询参数}`
2. 目标域名作为路径的第一段（path item）
3. 后续路径段按顺序拼接
4. 请求方法（Method）、请求头、请求体、响应头、响应体等保持不变
5. 请求协议（scheme）根据配置进行转换

#### 配置项说明

- `use_https`：是否使用 HTTPS 协议进行代理请求
  - `false`：使用 HTTP 协议
  - `true`：使用 HTTPS 协议
- `insecure`：在使用 HTTPS 时是否跳过 SSL 证书验证（仅在 `use_https` 为 `true` 时生效）

#### 示例

**请求示例：**
```
https://192.168.5.2:8080/www.qi.com/a/b/c?x=y
```

**转换规则：**

1. 当 `use_https` 为 `false` 时：
   - 转换为：`http://www.qi.com/a/b/c?x=y`

2. 当 `use_https` 为 `true` 时：
   - 转换为：`https://www.qi.com/a/b/c?x=y`
   - 如果 `insecure` 为 `true`，则在请求时跳过 SSL 证书验证

**处理流程：**

1. 服务接收请求：`https://192.168.5.2:8080/www.qi.com/a/b/c?x=y`
2. 解析路径第一段：`www.qi.com`
3. 检查是否命中代理配置
4. 如果命中，根据配置项进行协议转换和请求转发
5. 将后续路径段 `/a/b/c` 和查询参数 `?x=y` 拼接至目标 URL

## 安装

### 方式一：使用 go build（推荐）

```bash
go build -o serve ./cmd/serve
```

### 方式二：使用 go install

```bash
go install ./cmd/serve@latest
```

使用 `go install` 安装后，可执行文件会被安装到 `$GOPATH/bin` 或 `$GOBIN` 目录下（如果已设置），确保该目录在系统的 PATH 环境变量中，即可直接使用 `serve` 命令。

### 方式三：从 Release 下载预编译二进制

访问 [GitHub Releases](https://github.com/your-username/serve/releases) 下载对应平台的预编译二进制文件。

### 本地构建多平台 Release

项目提供了构建脚本来生成所有平台的二进制文件：

```bash
# 构建所有平台的二进制文件
./scripts/build-release.sh v1.0.0

# 构建文件将输出到 dist/ 目录
# 每个压缩包都包含二进制文件和 README.md
```

构建脚本支持以下平台和架构：
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

**注意：** 每个平台的压缩包中都包含了 `README.md` 文件，方便用户查看使用说明。

## 使用方法

### 基本用法

```bash
# 启动 HTTP 服务器
./serve --host :8080 --static-dir ./static

# 启动 HTTPS 服务器
./serve --host :8443 --cert-file cert.pem --key-file key.pem --static-dir ./static

# 配置代理
./serve --host :8080 --static-dir ./static --proxy www.example.com:true:false

# 配置多个代理
./serve --host :8080 --static-dir ./static --proxy "www.example.com:true:false,api.example.com:false:false"
```

### 命令行参数

- `--host`: 监听地址（默认：`:8080`）
- `--cert-file`: SSL 证书文件路径（启用 HTTPS）
- `--key-file`: SSL 私钥文件路径（启用 HTTPS）
- `--log-level`: 日志等级，可选值：debug, info, warn, error（默认：`info`）
- `--static-dir`: 静态文件目录路径（默认：`./static`）
- `--proxy`: 代理配置，格式：`domain:use_https:insecure`，多个配置用逗号分隔

### 代理配置格式

代理配置格式：`domain:use_https:insecure`

- `domain`: 目标域名（如：`www.example.com`）
- `use_https`: 是否使用 HTTPS，`true` 或 `false`
- `insecure`: 是否跳过 SSL 验证，`true` 或 `false`（仅在 `use_https` 为 `true` 时生效）

示例：
- `www.example.com:true:false` - 使用 HTTPS，验证证书
- `api.example.com:true:true` - 使用 HTTPS，跳过证书验证
- `test.example.com:false:false` - 使用 HTTP

## 项目结构

```
serve/
├── cmd/
│   └── serve/
│       └── main.go          # 程序入口，命令行参数解析
├── internal/
│   ├── config/
│   │   └── config.go        # 配置管理模块
│   ├── server/
│   │   └── server.go         # HTTP/HTTPS 服务器实现
│   ├── static/
│   │   └── static.go         # 静态文件服务实现
│   └── proxy/
│       └── proxy.go          # 反向代理服务实现
├── scripts/
│   └── build-release.sh      # 多平台构建脚本
├── .github/
│   └── workflows/
│       └── release.yml       # GitHub Actions 自动发布工作流
├── go.mod                    # Go 模块定义
├── go.sum                    # 依赖校验和
├── .gitignore               # Git 忽略文件配置
└── README.md                 # 项目说明文档
```

## 开发说明

### 代码规范

- 所有代码注释使用中文
- 日志打点和 HTTP 请求返回的错误信息使用英文
- 保持代码结构清晰，模块职责单一

### 构建

```bash
# 构建可执行文件
go build -o serve ./cmd/serve

# 运行测试
go test ./...

# 格式化代码
go fmt ./...

# 代码检查
go vet ./...
```

## 许可证

MIT License

