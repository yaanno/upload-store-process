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
		Filename:      "example.txt",
		FileSizeBytes: 1024,
		UserId:        "user_123",
	}

	// Call the RPC method
	response, err := client.PrepareUpload(ctx, request)
	if err != nil {
		log.Fatalf("Error preparing upload: %v", err)
	}

	// Process the response
	log.Printf("Upload token: %s", response)

	// Upload the file
	uploadRequest := &storagev1.UploadFileRequest{
		FileId:             response.GlobalUploadId,
		StorageUploadToken: response.StorageUploadToken,
		FileContent:        []byte("Hello, world!"),
		UserId:             "user_123",
	}

	uploadResponse, err := client.UploadFile(ctx, uploadRequest)
	if err != nil {
		log.Fatalf("Error uploading file: %v", err)
	}

	log.Printf("File uploaded: %v", uploadResponse)

	metadataRequest := &storagev1.GetFileMetadataRequest{
		FileId: uploadResponse.FileId,
		UserId: "user_123",
	}

	metadataResponse, err := client.GetFileMetadata(ctx, metadataRequest)
	if err != nil {
		log.Fatalf("Error getting file metadata: %v", err)
	}

	log.Printf("File metadata: %v", metadataResponse)
}
