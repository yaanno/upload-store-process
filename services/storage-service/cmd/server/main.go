package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/yaanno/upload-store-process/proto"
)

type storageServer struct {
	pb.UnimplementedStorageServiceServer
	uploadDir string
}

func (s *storageServer) PrepareUpload(ctx context.Context, req *pb.UploadPreparationRequest) (*pb.UploadPreparationResponse, error) {
	// Generate unique storage path
	uploadPath := filepath.Join(s.uploadDir, req.UploadId)

	// Ensure directory exists
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Generate a temporary upload token (in real-world, use more secure method)
	uploadToken := fmt.Sprintf("token-%s", req.UploadId)

	return &pb.UploadPreparationResponse{
		BasedResponse: &pb.Response{
			Success: true,
			Message: "Upload prepared successfully",
		},
		StoragePath: uploadPath,
		UploadToken: uploadToken,
	}, nil
}

func (s *storageServer) CompleteUpload(ctx context.Context, req *pb.UploadCompletionRequest) (*pb.Response, error) {
	// Validate upload
	uploadPath := filepath.Join(s.uploadDir, req.UploadId)

	// In a real implementation, you'd do more validation
	log.Printf("Upload completed for file: %s", req.FileMetadata.OriginalFilename)

	return &pb.Response{
		Success: true,
		Message: "Upload completed successfully",
	}, nil
}

func main() {
	// Configure upload directory
	uploadDir := "/tmp/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Create listener
	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register storage service
	storageService := &storageServer{
		uploadDir: uploadDir,
	}
	pb.RegisterStorageServiceServer(grpcServer, storageService)

	// Add reflection service for debugging
	reflection.Register(grpcServer)

	log.Printf("Storage service listening on port %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
