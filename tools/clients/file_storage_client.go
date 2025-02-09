package main

import (
	"context"
	"log"
	"time"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Create a connection to the server
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a client
	client := storagev1.NewFileStorageServiceClient(conn)

	// Prepare a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a request
	request := &storagev1.PrepareUploadRequest{
		Filename:       "example.txt",
		FileSizeBytes:  1024,
		GlobalUploadId: "upload_123",
		UserId:         "user_123",
	}

	// Call the RPC method
	response, err := client.PrepareUpload(ctx, request)
	if err != nil {
		log.Fatalf("Error preparing upload: %v", err)
	}

	// Process the response
	log.Printf("Upload token: %s", response.StorageUploadToken)
}
