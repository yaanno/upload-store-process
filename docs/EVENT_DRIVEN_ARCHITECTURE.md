# Event-Driven Architecture

## Event Types

### File Events
- FileUploaded
- FileProcessingStarted
- FileProcessingCompleted
- FileProcessingFailed
- FileDeleted

## Message Queue Structure
```yaml
queues:
  file_events:
    durable: true
    max_size: 1GB
    retention: 24h
  processing_events:
    durable: true
    max_size: 500MB
    retention: 12h
```

## Event Handlers
### Storage Service
- Handles: FileUploaded, FileDeleted
- Emits: FileMetadataUpdated
### Processing Service
- Handles: FileProcessingStarted
- Emits: FileProcessingCompleted, FileProcessingFailed
## Error Handling
- Dead Letter Queue
- Retry Policies
- Event Logging