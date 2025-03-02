# Development Guide

## Project Setup

### Prerequisites
- Go 1.21+
- Docker
- Task (Task runner)
- Protocol Buffers compiler

### Local Development Environment

1. **Clone Repository**
   ```bash
   git clone https://github.com/yourusername/upload-store-process.git
   cd upload-store-process
   ```

2. **Install Dependencies**
   ```bash
   task deps
   ```

3. **Generate Proto Files**
   ```bash
   task proto:generate
   ```

## Service Development Guide

### Project Services

1. **FileUploadService**
   - Located in: `services/file-upload-service`
   - Responsible for user authentication and upload coordination

2. **FileProcessorService**
   - Located in: `services/file-processor-service`
   - Handles file processing and metadata extraction

3. **FileStorageService**
   - Located in: `services/file-storage-service`
   - Manages file compression and storage

### Development Workflow

#### Service Setup
```bash
# Navigate to specific service directory
cd services/file-upload-service
# or
cd services/file-processor-service
# or
cd services/file-storage-service
```

#### Proto Definitions
- `proto/fileupload/v1/upload.proto`
- `proto/fileprocessor/v1/processor.proto`
- `proto/filestorage/v1/storage.proto`
- `proto/shared/v1/shared.proto`

### Local Development

#### Running Individual Services
```bash
# FileUploadService
task file-upload-service:dev

# FileProcessorService
task file-processor-service:dev

# FileStorageService
task file-storage-service:dev
```

## Running Services

### Development Mode
```bash
# Run all services
task services:up

# Run specific service
task service:backend
task service:storage
task service:processor
```

## Testing

### Unit Tests
```bash
# Run all unit tests
task test

# Test specific service
task test:backend
task test:storage
task test:processor
```

### Integration Tests
```bash
# Run integration tests
task test:integration
```

## Testing Strategy

#### Service-Specific Testing
- FileUploadService: Authentication, token generation
- FileProcessorService: File transformation tests
- FileStorageService: Compression, storage mechanisms

## Code Quality

### Linting
```bash
# Run golangci-lint
task lint
```

### Code Formatting
```bash
# Format all Go files
task fmt
```

## Debugging

### Logging
- Use structured logging
- Minimal, informative log levels
- Focus on key system events

### Debugging Tools
- Delve debugger
- Go's built-in race detector

### Debugging Guidelines

#### Debugging Services
```bash
# Enable verbose logging for specific service
LOG_LEVEL=debug task file-upload-service:dev
LOG_LEVEL=debug task file-processor-service:dev
LOG_LEVEL=debug task file-storage-service:dev
```

## Contribution Guidelines

### Branch Strategy
- `main`: Stable release
- `develop`: Active development
- Feature branches: `feature/description`

### Pull Request Process
1. Choose service directory
2. Implement changes
3. Write tests
4. Run service-specific tests
5. Submit pull request

### Code Review Checklist
- Clear, readable code
- Proper error handling
- Comprehensive tests
- Documentation updates
- Performance considerations

## Troubleshooting

### Common Issues
- Dependency conflicts
- Proto generation errors
- Docker configuration problems

### Getting Help
- Check documentation
- Review error logs
- Open GitHub issues

## Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof
go tool pprof mem.prof
```

## Best Practices

- Keep functions small and focused
- Use interfaces for flexibility
- Prioritize readability
- Write comprehensive tests
- Document non-obvious code

## Learning Resources

- Go official documentation
- gRPC guides
- Microservices design patterns
- Distributed systems concepts

## Taskfile Concepts

### Task Definition Principles
- Declarative task descriptions
- Cross-platform compatibility
- Minimal configuration overhead
- Composable and reusable tasks

### Taskfile Structure
```yaml
version: '3'

tasks:
  default:
    desc: "Default task"
    cmds:
      - echo "Welcome to Upload Store Process"

  init:
    desc: "Initialize project dependencies"
    cmds:
      - go mod download
      - go mod tidy

  proto:generate:
    desc: "Generate protobuf files"
    cmds:
      - protoc --go_out=. --go_opt=paths=source_relative ...
```

## Task-Driven Development
1. Create feature branch
2. Run `task init`
3. Implement changes
4. Run `task test`
5. Run `task lint`
6. Submit pull request

## Task Troubleshooting

### Common Task Issues
- Ensure Task is installed correctly
- Check Taskfile syntax
- Verify Go and protobuf installations

### Getting Task Help
```bash
# Task help
task --help

# Specific task help
task --help service:backend
```

## Task Best Practices

- Keep tasks small and focused
- Use task dependencies
- Leverage task variables
- Write self-documenting tasks
- Maintain cross-platform compatibility

## Task Learning Resources

- [Task Documentation](https://taskfile.dev/)
- Go module management
- Protobuf code generation
- Microservices design patterns
