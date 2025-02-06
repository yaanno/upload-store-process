# Upload Store Process

## Project Overview

### Purpose
A learning project to explore microservices architecture through a file storage and processing system.

### Core Objectives
- Implement a distributed file upload and processing service
- Learn microservices communication patterns
- Explore event-driven architectures
- Practice Go programming and system design

## System Components

### Services
1. **Backend Service**
   - Coordinate file upload workflow
   - Manage user interactions

2. **Storage Service**
   - Handle file storage
   - Manage file metadata
   - Implement compression

3. **Processor Service**
   - Process uploaded files
   - Extract metadata
   - Perform basic text transformations

## Technology Stack

- **Language**: Go (Golang) 1.21+
- **Communication**: gRPC, Message Queues
- **Database**: SQLite
- **Containerization**: Docker

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
