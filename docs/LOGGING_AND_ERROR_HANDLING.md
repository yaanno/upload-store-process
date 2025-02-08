# Logging and Error Handling Strategy

## Overview
This document outlines the logging and error handling strategy for the Upload Store Process project, focusing on consistency, clarity, and debuggability.

## Logging Principles

### Log Levels
We use the following log levels:
- `ERROR`: Critical issues that prevent normal operation
- `WARN`: Potential issues or unexpected states
- `INFO`: Important operational events
- `DEBUG`: Detailed information for troubleshooting (used sparingly)

### Log Structure
Logs should be structured and include:
- Method/Operation name
- Relevant context (user ID, file ID, etc.)
- Error details
- Trace information

### Example Log Format
```go
logger.Error().
    Str("method", "PrepareUpload").
    Str("user_id", claims.UserID).
    Str("filename", req.Filename).
    Err(err).
    Msg("failed to prepare upload")
```

## Error Handling Strategy

### Error Types
1. **Authentication Errors**
   - Unauthorized access
   - Invalid tokens
   - Permission denied

2. **Validation Errors**
   - Missing required fields
   - Invalid input parameters
   - Unsupported file types

3. **Resource Errors**
   - File not found
   - Insufficient permissions
   - Resource constraints

4. **System Errors**
   - Database failures
   - External service unavailability
   - Unexpected runtime errors

### gRPC Error Mapping
Map errors to appropriate gRPC status codes:
- `codes.Unauthenticated`: Authentication failures
- `codes.InvalidArgument`: Validation errors
- `codes.PermissionDenied`: Authorization issues
- `codes.NotFound`: Resource not found
- `codes.Internal`: Unexpected system errors

### Error Handling Best Practices
1. Always log errors before returning
2. Include contextual information
3. Use meaningful error messages
4. Avoid exposing sensitive information
5. Implement proper error rollback mechanisms

## Code Example
```go
func (s *FileStorageServiceImpl) PrepareUpload(ctx context.Context, req *storagev1.PrepareUploadRequest) (*storagev1.PrepareUploadResponse, error) {
    // Validate JWT token
    claims, err := s.tokenValidator.ValidateToken(req.JwtToken)
    if err != nil {
        s.logger.Error().
            Str("method", "PrepareUpload").
            Str("token", req.JwtToken).
            Err(err).
            Msg("failed to validate JWT token")
        return nil, status.Errorf(codes.Unauthenticated, "invalid JWT token")
    }

    // Additional validation and error handling
    if claims.UserID == "" {
        s.logger.Error().
            Str("method", "PrepareUpload").
            Msg("user ID is empty")
        return nil, status.Errorf(codes.InvalidArgument, "user ID is required")
    }
}
```

## Monitoring and Tracing
- Implement distributed tracing
- Use correlation IDs for request tracking
- Integrate with monitoring systems

## Recommendations
- Centralize error handling logic
- Create custom error types when necessary
- Implement global error handler
- Regularly review and improve error messages

## Tools and Libraries
- Logging: `zerolog`
- Error Handling: `status` from `google.golang.org/grpc/status`
- Tracing: OpenTelemetry

## Continuous Improvement
- Conduct regular log and error handling reviews
- Analyze production logs for patterns
- Refactor and improve error handling based on real-world usage
