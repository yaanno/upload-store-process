# API Documentation

## gRPC Services

### FileStorage Service
```protobuf
service FileStorage {
    rpc UploadFile(UploadRequest) returns (UploadResponse);
    rpc GetFileMetadata(MetadataRequest) returns (MetadataResponse);
    rpc ListFiles(ListRequest) returns (ListResponse);
    rpc DeleteFile(DeleteRequest) returns (DeleteResponse);
}

### Processing Service

```protobuf
service Processing {
    rpc ProcessFile(ProcessRequest) returns (ProcessResponse);
    rpc GetProcessingStatus(StatusRequest) returns (StatusResponse);
}
```

## REST Endpoints

### File Operations
- POST /api/v1/files - Upload file
- GET /api/v1/files/{id} - Get file metadata
- GET /api/v1/files - List files
- DELETE /api/v1/files/{id} - Delete file
### Processing Operations
- POST /api/v1/process/{id} - Start processing
- GET /api/v1/process/{id}/status - Get status