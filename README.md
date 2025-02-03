# Upload, Store, and Process Microservices Project

## Project Overview
A learning project demonstrating a microservices architecture for file upload, storage, and processing.

## Architecture

### Services
1. **API Service**
   - Exposes HTTP endpoints
   - Validates incoming requests
   - Coordinates upload process
   - Manages authentication

2. **Storage Service**
   - Manages file storage
   - Generates secure upload locations
   - Handles file persistence
   - Triggers file processing

3. **Processor Service**
   - Processes uploaded files
   - Receives tasks via message queue
   - Handles file transformations

### Communication Patterns
- API ↔ Storage Service: gRPC
- Storage Service → Processor Service: Message Queue (RabbitMQ/NATS)

## Technology Stack
- Language: Go (Golang) 1.21
- Containerization: Docker
- Inter-Service Communication: 
  - gRPC
  - Message Queue
- Monorepo Structure

## Project Structure
```
upload-store-process/
│
├── go.work                 # Go workspace configuration
│
├── services/               # Service implementations
│   ├── api-service/
│   ├── storage-service/
│   └── processor-service/
│
├── proto/                  # gRPC service definitions
│   ├── common.proto
│   ├── api_service.proto
│   ├── storage_service.proto
│   └── processor_service.proto
│
└── shared/                 # Potential shared packages
```

## API Contracts
- Defined using Protocol Buffers (protobuf)
- gRPC-based inter-service communication
- Contracts located in `proto/` directory
  - `common.proto`: Shared types
  - `api_service.proto`: API service endpoints
  - `storage_service.proto`: Storage service methods
  - `processor_service.proto`: File processing service

## Development Roadmap
- [x] Project structure setup
- [ ] Define gRPC contracts
- [ ] Implement basic service skeletons
- [ ] Set up message queue
- [ ] Implement file upload logic
- [ ] Add authentication
- [ ] Create Docker configurations
- [ ] Implement error handling
- [ ] Add logging and monitoring

## Local Development
1. Ensure Go 1.22+ is installed
2. Clone the repository
3. Install protobuf tools:
   ```bash
   brew install protobuf
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```
4. Generate proto code:
   ```bash
   protoc --go_out=proto/gen --go_opt=paths=source_relative \
          --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
          proto/*.proto
   ```

## Docker Deployment
1. Build and run services:
   ```bash
   docker-compose up --build
   ```

### Service Status
- [x] Storage Service: Basic gRPC implementation
- [ ] API Service: Pending implementation
- [ ] Processor Service: Pending implementation

## Learning Objectives
- Microservices architecture
- Go programming
- gRPC and message queue communication
- Containerization
- Distributed system design

## Future Improvements
- Add more robust error handling
- Implement comprehensive logging
- Create monitoring dashboards
- Add more advanced processing capabilities

## Contributing
Contributions, issues, and feature requests are welcome!

## License
[To be determined]
