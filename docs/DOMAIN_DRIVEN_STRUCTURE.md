### upload-store-process/services/file-storage-service/
```
├── internal/
│   ├── domain/
│   │   ├── file/              # Core file domain models
│   │   │   ├── entity.go      # File entity definitions
│   │   │   └── errors.go      # Domain-specific errors
│   │   └── metadata/          # Metadata domain models
│   │       ├── entity.go
│   │       └── validation.go
│   │
│   ├── storage/
│   │   ├── service.go         # Storage service implementation
│   │   ├── repository.go      # Repository interface
│   │   └── providers/         # Storage implementations
│   │       ├── local/
│   │       └── s3/           # Future implementation
│   │
│   ├── upload/
│   │   ├── service.go        # Upload service implementation
│   │   ├── token/           # Token management
│   │   └── validation/      # Upload-specific validation
│   │
│   ├── metadata/
│   │   ├── service.go       # Metadata service implementation
│   │   └── repository/      # Metadata storage
│   │       └── sqlite/
│   │
│   └── transport/           # External-facing layers
│       ├── grpc/           # gRPC handlers
│       └── http/           # HTTP handlers
│
├── pkg/                    # Public packages
│   ├── storage/           # Storage interfaces
│   └── upload/            # Upload interfaces
```