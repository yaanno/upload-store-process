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
1. Ensure Go 1.21+ is installed
2. Clone the repository
3. Run `go work sync`
4. Start services with Docker Compose

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
