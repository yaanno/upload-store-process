package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/repository"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

// MockFileMetadataRepository is a mock implementation of FileMetadataRepository
type MockFileMetadataRepository struct {
	mock.Mock
}

// RemoveFileMetadata implements repository.FileMetadataRepository.
func (m *MockFileMetadataRepository) RemoveFileMetadata(ctx context.Context, fileID string) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *MockFileMetadataRepository) CreateFileMetadata(ctx context.Context, metadata *models.FileMetadataRecord) error {
	args := m.Called(ctx, metadata)
	return args.Error(0)
}

func (m *MockFileMetadataRepository) UpdateFileMetadata(ctx context.Context, metadata *models.FileMetadataRecord) error {
	args := m.Called(ctx, metadata)
	return args.Error(0)
}

func (m *MockFileMetadataRepository) RetrieveFileMetadataByID(ctx context.Context, fileID string) (*models.FileMetadataRecord, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileMetadataRecord), args.Error(1)
}

func (m *MockFileMetadataRepository) ListFileMetadata(ctx context.Context, opts *repository.FileMetadataListOptions) ([]*models.FileMetadataRecord, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileMetadataRecord), args.Error(1)
}

// Helper function to create a test logger
func createTestLogger() *logger.Logger {
	return &logger.Logger{} // You might want to use a mock or no-op logger in tests
}

func TestPrepareUpload(t *testing.T) {
	testCases := []struct {
		name           string
		request        *storagev1.PrepareUploadRequest
		mockBehavior   func(*MockFileMetadataRepository)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Successful Upload Preparation",
			request: &storagev1.PrepareUploadRequest{
				Filename:      "test.jpg",
				FileSizeBytes: 1024,
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository) {
				mfmr.On("CreateFileMetadata", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:           "Nil Request",
			request:        nil,
			mockBehavior:   func(mfmr *MockFileMetadataRepository) {},
			expectedError:  true,
			expectedErrMsg: "upload request cannot be nil",
		},
		{
			name: "Empty Filename",
			request: &storagev1.PrepareUploadRequest{
				Filename:      "",
				FileSizeBytes: 1024,
			},
			mockBehavior:   func(mfmr *MockFileMetadataRepository) {},
			expectedError:  true,
			expectedErrMsg: "filename is required",
		},
		{
			name: "Invalid File Size",
			request: &storagev1.PrepareUploadRequest{
				Filename:      "test.jpg",
				FileSizeBytes: 0,
			},
			mockBehavior:   func(mfmr *MockFileMetadataRepository) {},
			expectedError:  true,
			expectedErrMsg: "invalid file size",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockFileMetadataRepository)

			// Set up mock behavior
			tc.mockBehavior(mockRepo)

			// Create service
			service := NewFileStorageService(mockRepo, createTestLogger())

			// Call method
			resp, err := service.PrepareUpload(context.Background(), tc.request)

			// Validate results
			if tc.expectedError {
				assert.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				// assert.NotEmpty(t, 1)
				// assert.NotEmpty(t, "")
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCompleteUpload(t *testing.T) {
	testCases := []struct {
		name           string
		request        *storagev1.CompleteUploadRequest
		mockBehavior   func(*MockFileMetadataRepository)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Successful Upload Completion",
			request: &storagev1.CompleteUploadRequest{
				UploadId: "file_123",
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository) {
				// Mock retrieving existing metadata
				existingMetadata := &models.FileMetadataRecord{
					ID:               "file_123",
					ProcessingStatus: "PENDING",
					Metadata: &sharedv1.FileMetadata{
						OriginalFilename: "test_file.txt",
					},
				}
				mfmr.On("RetrieveFileMetadataByID", mock.Anything, "file_123").Return(existingMetadata, nil)

				// Mock creating metadata (which is actually an update in this case)
				mfmr.On("CreateFileMetadata", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:           "Nil Request",
			request:        nil,
			mockBehavior:   func(mfmr *MockFileMetadataRepository) {},
			expectedError:  true,
			expectedErrMsg: "complete upload request cannot be nil",
		},
		{
			name: "Empty Upload ID",
			request: &storagev1.CompleteUploadRequest{
				UploadId: "",
			},
			mockBehavior:   func(mfmr *MockFileMetadataRepository) {},
			expectedError:  true,
			expectedErrMsg: "upload ID is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockFileMetadataRepository)

			// Set up mock behavior
			tc.mockBehavior(mockRepo)

			// Create service
			service := NewFileStorageService(mockRepo, createTestLogger())

			// Call method
			resp, err := service.CompleteUpload(context.Background(), tc.request)

			// Validate results
			if tc.expectedError {
				assert.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "file_123", resp.ProcessedFileId)
				assert.True(t, resp.ProcessingStarted)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListFiles(t *testing.T) {
	testCases := []struct {
		name           string
		request        *storagev1.ListFilesRequest
		mockBehavior   func(*MockFileMetadataRepository)
		expectedError  bool
		expectedErrMsg string
		expectedCount  int
		expectedPages  int32
	}{
		{
			name: "Successful File Listing",
			request: &storagev1.ListFilesRequest{
				PageSize: 10,
				Page:     1,
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository) {
				mockFileMetadata := []*models.FileMetadataRecord{
					{
						ID: "file1",
						Metadata: &sharedv1.FileMetadata{
							OriginalFilename: "test1.txt",
							FileSizeBytes:    1024,
							UploadTimestamp:  time.Now().Unix(),
						},
					},
					{
						ID: "file2",
						Metadata: &sharedv1.FileMetadata{
							OriginalFilename: "test2.txt",
							FileSizeBytes:    2048,
							UploadTimestamp:  time.Now().Unix(),
						},
					},
				}

				mfmr.On("ListFileMetadata", mock.Anything, mock.Anything).Return(mockFileMetadata, nil)
			},
			expectedError: false,
			expectedCount: 2,
			expectedPages: 1,
		},
		{
			name: "Zero Page Size",
			request: &storagev1.ListFilesRequest{
				PageSize: 0,
				Page:     1,
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository) {
				mockFileMetadata := []*models.FileMetadataRecord{
					{
						ID: "file1",
						Metadata: &sharedv1.FileMetadata{
							OriginalFilename: "test1.txt",
							FileSizeBytes:    1024,
							UploadTimestamp:  time.Now().Unix(),
						},
					},
				}

				mfmr.On("ListFileMetadata", mock.Anything, mock.Anything).Return(mockFileMetadata, nil)
			},
			expectedError: false,
			expectedCount: 1,
			expectedPages: 1, // Default page size is 10
		},
		{
			name:           "Nil Request",
			request:        nil,
			mockBehavior:   func(mfmr *MockFileMetadataRepository) {},
			expectedError:  true,
			expectedErrMsg: "list files request cannot be nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock repository
			mockRepo := new(MockFileMetadataRepository)

			// Set up mock behavior
			tc.mockBehavior(mockRepo)

			// Create service
			service := NewFileStorageService(mockRepo, createTestLogger())

			// Call method
			resp, err := service.ListFiles(context.Background(), tc.request)

			// Validate results
			if tc.expectedError {
				assert.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.Files, tc.expectedCount)
				assert.Equal(t, int32(tc.expectedCount), resp.TotalFiles)
				assert.Equal(t, tc.expectedPages, resp.TotalPages)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPrepareUploadBasic(t *testing.T) {
	mockRepo := new(MockFileMetadataRepository)
	mockLogger := &logger.Logger{}

	service := &FileStorageService{
		repo:   mockRepo,
		logger: mockLogger,
	}

	ctx := context.Background()
	req := &storagev1.PrepareUploadRequest{
		Filename:      "test_file.txt",
		FileSizeBytes: 1024,
	}

	mockRepo.On("CreateFileMetadata", mock.Anything, mock.Anything).Return(nil)

	resp, err := service.PrepareUpload(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.UploadToken)
	assert.NotEmpty(t, resp.StoragePath)

	mockRepo.AssertExpectations(t)
}

func TestCompleteUploadBasic(t *testing.T) {
	mockRepo := new(MockFileMetadataRepository)
	mockLogger := createTestLogger()

	service := &FileStorageService{
		repo:   mockRepo,
		logger: mockLogger,
	}

	ctx := context.Background()
	existingMetadata := &models.FileMetadataRecord{
		ID: "test-file-id",
		Metadata: &sharedv1.FileMetadata{
			OriginalFilename: "test_file.txt",
			UploadTimestamp:  time.Now().Unix(),
		},
	}

	req := &storagev1.CompleteUploadRequest{
		UploadId: "test-file-id",
	}

	mockRepo.On("RetrieveFileMetadataByID", mock.Anything, "test-file-id").Return(existingMetadata, nil)
	mockRepo.On("CreateFileMetadata", mock.Anything, mock.Anything).Return(nil)

	resp, err := service.CompleteUpload(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, existingMetadata.ID, resp.ProcessedFileId)
	assert.True(t, resp.ProcessingStarted)

	mockRepo.AssertExpectations(t)
}

func TestListFilesBasic(t *testing.T) {
	mockRepo := new(MockFileMetadataRepository)
	mockLogger := &logger.Logger{}

	service := &FileStorageService{
		repo:   mockRepo,
		logger: mockLogger,
	}

	ctx := context.Background()
	req := &storagev1.ListFilesRequest{
		PageSize: 10,
		Page:     1,
	}

	mockFileMetadata := []*models.FileMetadataRecord{
		{
			ID: "file1",
			Metadata: &sharedv1.FileMetadata{
				OriginalFilename: "test1.txt",
				FileSizeBytes:    1024,
				UploadTimestamp:  time.Now().Unix(),
			},
		},
	}

	mockRepo.On("ListFileMetadata", mock.Anything, mock.Anything).Return(mockFileMetadata, nil)

	resp, err := service.ListFiles(ctx, req)

	assert.NoError(t, err)
	assert.Len(t, resp.Files, 1)
	assert.Equal(t, int32(len(mockFileMetadata)), resp.TotalFiles)
	assert.Equal(t, int32(1), resp.TotalPages)

	mockRepo.AssertExpectations(t)
}
