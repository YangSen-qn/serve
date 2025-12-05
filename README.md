# serve

一个基于 Go + cobra + logrus 实现的 HTTP/HTTPS 服务器，集成了静态文件服务和反向代理功能。
提供一种在手机端浏览器无法访问 http 服务的解决方案。

## 功能特性

### 服务配置

- 支持 HTTP 和 HTTPS 协议
- 支持配置 SSL 证书路径
- 支持配置日志等级（debug, info, warn, error）
- **兼容性支持**：支持 Android 4 等旧版本浏览器（TLS 1.0+，兼容加密套件）

### 静态文件服务

- 支持配置静态文件目录路径
- 提供静态资源访问能力
- 自动处理路径清理，防止路径遍历攻击

### 反向代理服务

代理服务采用路径前缀匹配的方式：
- 请求路径的第一段作为路径前缀，用于匹配代理配置
- 如果找到匹配的配置：
  - 如果配置中指定了目标域名，使用配置的目标域名
  - 如果配置中未指定目标域名，使用路径第一段作为目标域名
- 如果未找到匹配的配置，则不会进行代理转发

#### 工作原理

1. **请求路径格式**：`/{路径前缀}/{路径段1}/{路径段2}/...?{查询参数}`
2. **域名确定规则**：
   - 解析请求路径的第一段作为路径前缀
   - 查找是否存在匹配该前缀的代理配置
   - 如果找到配置：
     - 如果配置中指定了 `target_domain`，使用配置的目标域名
     - 如果配置中未指定 `target_domain`（为空），使用路径第一段作为目标域名
3. **路径处理**：
   - 如果配置中指定了 `target_domain`，保留路径前缀，完整路径转发
   - 如果配置中未指定 `target_domain`（为空），移除路径第一段（路径前缀），后续路径段按顺序拼接（第一段做了域名）
   - 保留查询参数
4. **请求方法（Method）、请求头、请求体、响应头、响应体等保持不变**
5. **请求协议（scheme）根据配置进行转换**

#### 配置项说明

代理配置格式：`path_prefix:target_domain:use_https:insecure`

- `path_prefix`：路径前缀，用于匹配请求路径第一段
- `target_domain`：目标域名
  - 如果指定了目标域名，则使用该域名进行代理转发
  - 如果为空（两个连续冒号之间为空），则使用 `path_prefix` 作为目标域名
- `use_https`：是否使用 HTTPS 协议（`true` 或 `false`）
- `insecure`：是否跳过 SSL 证书验证（`true` 或 `false`，仅在 `use_https` 为 `true` 时生效）

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

**场景一：使用配置中指定的目标域名**
1. 配置：`api:api.example.com:true:false`
2. 服务接收请求：`https://192.168.5.2:8080/api/a/b/c?x=y`
3. 解析路径第一段：`api`
4. 检查是否存在路径前缀 `api` 的代理配置
5. 如果存在配置，使用配置的目标域名 `api.example.com`
6. 保留路径前缀（因为配置了目标域名），将完整路径 `/api/a/b/c` 和查询参数 `?x=y` 拼接至目标 URL：`https://api.example.com/api/a/b/c?x=y`

**场景二：使用路径第一段作为目标域名**
1. 配置：`api.com::true:false`（target_domain 为空）
2. 服务接收请求：`https://192.168.5.2:8080/api.com/a/b/c?x=y`
3. 解析路径第一段：`api.com`
4. 检查是否存在路径前缀 `api.com` 的代理配置
5. 如果存在配置，由于 target_domain 为空，使用路径第一段 `api.com` 作为目标域名
6. 移除路径前缀，将后续路径段 `/a/b/c` 和查询参数 `?x=y` 拼接至目标 URL：`https://api.com/a/b/c?x=y`

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
# 构建所有平台的二进制文件（版本号会自动注入到二进制文件中）
./scripts/build-release.sh v1.0.0

# 构建文件将输出到 .build/ 目录
# 每个压缩包都包含二进制文件和 README.md
```

构建脚本支持以下平台和架构：
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

**注意：**
- 每个平台的压缩包中都包含了 `README.md` 文件，方便用户查看使用说明
- 构建时指定的版本号会通过 `-ldflags` 注入到二进制文件中，可通过 `./serve -v` 查看
- GitHub Actions 工作流会在创建 Release 时自动触发，从 Release 的 tag 中提取版本号并注入到所有平台的二进制文件中

## 使用方法

### 基本用法

```bash
# 启动 HTTP 服务器
./serve --host :8080 --static-dir ./static

# 启动 HTTPS 服务器
./serve --host :8443 --cert-file cert.pem --key-file key.pem --static-dir ./static

# 配置代理（指定目标域名）
./serve --host :8080 --static-dir ./static --proxy api:api.example.com:true:false

# 配置代理（使用路径第一段作为目标域名，target_domain 为空）
./serve --host :8080 --static-dir ./static --proxy api::true:false

# 配置多个代理（使用多个 --proxy 参数）
./serve --host :8080 --static-dir ./static --proxy api:api.example.com:true:false --proxy www:www.example.com:false:false
```

### 命令行参数

- `-v, --version`: 显示版本信息并退出
- `--host`: 监听地址（默认：`:8080`）
- `--cert-file`: SSL 证书文件路径（启用 HTTPS）
- `--key-file`: SSL 私钥文件路径（启用 HTTPS）
- `--log-level`: 日志等级，可选值：debug, info, warn, error（默认：`info`）
- `--static-dir`: 静态文件目录路径（默认：`./static`）
- `--proxy`: 代理配置，格式：`path_prefix:target_domain:use_https:insecure`
  - `path_prefix`: 路径前缀，用于匹配请求路径第一段
  - `target_domain`: 目标域名，如果为空则使用 `path_prefix` 作为目标域名
  - `use_https`: 是否使用 HTTPS，`true` 或 `false`
  - `insecure`: 是否跳过 SSL 验证，`true` 或 `false`（仅在 `use_https` 为 `true` 时生效）
  - 可以多次使用 `--proxy` 参数来配置多个代理

### 查看版本

```bash
# 显示版本信息
./serve -v
# 或
./serve --version
```

### 代理配置格式

代理配置格式：`path_prefix:target_domain:use_https:insecure`

- `path_prefix`: 路径前缀，用于匹配请求路径第一段（如：`api`）
- `target_domain`: 目标域名（如：`api.example.com`）
  - 如果指定了目标域名，则使用该域名进行代理转发
  - 如果为空（两个连续冒号之间为空），则使用 `path_prefix` 作为目标域名
- `use_https`: 是否使用 HTTPS，`true` 或 `false`
- `insecure`: 是否跳过 SSL 验证，`true` 或 `false`（仅在 `use_https` 为 `true` 时生效）

**工作流程：**
1. 解析请求路径第一段作为路径前缀
2. 检查是否存在匹配的代理配置（路径第一段等于配置的 `path_prefix`）
3. 如果匹配成功：
   - 如果配置中指定了 `target_domain`，使用配置的目标域名，并保留路径前缀（完整路径转发）
   - 如果配置中未指定 `target_domain`（为空），使用路径第一段作为目标域名，并移除路径前缀
   - 根据配置的协议（HTTP/HTTPS）进行转发
4. 如果未匹配，则不会进行代理转发（可能由静态文件服务处理）

示例：
- `api:api.example.com:true:false` - 匹配路径 `/api/...`，代理到 `https://api.example.com/api/...`（保留路径前缀），验证证书
- `api.com::true:false` - 匹配路径 `/api.com/...`，代理到 `https://api.com/...`（移除路径前缀），验证证书（target_domain 为空，使用路径第一段作为目标域名）
- `abc:api.example.com:true:true` - 匹配路径 `/abc/...`，代理到 `https://api.example.com/abc/...`（保留路径前缀），跳过证书验证

**配置多个代理：**
```bash
# 使用多个 --proxy 参数
./serve --host :8080 --static-dir ./static \
  --proxy api:api.example.com:true:false \
  --proxy www:www.example.com:false:false \
  --proxy test::true:true
```

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

