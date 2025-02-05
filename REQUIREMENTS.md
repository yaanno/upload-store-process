# System Requirements Specification

## 1. Project Overview
A microservices-based file upload, storage, and processing system designed as a learning project, focusing on a local, self-contained file management ecosystem.

## 2. System Architecture

### 2.1 Architectural Principles
- Microservices-based design
- Decoupled service communication
- Direct file upload mechanism
- Asynchronous processing
- Local storage implementation

### 2.2 Core Services
1. **Frontend Service**
   - User interface for file uploads
   - Handles direct file uploads
   - Receives processing status updates

2. **Backend (API) Service**
   - Coordination point between services
   - Manages upload workflow
   - Handles service communication
   - Provides status tracking

3. **Storage Service**
   - Local, S3-like file storage
   - Generates unique file paths
   - Manages file metadata
   - Supports secure file uploads

4. **Processor Service**
   - Handles file processing
   - Generates processing reports
   - Provides status updates

## 3. Storage Service Specifications

### 3.1 Minimum Viable Product (MVP) Features

#### 3.1.1 Core Functional Requirements
1. **Unique File Identification**
   - Generate cryptographically secure unique identifiers
   - Prevent identifier guessing or enumeration
   - Ensure global uniqueness across uploads

2. **Secure Upload Mechanism**
   - Implement token-based upload authentication
   - Generate time-limited access tokens
   - Validate file before storage
     * Check file size
     * Verify file type
     * Sanitize file paths

3. **Basic Metadata Tracking**
   - Store essential file metadata in SQLite
   - Capture:
     * Unique file identifier
     * Original filename
     * File size
     * Upload timestamp
     * Storage path

4. **Simple File Retrieval**
   - Retrieve files by unique identifiers
   - Support basic file listing
   - Provide simple pagination for file lists

5. **Minimal Access Controls**
   - Implement basic token-based access
   - Support upload and download access
   - Configure token expiration

### 3.2 Technical Implementation Details

#### 3.2.1 Storage Backend
- **Primary Storage**: Local filesystem
- **Metadata Storage**: SQLite database
- **Directory Structure**:
  ```
  /data/uploads/
  └── {year}/
      └── {month}/
          └── {day}/
              └── {unique_identifier}_{original_filename}
  ```

#### 3.2.2 Implementation Constraints
- Minimal error handling
- Simple, straightforward implementation
- Focus on core functionality

### 3.3 Potential Future Enhancements

#### 3.3.1 File Management Improvements
1. **Compression Support**
   - Add optional file compression
   - Support for compression/decompression
   - Configurable compression levels

2. **Security Enhancements**
   - Virus/Malware Scanning
     * Integrate with external scanning tools
     * Optional virus check during upload
   - Advanced access controls
     * Role-based access
     * Granular permissions

3. **Metadata and Processing
   - Thumbnail Generation
     * Support for image file thumbnails
     * Configurable thumbnail sizes
   - Advanced Metadata Extraction
     * File type-specific metadata parsing
     * Deeper file content analysis

4. **Lifecycle Management**
   - File Expiration Policies
     * Automatic file deletion after set period
     * Configurable retention rules
   - Versioning Support
     * Track file versions
     * Restore previous file versions

### 3.4 Non-Functional Requirements

#### 3.4.1 Performance Considerations
- Efficient metadata indexing
- Minimal I/O overhead
- Support concurrent file operations
- Quick metadata retrieval

#### 3.4.2 Scalability
- Support for increasing file count
- Efficient storage path generation
- Minimal performance degradation

### 3.5 Limitations and Constraints
- Maximum file size: 500 MB
- Supported file types: CSV, JSON, TXT
- UTF-8 encoded text files only
- Local storage only

### 3.6 Open Research Areas
- Performance benchmarking
- Concurrent upload stress testing
- Metadata query optimization
- Token generation security

## 4. File Upload Workflow

### 4.1 Upload Initiation
- User selects file in Frontend
- Frontend sends initial upload request to Backend
- Request includes:
  * Filename
  * File size

### 4.2 Backend Coordination
- Validates basic file parameters
- Generates unique upload identifier
- Communicates with Storage Service
- Obtains upload token/presigned URL
- Returns upload details to Frontend

### 4.3 Direct Upload Mechanism
- Frontend uploads file directly to Storage Service
- Uses provided upload token/URL
- Storage Service:
  * Receives file
  * Generates file metadata
  * Stores file locally
  * Signals Backend about successful upload

### 4.4 Processing Workflow
- Backend notifies Processor Service
- Passes file metadata
- Processor Service:
  * Begins asynchronous file processing
  * Generates periodic status updates
  * Signals processing completion
- Backend relays status updates

## 5. Supported File Types
- CSV
- JSON
- TXT
- UTF-8 encoded text files
- Maximum file size: 500 MB

## 6. Non-Functional Requirements

### 6.1 Performance
- Support concurrent file uploads
- Minimum 5 simultaneous uploads
- Maximum processing time per file: 5 minutes

### 6.2 Storage
- Local filesystem-based storage
- Hierarchical file organization
- Metadata tracking
- Configurable storage root directory

### 6.3 Security
- Upload tokens
- File size validation
- Basic access controls

## 7. Future Considerations
- WebSocket for real-time updates
- Enhanced error handling
- Distributed storage support
- Advanced processing capabilities

## 8. Open Questions
- Retry mechanism for failed uploads
- Granularity of processing status
- Authentication mechanisms

## 9. Technical Constraints
- Language: Go (Golang) 1.21+
- Containerization: Docker
- Inter-service Communication: gRPC, Message Queue
- No external cloud service dependencies

## 10. Testing Requirements
- Unit tests for each service
- Integration tests for service interactions
- Stress testing for concurrent uploads
- Security vulnerability scanning

## 11. Documentation
- Maintain up-to-date README
- Document API contracts
- Create inline code documentation
- Maintain CHANGELOG for significant changes
