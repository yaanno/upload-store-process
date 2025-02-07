# Shared Authentication Package

## Features
- JWT Token Generation
- Token Validation
- gRPC Interceptor
- Claims Management

## Usage Example

### Token Generation
```go
generator := auth.NewTokenGenerator("secret_key", "myservice")
token, err := generator.GenerateToken(
    userID, 
    email, 
    []string{"user"}, 
    []string{"read:files", "write:files"}, 
    24*time.Hour
)
```

### gRPC Server Setup
```go
tokenGenerator := auth.NewTokenGenerator(secretKey, "myservice")
server := grpc.NewServer(
    grpc.UnaryInterceptor(
        auth.UnaryServerInterceptor(
            tokenGenerator, 
            // Optional: skip authentication for certain methods
            "/myservice.PublicMethod"
        )
    )
)
```

### Method-Level Authorization
```go
func (s *Service) SomeMethod(ctx context.Context, req *Request) (*Response, error) {
    claims, ok := auth.GetClaimsFromContext(ctx)
    if !ok {
        return nil, status.Errorf(codes.Unauthenticated, "no claims")
    }

    if !claims.HasPermission("write:files") {
        return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
    }

    // Method logic
}
```

## Security Considerations
- Use strong, unique secret keys
- Set appropriate token expiration
- Implement token rotation
- Use HTTPS for all communications
