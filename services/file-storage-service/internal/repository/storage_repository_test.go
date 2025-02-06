package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
)

func TestSQLiteStorageRepository_Store(t *testing.T) {
	// Initialize test database
	migrator, err := InitializeTestDatabase()
	require.NoError(t, err)
	defer migrator.Close()

	// Create repository
	repo := NewSQLiteStorageRepository(migrator)

	// Prepare test metadata
	testMetadata := &sharedv1.FileMetadata{
		FileId:            "test-file-id-1",
		OriginalFilename:  "test_file.txt",
		StoragePath:       "/tmp/test_file.txt",
		FileSizeBytes:     1024,
		FileType:          "text/plain",
		UploadTimestamp:   time.Now().Unix(),
		UserId:            "test-user-1",
	}

	// Test cases
	testCases := []struct {
		name        string
		storage     *models.Storage
		expectError bool
	}{
		{
			name: "Valid Storage",
			storage: &models.Storage{
				ID:             "test-file-id-1",
				FileMetadata:   testMetadata,
				StoragePath:    "/tmp/test_file.txt",
				ProcessingStatus: "PENDING",
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			},
			expectError: false,
		},
		{
			name: "Nil Storage",
			storage: nil,
			expectError: true,
		},
		{
			name: "Empty ID",
			storage: &models.Storage{
				ID:             "",
				FileMetadata:   testMetadata,
				ProcessingStatus: "PENDING",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			err := repo.Store(ctx, tc.storage)

			if tc.expectError {
				assert.Error(t, err, "Expected an error for test case: %s", tc.name)
			} else {
				assert.NoError(t, err, "Unexpected error for test case: %s", tc.name)

				// Verify storage by retrieving
				if tc.storage != nil && tc.storage.ID != "" {
					storedFile, err := repo.FindByID(ctx, tc.storage.ID)
					assert.NoError(t, err, "Error finding stored file")
					assert.NotNil(t, storedFile, "Stored file should not be nil")
					assert.Equal(t, tc.storage.ID, storedFile.ID, "Stored file ID should match")
					
					// Additional metadata verification
					if tc.storage.FileMetadata != nil {
						assert.NotNil(t, storedFile.FileMetadata, "File metadata should not be nil")
						assert.Equal(t, tc.storage.FileMetadata.OriginalFilename, 
							storedFile.FileMetadata.OriginalFilename, 
							"Original filename should match")
					}
				}
			}
		})
	}
}

func TestSQLiteStorageRepository_Upsert(t *testing.T) {
	// Initialize test database
	migrator, err := InitializeTestDatabase()
	require.NoError(t, err)
	defer migrator.Close()

	// Create repository
	repo := NewSQLiteStorageRepository(migrator)

	// Prepare initial metadata
	initialMetadata := &sharedv1.FileMetadata{
		FileId:            "test-upsert-id",
		OriginalFilename:  "initial_file.txt",
		StoragePath:       "/tmp/initial_file.txt",
		FileSizeBytes:     1024,
		FileType:          "text/plain",
		UploadTimestamp:   time.Now().Unix(),
		UserId:            "test-user-1",
	}

	initialStorage := &models.Storage{
		ID:             "test-upsert-id",
		FileMetadata:   initialMetadata,
		StoragePath:    "/tmp/initial_file.txt",
		ProcessingStatus: "PENDING",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	// First store
	err = repo.Store(context.Background(), initialStorage)
	require.NoError(t, err, "First store should succeed")

	// Update metadata (create a new instance to avoid mutex copy)
	updatedMetadata := sharedv1.FileMetadata{
		FileId:            "test-upsert-id",
		OriginalFilename:  "updated_file.txt",
		StoragePath:       "/tmp/updated_file.txt",
		FileSizeBytes:     initialMetadata.FileSizeBytes,
		FileType:          initialMetadata.FileType,
		UploadTimestamp:   initialMetadata.UploadTimestamp,
		UserId:            initialMetadata.UserId,
	}

	updatedStorage := &models.Storage{
		ID:             "test-upsert-id",
		FileMetadata:   &updatedMetadata,
		StoragePath:    "/tmp/updated_file.txt",
		ProcessingStatus: "PROCESSING",
		CreatedAt:      initialStorage.CreatedAt,
		UpdatedAt:      time.Now().UTC(),
	}

	// Update (upsert)
	err = repo.Store(context.Background(), updatedStorage)
	require.NoError(t, err, "Upsert should succeed")

	// Retrieve and verify
	storedFile, err := repo.FindByID(context.Background(), "test-upsert-id")
	require.NoError(t, err, "Should find updated file")
	
	assert.Equal(t, "updated_file.txt", storedFile.FileMetadata.OriginalFilename, "Filename should be updated")
	assert.Equal(t, "/tmp/updated_file.txt", storedFile.StoragePath, "Storage path should be updated")
	assert.Equal(t, "PROCESSING", storedFile.ProcessingStatus, "Processing status should be updated")
}
