# System Architecture

## 1. Design Principles

### Core Philosophy
- Learning-focused microservices implementation
- Simplicity over complexity
- Clear, educational design patterns

### Architectural Goals
- Demonstrate microservices communication
- Implement event-driven workflows
- Showcase Go programming practices

## 2. Communication Strategies

### Hybrid Communication Model
- **Synchronous (gRPC)**
  * Critical, immediate interactions
  * Presigning upload URLs
  * Authentication checks

- **Asynchronous (Message Queues)**
  * Background processing
  * Event-driven workflows
  * Decoupled service interactions

## 3. Service Interactions

### FileUploadService
- Handles user authentication
- Manages file upload workflows
- Coordinates upload processes

### FileProcessorService
- Processes uploaded files
- Extracts file metadata
- Applies transformations

### FileStorageService
- Manages file storage
- Handles compression
- Provides file retrieval mechanisms

## 4. Data Flow

```
User Upload Request
↓
FileUploadService (gRPC)
├── Validate Request
├── Generate Upload Token
└── Initiate Storage Process
    ↓
    FileStorageService
    ├── Compress File
    ├── Store Metadata
    └── Publish Upload Event
        ↓
        Message Queue
        └── FileProcessorService
            ├── Process File
            └── Store Results
```

## 5. Key Design Decisions

### Communication
- gRPC for immediate interactions
- NATS for event-driven communication

### Storage
- Local filesystem
- SQLite for metadata
- Zstandard compression

### Processing
- Minimal, learning-focused transformations
- Support for text-based files

## 6. Performance Considerations

- Concurrent upload support
- Minimal processing overhead
- Efficient message routing

## 7. Scalability and Limitations

### Current Scope
- Single-node deployment
- Limited to text file processing
- Minimal error handling

### Future Potential
- Distributed architecture
- Advanced processing capabilities
- Enhanced error management

## 8. Technology Choices Rationale

### Go Language
- Strong typing
- Excellent concurrency support
- Compiled performance
- Simple syntax

### gRPC
- Efficient binary communication
- Strong typing
- Built-in streaming
- Language-agnostic

### Message Queues
- Decoupled service communication
- Scalable event handling
- Reliable message delivery

## 9. Security Considerations

- Basic token-based authentication
- Secure file upload mechanisms
- Minimal attack surface
- Learning-focused security model

## 10. Monitoring and Observability

- Basic logging
- Performance metrics
- Simple error tracking

## Communication Architecture

### Hybrid Communication Strategy

Our microservices architecture employs a hybrid communication approach:
- **Synchronous Communication**: gRPC
- **Asynchronous Communication**: NATS Message Queue

#### Communication Patterns

1. **Synchronous gRPC**
   - Critical, immediate interactions
   - Real-time request/response
   - Strong typing via Protocol Buffers
   - Used for:
     * Authentication
     * Upload URL generation
     * Metadata retrieval

2. **Asynchronous NATS**
   - Background processing
   - Decoupled service interactions
   - Event-driven workflows
   - Used for:
     * File processing triggers
     * Notification broadcasts
     * Asynchronous task management

### Communication Flow Diagram

```
Client Request
│
├── API Service (gRPC)
│   ├── Authentication
│   ├── Request Validation
│   └── Immediate Responses
│
└── Background Processing
    └── NATS Message Queue
        ├── File Processing Events
        ├── Service Notifications
        └── Asynchronous Workflows
```

### NATS Implementation Details

#### Key Features
- Lightweight message broker
- High-performance pub/sub system
- Simple, intuitive API
- Native Go support
- JetStream for persistent messaging

#### Message Queue Topology

1. **Topics**
   - `file.upload.initiated`
   - `file.process.request`
   - `file.process.completed`
   - `service.notification`

2. **Queue Groups**
   - Load balanced message consumption
   - Guaranteed single processing of messages

#### Example NATS Workflow

```go
// Publishing a file processing event
nc.Publish("file.process.request", &FileProcessEvent{
    FileID:   "unique-file-id",
    Metadata: fileMetadata,
})

// Subscribing to processing events
nc.Subscribe("file.process.request", func(msg *nats.Msg) {
    // Process file asynchronously
})
```

### Advantages of Hybrid Approach

1. **Flexibility**
   - Choose right communication method per use case
   - Optimize for performance and complexity

2. **Scalability**
   - Decoupled service interactions
   - Independent service scaling
   - Resilient to temporary service unavailability

3. **Performance**
   - Low-latency gRPC for critical paths
   - Efficient message routing with NATS
   - Minimal overhead

### Service Interaction Examples

1. **File Upload**
   - gRPC: Generate upload URL
   - NATS: Trigger background processing

2. **File Processing**
   - gRPC: Retrieve file metadata
   - NATS: Distribute processing tasks

### Monitoring and Observability

- Distributed tracing
- Message queue metrics
- Service health checks
- Performance logging

### Potential Future Enhancements

- Implement circuit breakers
- Add more sophisticated error handling
- Explore advanced NATS features (JetStream)
- Implement comprehensive logging

## Processing Service Architecture

### Service Responsibilities

#### Core Processing Capabilities
- Stream-based file processing
- Compressed file handling
- Metadata extraction
- Data transformation
- Flexible JSON processing

### File Processing Workflow

```
File Upload Complete
│
├── FileStorageService
│   ├── Compress File
│   └── Store Compressed File
│
└── FileProcessorService
    ├── Receive Processing Event
    ├── Retrieve Compressed File
    ├── Streaming Decompression
    ├── JSON Parsing
    │   ├── Chunk-based Processing
    │   ├── Memory Efficient Parsing
    │   └── Configurable Extractors
    │
    ├── Data Extraction
    │   ├── Apply Predefined Rules
    │   ├── Transform Data
    │   └── Generate Metadata
    │
    └── Result Handling
        ├── Store Processed Metadata
        └── Publish Completion Event
```

### Storage-Processor Service Interaction

#### Communication Mechanisms
- **Synchronous**: gRPC for metadata retrieval
- **Asynchronous**: NATS for processing events

#### Event-Driven Processing Flow

1. **File Upload Completion**
   - FileStorageService compresses file
   - Publishes file ready event to NATS

2. **Processing Initiation**
   ```go
   // NATS Event Structure
   type FileProcessEvent struct {
       FileID       string
       Filename     string
       StoragePath  string
       Compression  string  // zstd, lz4
       UploadedAt   time.Time
   }
   ```

3. **FileProcessorService Workflow**
   ```go
   func ProcessFile(event FileProcessEvent) {
       // 1. Retrieve compressed file
       compressedFile := FileStorageService.RetrieveFile(event.StoragePath)
       
       // 2. Create streaming processor
       processor := NewStreamingJSONProcessor(compressedFile)
       
       // 3. Process file
       result, err := processor.Process()
       
       // 4. Handle processing result
       if err == nil {
           // Store metadata
           // Publish success event
       } else {
           // Publish failure event
       }
   }
   ```

### Processing Strategies

#### JSON Processing Approach
- Streaming parser
- Chunk-based processing
- Configurable extractors
- Memory-efficient design

#### Key Processing Components
1. **Decompression Stream**
   - Support for zstd/lz4
   - Minimal memory overhead
   - Efficient for large files

2. **JSON Extractor**
   ```go
   type JSONExtractor struct {
       Name        string
       Path        string
       Transformer func(interface{}) (interface{}, error)
   }
   ```

3. **Transformation Rules**
   - Data normalization
   - Sensitive information removal
   - Custom transformation logic

### Metadata Generation

```go
type ProcessingMetadata struct {
    FileID             string
    TotalRecords       int
    ProcessingTime     time.Duration
    ExtractedFields    []string
    CompressionMethod  string
    ProcessingStatus   ProcessingStatus
}
```

### Error Handling and Resilience

- Graceful error management
- Partial processing support
- Detailed error logging
- Event-based error notification

### Monitoring and Observability

- Processing duration tracking
- Success/failure metrics
- Resource utilization monitoring
- Distributed tracing support

### Future Enhancements

1. Machine learning-based extraction
2. More complex transformation rules
3. Support for additional file types
4. Advanced error recovery mechanisms

### Technology Stack

- **Language**: Go
- **Parsing**: Streaming JSON parser
- **Compression**: zstd, lz4
- **Communication**: gRPC, NATS
- **Monitoring**: Prometheus, OpenTelemetry

## Architectural Approach: Modular Monolith

### Design Philosophy

#### Overview
Our project adopts a **Modular Monolith** architectural approach, balancing simplicity with future extensibility. This design provides a pragmatic solution for our learning-focused project, offering:
- Simplified initial development
- Reduced operational complexity
- Clear internal boundaries
- Preparation for potential future microservices

### Core Service Design

#### Responsibilities
The core service integrates multiple responsibilities:
- User management
- Authentication
- File upload processing
- Data storage handling
- Permission management

#### Architectural Structure

```go
type CoreService struct {
    UserManager     *UserManagement
    FileManager     *FileOperations
    StorageManager  *StorageHandler
    AuthManager     *AuthenticationProvider
}
```

### Design Principles

1. **Modularity**
   - Clear internal service boundaries
   - Use of interfaces and composition
   - Minimal coupling between components

2. **Flexibility**
   - Prepare for potential future service extraction
   - Implement dependency injection
   - Design with scalability in mind

### Architectural Benefits

#### Advantages
- Simplified initial implementation
- Lower operational complexity
- Faster inter-component communication
- Easier deployment and management

#### Potential Evolution
```
Current State: Single Service
[Core Service]
│
├── User Management
├── File Upload
├── Storage Handling
└── Authentication

Future Potential Microservices
[UserService]   [FileUploadService]   [FileStorageService]
```

### Scalability Considerations

#### When to Consider Microservices
- Increasing system complexity
- Divergent scaling requirements
- Performance bottlenecks
- Team growth and independent development needs

### Implementation Guidelines

1. Maintain clear internal module boundaries
2. Use interfaces for component interactions
3. Implement dependency injection
4. Design with potential future decomposition in mind

### Code Example

```go
// Unified interface for multiple operations
func (s *CoreService) ProcessUpload(
    ctx context.Context, 
    user *User, 
    uploadRequest *UploadRequest,
) (*UploadResult, error) {
    // Coordinate:
    // 1. User authentication
    // 2. Permission checking
    // 3. File processing
    // 4. Storage handling
    // 5. Metadata generation
}
```

### Conclusion

The Modular Monolith approach provides a balanced, pragmatic solution for our project. It offers the benefits of a monolithic architecture while maintaining the flexibility to evolve into microservices if required.

**Key Takeaway**: Prioritize clear design and modularity over premature architectural complexity.

## Frontend-Backend API Contract

### Overview
This section outlines the preliminary REST API design for frontend-backend communication.

### Design Principles
- RESTful API design
- OpenAPI/Swagger specification
- Clear, predictable endpoints
- Flexible metadata handling

### Key Endpoints
- File Upload Initiation
- File Upload Completion
- File Listing
- File Metadata Retrieval
- File Processing Trigger

### API Specification Location
- Detailed OpenAPI specification stored in `docs/openapi.yaml`
- Generated client SDKs will be available in future iterations

### Considerations
- Support for multiple frontend frameworks
- Easy integration with modern web technologies
- Minimal coupling between frontend and backend

### Future Refinements
- Comprehensive error handling
- Advanced filtering and pagination
- Potential GraphQL exploration
- Client SDK generation

### Technology Stack for API
- REST API
- OpenAPI 3.0
- JSON payload format
- HTTP/HTTPS communication

### Temporary Status
**Note**: This section is a placeholder and will be refined as the project evolves.

## Technology Stack

### Communication
- **Synchronous**: gRPC
- **Asynchronous**: NATS
- **Serialization**: Protocol Buffers

### Languages and Frameworks
- Go (Golang)
- gRPC-Go
- NATS Go Client
- OpenTelemetry (Tracing)

## Conclusion

A deliberately simple, educational microservices architecture focusing on core distributed systems concepts. Our hybrid communication architecture provides a robust, flexible, and educational approach to building microservices, balancing simplicity with powerful communication patterns.
