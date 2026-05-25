# CLI Proxy API

An OpenAI/Gemini/Claude compatible API proxy service based on [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI), providing complete user management, API key management, usage statistics, and billing functionality.

## Project Overview

CLI Proxy API is a powerful AI model proxy service that provides standardized API interfaces for CLI models, enabling seamless integration with tools and libraries designed for OpenAI, Gemini, Claude, and other AI services.

### Key Features

- **Flexible Model Adaptation**: Support for custom configuration of any mainstream large language models, perfectly adapted for Codex, Claude Code, Gemini CLI, and other development tools
- **Multi-Model Compatibility**: Supports API interfaces for mainstream AI models including OpenAI, Gemini, and Claude
- **User Management**: Complete user registration, login, and permission management system
- **API Key Management**: Users can create and manage multiple API keys
- **Usage Statistics**: Real-time statistics of API calls and token usage
- **Billing System**: Support for usage-based billing with configurable model-rates by administrators
- **Management Dashboard**: Modern management interface based on Vue 3
- **High-Availability Architecture**: Support for high-performance storage including MongoDB and Redis
- **Docker Deployment**: Complete Docker and Docker Compose deployment solutions

## Technical Architecture

### Backend Technology Stack

- **Language**: Go 1.26
- **Web Framework**: Gin v1.10.1
- **Databases**: 
  - MongoDB (user data, configuration data)
  - Redis (cache, queue)
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Other Core Libraries**:
  - gorilla/websocket (WebSocket support)
  - minio-go/v7 (object storage)
  - go-git/v6 (Git operations)
  - logrus (logging)

### Frontend Technology Stack

- **Framework**: Vue 3.5.13
- **Build Tool**: Vite 6.0.5
- **Language**: TypeScript 5.6.3
- **UI Component Library**: Element Plus 2.9.1
- **State Management**: Pinia 2.3.0
- **Router**: Vue Router 4.5.0
- **HTTP Client**: Axios 1.7.9
- **Styling**: TailwindCSS 3.4.17

### Project Structure

```
.
├── cmd/                    # Command-line tools and main program entry
│   ├── server/            # Main server program
│   └── fetch_antigravity_models/
├── internal/              # Internal packages
│   ├── api/               # API handlers and routes
│   ├── auth/              # Authentication and authorization module
│   ├── config/            # Configuration management
│   ├── interfaces/        # Interface definitions
│   ├── logging/           # Logging system
│   ├── registry/          # Model registration and management
│   └── ...
├── sdk/                   # SDK packages
├── ui/                    # Frontend project
│   ├── src/               # Source code
│   ├── dist/              # Build artifacts
│   └── package.json
├── config.yaml           # Configuration file
├── Dockerfile            # Docker build file
├── docker-compose.yml    # Docker Compose configuration
└── go.mod                # Go module definition
```

## Quick Start

### Prerequisites

- Go 1.26+
- Node.js 18+
- MongoDB 4.4+
- Redis 6.0+
- (Optional) Docker & Docker Compose

### Installation Steps

#### 1. Clone the Project

```bash
git clone https://github.com/yourusername/open-code-api.git
cd open-code-api
```

#### 2. Configure Environment

Copy and modify the configuration file:

```bash
cp config.example.yaml config.yaml
```

Edit the `config.yaml` file to configure database connections and other parameters:

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

#### 3. Install Dependencies

**Backend dependencies**:
```bash
go mod download
```

**Frontend dependencies**:
```bash
cd ui
npm install
cd ..
```

#### 4. Start the Service

**Development mode - backend**:
```bash
go run ./cmd/server
```

**Development mode - frontend**:
```bash
cd ui
npm run dev
```

**Production mode**:
```bash
# Build frontend
cd ui
npm run build
cd ..

# Start backend (frontend static files integrated)
go run ./cmd/server
```

## Docker Deployment

### Using Docker Compose (Recommended)

1. Ensure Docker and Docker Compose are installed
2. Create necessary configuration files and directories
3. Start the service:

```bash
docker-compose up -d
```

### Using Dockerfile

#### Build Image

```bash
docker build -t cli-proxy-api:latest .
```

#### Run Container

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

## Building and Packaging

### Local Build

#### Linux x64
```bash
CGCG_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cli-proxy-api-linux-amd64 ./cmd/server
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

### Using GoReleaser

The project is configured with GoReleaser for automated multi-platform packaging:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Build release version
goreleaser release --snapshot

# Build snapshot version
goreleaser build --snapshot
```

### Frontend Build

```bash
cd ui
npm run build
```

Build artifacts will be output to the `ui/dist/` directory.

## Configuration

### Port Configuration

The service uses the following ports by default:

- `8317`: Main API service port
- `8085`: Management dashboard port
- `1455`: WebSocket port
- `54545`: Proxy service port
- `51121`: Additional service port
- `11451`: Monitoring port

### Environment Variables

Configuration can be overridden via environment variables:

- `DEPLOY`: Deployment environment identifier
- `CLI_PROXY_IMAGE`: Docker image name
- `CLI_PROXY_CONFIG_PATH`: Configuration file path
- `CLI_PROXY_AUTH_PATH`: Authentication file path
- `CLI_PROXY_LOG_PATH`: Log file path

## API Usage

### Compatible Interfaces

The service provides OpenAI API-compatible interfaces, mainly including:

- `POST /v1/chat/completions` - Chat completions
- `POST /v1/completions` - Text completions
- `GET /v1/models` - Model list
- `POST /v1/embeddings` - Text embeddings

### Authentication

Use API key for authentication:

```bash
curl -X POST http://localhost:8317/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

## Management Dashboard

Access the management dashboard at: `http://localhost:8085`

Main features include:

- User management
- API key management
- Model configuration
- Usage statistics
- Billing settings
- System monitoring

## Development Guide

### Adding New API Endpoints

1. Add handlers in `internal/api/handlers/`
2. Register routes in `internal/api/server.go`
3. Update relevant documentation

### Frontend Development

```bash
cd ui
npm run dev       # Development server
npm run build     # Production build
npm run preview   # Preview build results
```

### Code Standards

- Follow Go official code standards
- Use `gofmt` for code formatting
- Frontend follows ESLint and Prettier standards

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check if MongoDB and Redis services are running properly
   - Verify connection strings in configuration file

2. **Port Already in Use**
   - Modify port settings in configuration file
   - Or stop the service occupying the port

3. **Frontend Build Failed**
   - Ensure Node.js version meets requirements
   - Delete `node_modules` and `package-lock.json` and reinstall

## License

This project is a secondary development based on the original project. Please comply with the corresponding open source license.

## Contributing

Issues and Pull Requests are welcome to improve the project.

## Contact

For questions or suggestions, please contact via:

- Submit GitHub Issues
- Email project maintainers

## Acknowledgments

Thanks to [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) for providing an excellent foundation architecture.
Thanks to the [TopBeeAI](https://www.topbeeai.com) technical team for providing technical support.