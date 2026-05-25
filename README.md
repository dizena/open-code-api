# CLI Proxy API

基于 [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 二次开发的 OpenAI/Gemini/Claude 兼容 API 代理服务，提供完整的用户管理、API Key 管理、用量统计和计费功能。

## 项目简介

CLI Proxy API 是一个功能强大的 AI 模型代理服务，为 CLI 模型提供标准化的 API 接口，使其能够与为 OpenAI、Gemini、Claude 等 AI 服务设计的工具和库无缝集成。

### 主要特性
- **任意模型适配**: 支持自定义配置任意主流大模型，完美适配 Codex、Claude Code、Gemini CLI 等开发工具
- **多模型兼容**: 支持 OpenAI、Gemini、Claude 等主流 AI 模型的 API 接口
- **用户管理**: 完整的用户注册、登录、权限管理系统
- **API Key 管理**: 用户可创建和管理多个 API Key
- **用量统计**: 实时统计 API 调用次数和 Token 使用量
- **计费系统**: 支持按用量计费，管理员可配置模型倍率
- **管理后台**: 基于 Vue 3 的现代化管理界面
- **高可用架构**: 支持 MongoDB、Redis 等高性能存储
- **Docker 部署**: 提供完整的 Docker 和 Docker Compose 部署方案

## 技术架构

### 后端技术栈

- **语言**: Go 1.26
- **Web 框架**: Gin v1.10.1
- **数据库**: 
  - MongoDB (用户数据、配置数据)
  - Redis (缓存、队列)
- **认证**: JWT (golang-jwt/jwt/v5)
- **其他核心库**:
  - gorilla/websocket (WebSocket 支持)
  - minio-go/v7 (对象存储)
  - go-git/v6 (Git 操作)
  - logrus (日志记录)

### 前端技术栈

- **框架**: Vue 3.5.13
- **构建工具**: Vite 6.0.5
- **语言**: TypeScript 5.6.3
- **UI 组件库**: Element Plus 2.9.1
- **状态管理**: Pinia 2.3.0
- **路由**: Vue Router 4.5.0
- **HTTP 客户端**: Axios 1.7.9
- **样式**: TailwindCSS 3.4.17

### 项目结构

```
.
├── cmd/                    # 命令行工具和主程序入口
│   ├── server/            # 主服务器程序
│   └── fetch_antigravity_models/
├── internal/              # 内部包
│   ├── api/               # API 处理器和路由
│   ├── auth/              # 认证授权模块
│   ├── config/            # 配置管理
│   ├── interfaces/        # 接口定义
│   ├── logging/           # 日志系统
│   ├── registry/          # 模型注册和管理
│   └── ...
├── sdk/                   # SDK 包
├── ui/                    # 前端项目
│   ├── src/               # 源代码
│   ├── dist/              # 构建产物
│   └── package.json
├── config.yaml           # 配置文件
├── Dockerfile            # Docker 构建文件
├── docker-compose.yml    # Docker Compose 配置
└── go.mod                # Go 模块定义
```

## 快速开始

### 环境要求

- Go 1.26+
- Node.js 18+
- MongoDB 4.4+
- Redis 6.0+
- (可选) Docker & Docker Compose

### 安装步骤

#### 1. 克隆项目

```bash
git clone https://github.com/yourusername/open-code-api.git
cd open-code-api
```

#### 2. 配置环境

复制并修改配置文件：

```bash
cp config.example.yaml config.yaml
```

编辑 `config.yaml` 文件，配置数据库连接和其他参数：

```yaml
host: "0.0.0.0"
port: 8080

mongodb:
  enabled: true
  uri: "mongodb://user:password@ip:port/db"

usage-statistics-enabled: true

smtp:
  enabled: false
  host: smtp.exmail.qq.com
  port: 587
  user: your-email
  password: your-email-password

remote-management:
  disable-control-panel: true

debug: true
```

#### 3. 安装依赖

**后端依赖**:
```bash
go mod download
```

**前端依赖**:
```bash
cd ui
npm install
cd ..
```

#### 4. 启动服务

**开发模式启动后端**:
```bash
go run ./cmd/server
```

**开发模式启动前端**:
```bash
cd ui
npm run dev
```

**生产模式启动**:
```bash
# 构建前端
cd ui
npm run build
cd ..

# 启动后端（前端静态文件已集成）
go run ./cmd/server
```

## Docker 部署

### 使用 Docker Compose（推荐）

1. 确保已安装 Docker 和 Docker Compose
2. 创建必要的配置文件和目录
3. 启动服务：

```bash
docker-compose up -d
```

### 使用 Dockerfile

#### 构建镜像

```bash
docker build -t cli-proxy-api:latest .
```

#### 运行容器

```bash
docker run -d \
  --name cli-proxy-api \
  -p 8317:8317 \
  -p 8085:8085 \
  -p 1455:1455 \
  -p 54545:54545 \
  -p 51121:51121 \
  -p 11451:11451 \
  -v $(pwd)/config.yaml:/CLIProxyAPI/config.yaml \
  -v $(pwd)/auths:/root/.cli-proxy-api \
  -v $(pwd)/logs:/CLIProxyAPI/logs \
  cli-proxy-api:latest
```

## 打包构建

### 本地构建

#### Linux x64
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cli-proxy-api-linux-amd64 ./cmd/server
```

#### Linux ARM64
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o cli-proxy-api-linux-arm64 ./cmd/server
```

#### macOS
```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o cli-proxy-api-darwin-amd64 ./cmd/server
```

#### Windows
```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o cli-proxy-api-windows-amd64.exe ./cmd/server
```

### 使用 GoReleaser

项目已配置 GoReleaser，支持多平台自动打包：

```bash
# 安装 GoReleaser
go install github.com/goreleaser/goreleaser@latest

# 构建发布版本
goreleaser release --snapshot

# 构建快照版本
goreleaser build --snapshot
```

### 前端构建

```bash
cd ui
npm run build
```

构建产物将输出到 `ui/dist/` 目录。

## 配置说明

### 端口说明

服务默认占用以下端口：

- `8317`: 主 API 服务端口
- `8085`: 管理后台端口
- `1455`: WebSocket 端口
- `54545`: 代理服务端口
- `51121`: 额外服务端口
- `11451`: 监控端口

### 环境变量

支持通过环境变量覆盖配置：

- `DEPLOY`: 部署环境标识
- `CLI_PROXY_IMAGE`: Docker 镜像名称
- `CLI_PROXY_CONFIG_PATH`: 配置文件路径
- `CLI_PROXY_AUTH_PATH`: 认证文件路径
- `CLI_PROXY_LOG_PATH`: 日志文件路径

## API 使用

### 兼容接口

服务提供与 OpenAI API 兼容的接口，主要包括：

- `POST /v1/chat/completions` - 聊天完成
- `POST /v1/completions` - 文本完成
- `GET /v1/models` - 模型列表
- `POST /v1/embeddings` - 文本嵌入

### 认证方式

使用 API Key 进行认证：

```bash
curl -X POST http://localhost:8317/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## 管理后台

访问管理后台：`http://localhost:8085`

主要功能包括：

- 用户管理
- API Key 管理
- 模型配置
- 用量统计
- 计费设置
- 系统监控

## 开发指南

### 添加新的 API 接口

1. 在 `internal/api/handlers/` 中添加处理器
2. 在 `internal/api/server.go` 中注册路由
3. 更新相关文档

### 前端开发

```bash
cd ui
npm run dev    # 开发服务器
npm run build  # 生产构建
npm run preview  # 预览构建结果
```

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 前端遵循 ESLint 和 Prettier 规范

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查 MongoDB 和 Redis 服务是否正常运行
   - 验证配置文件中的连接字符串

2. **端口被占用**
   - 修改配置文件中的端口设置
   - 或停止占用端口的服务

3. **前端构建失败**
   - 确保 Node.js 版本符合要求
   - 删除 `node_modules` 和 `package-lock.json` 后重新安装

## 许可证

本项目基于原项目进行二次开发，请遵守相应的开源许可证。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进项目。

## 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 GitHub Issue
- 发送邮件至项目维护者

## 致谢

感谢 [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) 原项目提供的优秀基础架构。
感谢 [TopBeeAI](https://www.topbeeai.com) 技术团队提供技术支持。
