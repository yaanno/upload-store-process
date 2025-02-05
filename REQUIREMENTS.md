# System Requirements Specification

## 1. Project Overview
A microservices-based file upload, storage, and processing system designed as a learning project.

## 2. Functional Requirements

### 2.1 API Service
#### 2.1.1 File Upload
- SHALL accept UTF-8 encoded text-based file uploads via HTTP POST
  - Supported file types: CSV, JSON, TXT
- SHALL generate a unique upload identifier for each request
- SHALL validate incoming file before processing
  - File type MUST be one of: CSV, JSON, TXT
  - Maximum file size: 500 MB
  - Minimum file size: > 0 bytes
- SHALL reject invalid file uploads with descriptive error messages

### 2.2 Storage Service
#### 2.2.1 File Storage
- SHALL generate unique, secure file storage paths
- SHALL support local filesystem storage
- SHALL implement file naming convention:
  - Format: `{upload-id}_{timestamp}.{ext}` (ext: csv, json, or txt)
- SHALL store file metadata:
  - Original filename
  - Upload timestamp
  - File size
  - Upload source IP
  - File type

#### 2.2.2 File Validation
- SHALL validate text file encoding (UTF-8)
- SHALL reject files with unsupported encodings or file types

### 2.3 Processor Service
#### 2.3.1 Text Processing
- SHALL process uploaded UTF-8 text-based files
- SHALL extract file contents
- SHALL generate processing report for each uploaded file
- SHALL support text analysis specific to file type:
  - CSV: column count, header analysis
  - JSON: structure validation, key analysis
  - TXT: word count, language detection

#### 2.3.2 Processing Workflow
- SHALL handle files asynchronously via message queue
- SHALL support retry mechanism for failed processing
- SHALL log detailed processing steps and outcomes

## 3. Non-Functional Requirements

### 3.1 Performance
- SHALL support concurrent file uploads
- Recommended: Minimum 5 simultaneous uploads
- Maximum processing time per file: 5 minutes

### 3.2 Reliability
- SHALL implement comprehensive error handling
- SHALL provide detailed error logs
- SHALL support partial file processing recovery

### 3.3 Security
- SHALL sanitize all file paths
- SHALL prevent directory traversal attacks
- SHALL not store files outside designated storage directory

## 4. Technical Constraints
- Language: Go (Golang) 1.21+
- Containerization: Docker
- Inter-service Communication: gRPC, Message Queue
- No external cloud service dependencies

## 5. Future Considerations
- Support for additional archive formats
- Advanced file type processing
- Machine learning-based file analysis
- Distributed tracing
- Comprehensive monitoring dashboard
- Virus/malware scanning
- Advanced authentication
  - Implement basic authentication mechanism
  - Generate and validate upload tokens
  - Support rate limiting per user/IP
- Complex file transformations
- Advanced Observability
  - Implement structured logging
  - Support detailed log levels (DEBUG, INFO, WARN, ERROR)
  - Generate comprehensive processing metrics
    - Total uploads
    - Successful/failed uploads
    - Processing time distribution
  - Create monitoring dashboards
  - Implement distributed tracing

## 6. Out of Scope
- Long-term file storage
- Advanced authentication
- Complex file transformations

## 7. Acceptance Criteria
- All services MUST start successfully in Docker environment
- File upload workflow MUST complete end-to-end
- Processing reports MUST be generated for each upload
- System MUST handle various file scenarios

## 8. Compliance and Standards
- Follow Go best practices
- Adhere to OWASP security guidelines
- Implement graceful error handling
- Use semantic versioning

## 9. Testing Requirements
- Unit tests for each service
- Integration tests for service interactions
- Stress testing for concurrent uploads
- Security vulnerability scanning

## 10. Documentation
- Maintain up-to-date README
- Document API contracts
- Create inline code documentation
- Maintain CHANGELOG for significant changes
