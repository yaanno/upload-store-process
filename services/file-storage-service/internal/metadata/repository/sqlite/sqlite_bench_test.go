package repository_test

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	repository "github.com/yaanno/upload-store-process/services/file-storage-service/internal/metadata/repository/sqlite"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

func BenchmarkCreateFileMetadata(b *testing.B) {
	testDatabase, err := repository.InitializeTestDatabase()
	if err != nil {
		b.Fatalf("Failed to initialize test database: %v", err)
	}

	db := testDatabase.GetDB()
	defer db.Close()

	// Initialize repository
	repo := repository.NewSQLiteFileMetadataRepository(db, logger.Logger{})
	// if err != nil {
	// 	b.Fatalf("Failed to create repository: %v", err)
	// }

	// Benchmark with different batch sizes
	benchmarks := []struct {
		name      string
		batchSize int
	}{
		{"Single Record", 1},
		{"Small Batch", 10},
		{"Large Batch", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Prepare test data
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Create batch of file metadata records
				for j := 0; j < bm.batchSize; j++ {
					metadata := &domain.FileMetadataRecord{
						ID: fmt.Sprintf("file-%d-%d", i, j),
						Metadata: &sharedv1.FileMetadata{
							OriginalFilename: fmt.Sprintf("test-file-%d-%d.txt", i, j),
						},
					}

					err := repo.CreateFileMetadata(ctx, metadata)
					if err != nil {
						b.Fatalf("Failed to create metadata: %v", err)
					}
				}
			}
		})
	}
}

func BenchmarkListFileMetadata(b *testing.B) {
	testDatabase, err := repository.InitializeTestDatabase()
	if err != nil {
		b.Fatalf("Failed to initialize test database: %v", err)
	}

	db := testDatabase.GetDB()
	defer db.Close()

	repo := repository.NewSQLiteFileMetadataRepository(db, logger.Logger{})

	// Prepare database with test data
	ctx := context.Background()
	totalRecords := 1000
	for i := 0; i < totalRecords; i++ {
		metadata := &domain.FileMetadataRecord{
			ID: fmt.Sprintf("file-%d", i),
			Metadata: &sharedv1.FileMetadata{
				OriginalFilename: fmt.Sprintf("test-file-%d.txt", i),
			},
		}
		repo.CreateFileMetadata(ctx, metadata)
	}

	// Benchmark different pagination scenarios
	benchmarks := []struct {
		name     string
		pageSize int32
	}{
		{"Small Page", 10},
		{"Medium Page", 50},
		{"Large Page", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			listOpts := &domain.FileMetadataListOptions{
				UserID: "bench-user",
				Limit:  1,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				files, err := repo.ListFileMetadata(ctx, listOpts)
				if err != nil {
					b.Fatalf("Failed to list files: %v", err)
				}

				// Consume results to prevent compiler optimizations
				_ = files
			}
		})
	}
}

// Benchmark complex query performance
func BenchmarkComplexFileMetadataQuery(b *testing.B) {
	testDatabase, err := repository.InitializeTestDatabase()
	if err != nil {
		b.Fatalf("Failed to initialize test database: %v", err)
	}

	db := testDatabase.GetDB()
	defer db.Close()

	repo := repository.NewSQLiteFileMetadataRepository(db, logger.Logger{})

	// Prepare complex scenario
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Example of a more complex query operation
		complexOpts := &domain.FileMetadataListOptions{
			UserID:    "bench-user",
			Limit:     50,
			Offset:    0,
			SortBy:    "created_at",
			SortOrder: "DESC",
		}

		files, err := repo.ListFileMetadata(ctx, complexOpts)
		if err != nil {
			b.Fatalf("Failed to execute complex query: %v", err)
		}

		_ = files
	}
}
