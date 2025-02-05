# Development Guide

## Project Setup

### Prerequisites
- Go 1.21+
- Docker
- Make
- Protocol Buffers compiler

### Local Development Environment

1. **Clone Repository**
   ```bash
   git clone https://github.com/yourusername/upload-store-process.git
   cd upload-store-process
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Generate Proto Files**
   ```bash
   make proto-gen
   ```

## Running Services

### Development Mode
```bash
# Build all services
make build

# Run services
make run

# Run specific service
make run-backend
make run-storage
make run-processor
```

## Testing

### Unit Tests
```bash
# Run all unit tests
make test

# Test specific service
make test-backend
make test-storage
make test-processor
```

### Integration Tests
```bash
# Run integration tests
make integration-test
```

## Code Quality

### Linting
```bash
# Run golangci-lint
make lint
```

### Code Formatting
```bash
# Format all Go files
make fmt
```

## Debugging

### Logging
- Use structured logging
- Minimal, informative log levels
- Focus on key system events

### Debugging Tools
- Delve debugger
- Go's built-in race detector

## Contribution Guidelines

### Branch Strategy
- `main`: Stable release
- `develop`: Active development
- Feature branches: `feature/description`

### Pull Request Process
1. Create feature branch
2. Write tests
3. Ensure all tests pass
4. Update documentation
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
