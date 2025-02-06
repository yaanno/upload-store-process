package repository

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
)

func TestSQLiteFileMetadataRepository_CreateFileMetadata(t *testing.T) {
	// Initialize test database
	migrator, err := InitializeTestDatabase()
	require.NoError(t, err)
	defer migrator.Close()

	// Create repository
	repo := NewSQLiteFileMetadataRepository(migrator, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// Prepare test metadata
	testMetadata := &sharedv1.FileMetadata{
		FileId:           "test-file-id-1",
		OriginalFilename: "test_file.txt",
		StoragePath:      "/tmp/test_file.txt",
		FileSizeBytes:    1024,
		FileType:         "text/plain",
		UploadTimestamp:  time.Now().Unix(),
		UserId:           "test-user-1",
	}

	// Test cases
	testCases := []struct {
		name        string
		metadata    *models.FileMetadataRecord
		expectError bool
	}{
		{
			name: "Valid Metadata Storage",
			metadata: &models.FileMetadataRecord{
				ID:               "test-file-id-1",
				Metadata:         testMetadata,
				StoragePath:      "/tmp/test_file.txt",
				ProcessingStatus: "PENDING",
				CreatedAt:        time.Now().UTC(),
				UpdatedAt:        time.Now().UTC(),
			},
			expectError: false,
		},
		{
			name:        "Nil Metadata",
			metadata:    nil,
			expectError: true,
		},
		{
			name: "Empty File ID",
			metadata: &models.FileMetadataRecord{
				ID:               "",
				Metadata:         testMetadata,
				ProcessingStatus: "PENDING",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.CreateFileMetadata(context.Background(), tc.metadata)
			
			if tc.expectError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Unexpected error")

				// Verify storage by retrieving
				storedFile, err := repo.RetrieveFileMetadataByID(context.Background(), tc.metadata.ID)
				assert.NoError(t, err, "Error finding stored file")
				assert.NotNil(t, storedFile, "Stored file should not be nil")
				assert.Equal(t, tc.metadata.ID, storedFile.ID, "Stored file ID should match")
				
				// Additional metadata verification
				if tc.metadata.Metadata != nil {
					assert.NotNil(t, storedFile.Metadata, "File metadata should not be nil")
					assert.Equal(t, tc.metadata.Metadata.OriginalFilename, 
						storedFile.Metadata.OriginalFilename, 
						"Original filename should match")
				}
			}
		})
	}
}

func TestSQLiteFileMetadataRepository_Upsert(t *testing.T) {
	// Initialize test database
	migrator, err := InitializeTestDatabase()
	require.NoError(t, err)
	defer migrator.Close()

	// Create repository
	repo := NewSQLiteFileMetadataRepository(migrator, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// Prepare initial metadata
	initialMetadata := &sharedv1.FileMetadata{
		FileId:           "test-upsert-id",
		OriginalFilename: "initial_file.txt",
		StoragePath:      "/tmp/initial_file.txt",
		FileSizeBytes:    1024,
		FileType:         "text/plain",
		UploadTimestamp:  time.Now().Unix(),
		UserId:           "test-user-1",
	}

	initialFileMetadata := &models.FileMetadataRecord{
		ID:               "test-upsert-id",
		Metadata:         initialMetadata,
		StoragePath:      "/tmp/initial_file.txt",
		ProcessingStatus: "PENDING",
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	// First store
	err = repo.CreateFileMetadata(context.Background(), initialFileMetadata)
	require.NoError(t, err, "First store should succeed")

	// Update metadata (create a new instance to avoid mutex copy)
	updatedMetadata := sharedv1.FileMetadata{
		FileId:           "test-upsert-id",
		OriginalFilename: "updated_file.txt",
		StoragePath:      "/tmp/updated_file.txt",
		FileSizeBytes:    initialMetadata.FileSizeBytes,
		FileType:         initialMetadata.FileType,
		UploadTimestamp:  initialMetadata.UploadTimestamp,
		UserId:           initialMetadata.UserId,
	}

	updatedFileMetadata := &models.FileMetadataRecord{
		ID:               "test-upsert-id",
		Metadata:         &updatedMetadata,
		StoragePath:      "/tmp/updated_file.txt",
		ProcessingStatus: "PROCESSING",
		CreatedAt:        initialFileMetadata.CreatedAt,
		UpdatedAt:        time.Now().UTC(),
	}

	// Update (upsert)
	err = repo.CreateFileMetadata(context.Background(), updatedFileMetadata)
	require.NoError(t, err, "Upsert should succeed")

	// Retrieve and verify
	storedFile, err := repo.RetrieveFileMetadataByID(context.Background(), "test-upsert-id")
	require.NoError(t, err, "Should find updated file")
	
	assert.Equal(t, "updated_file.txt", storedFile.Metadata.OriginalFilename, "Filename should be updated")
	assert.Equal(t, "/tmp/updated_file.txt", storedFile.StoragePath, "Storage path should be updated")
	assert.Equal(t, "PROCESSING", storedFile.ProcessingStatus, "Processing status should be updated")
}
