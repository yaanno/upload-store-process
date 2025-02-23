# Upload Store Process

A distributed file storage and processing system built with Go, focusing on efficient file handling, metadata management, and secure processing.

## Overview

This system provides a robust solution for:
- Secure file uploads and storage
- File metadata management with SQLite
- Deduplication and compression
- Distributed processing capabilities
- Real-time status tracking

## Technology Stack

- **Language**: Go (Golang) 1.21+
- **Communication**: gRPC, Message Queues
- **Database**: SQLite
- **Containerization**: Docker
- **File Processing**: Custom compression and deduplication
- **Security**: JWT-based authentication

## Features

- **File Storage Service**
  - Local filesystem storage
  - Metadata tracking
  - File deduplication
  - Compression support (zstd, LZ4)

- **Processing Service**
  - Asynchronous file processing
  - Status tracking
  - Error handling with retries
  - Configurable processing rules

- **API Layer**
  - gRPC interfaces
  - REST endpoints
  - Rate limiting
  - CORS support

### Project structure
```
.
├── docs/           # Comprehensive documentation
├── gen/           # Generated code (protobuf, etc.)
├── proto/         # Protocol buffer definitions
├── services/      # Core services implementation
└── tools/         # Development and maintenance tools
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker
- Make

### Running the Project

```bash
# Clone the repository
git clone https://github.com/yourusername/upload-store-process.git

# Navigate to project directory
cd upload-store-process

# Build and run services
make build
make run
```

## Learning Focus

- Microservices architecture
- Event-driven communication
- Basic distributed systems concepts
- File handling and processing

## Documentation

### Detailed Guides
- [Metadata Generation Strategy](/docs/METADATA_GENERATION.md): Learn about our approach to secure and flexible file metadata generation
- [Architecture Details](docs/ARCHITECTURE.md)
- [Detailed Requirements](docs/REQUIREMENTS.md)
- [Development Guide](docs/DEVELOPMENT.md)

## License

MIT License
