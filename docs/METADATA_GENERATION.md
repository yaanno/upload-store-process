# File Metadata Generation Strategy

## Overview

Our file metadata generation follows a hybrid approach that leverages the strengths of both the Upload and Storage services while maintaining flexibility and security.

## Design Principles

1. **Collaborative Metadata Generation**
   - Upload service prepares initial metadata
   - Storage service validates and completes metadata
   - Consistent ID generation across services

2. **Metadata Lifecycle**

### Preliminary Metadata (Upload Service)
- Generate preliminary upload ID
- Capture initial file information
  - Original filename
  - File size
  - Initial upload timestamp
  - User ID (if available)

### Finalization (Storage Service)
- Validate or generate secure file ID
- Determine final storage path
- Ensure all metadata fields are complete
- Persist metadata in storage system

## ID Generation Strategy

### Requirements
- Cryptographically secure
- Globally unique
- URL-safe
- Unpredictable

### Generation Process
1. Use cryptographically secure random number generator
2. Generate 32 bytes of random data
3. Encode using base64 URL-safe encoding

### Example Generation Pseudocode
```go
func generateSecureFileID() (string, error) {
    randomBytes := make([]byte, 32)
    _, err := rand.Read(randomBytes)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(randomBytes), nil
}
```

## Potential Future Enhancements
- Distributed ID generation (e.g., Snowflake algorithm)
- Enhanced collision detection
- Configurable ID generation strategies

## Security Considerations
- Never expose raw ID generation mechanism
- Use cryptographically secure random sources
- Implement rate limiting on ID generation
- Validate IDs before storage

## Trade-offs
### Pros
- Flexible metadata completion
- Clear service responsibilities
- Secure ID generation
- Minimal proto contract changes

### Cons
- Slightly more complex implementation
- Potential for minor performance overhead

## Recommended Implementation Pattern

```go
// Upload Service
func PrepareUpload(req *PrepareUploadRequest) {
    // Generate preliminary metadata
    // Partial information only
}

// Storage Service
func CompleteUpload(req *CompleteUploadRequest) {
    // Validate and complete metadata
    // Generate/validate file ID
    // Finalize storage path
}
```

## Monitoring and Observability
- Log metadata generation events
- Track ID generation success/failure rates
- Monitor metadata completion latency

## Open Questions and Future Work
- Investigate distributed ID generation strategies
- Explore potential performance optimizations
- Develop comprehensive testing for ID generation
