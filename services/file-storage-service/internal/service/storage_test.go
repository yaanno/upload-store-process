package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"io"

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

// MockStorageProvider implements FileStorageProvider for testing
type MockStorageProvider struct {
	mock.Mock
}

func (m *MockStorageProvider) StoreFile(
	ctx context.Context,
	fileID string,
	originalFilename string,
	fileReader io.Reader,
) (string, error) {
	args := m.Called(ctx, fileID, originalFilename, fileReader)
	return args.String(0), args.Error(1)
}

func (m *MockStorageProvider) RetrieveFile(
	ctx context.Context,
	storagePath string,
) (io.Reader, error) {
	args := m.Called(ctx, storagePath)
	return args.Get(0).(io.Reader), args.Error(1)
}

func (m *MockStorageProvider) DeleteFile(
	ctx context.Context,
	storagePath string,
) error {
	args := m.Called(ctx, storagePath)
	return args.Error(0)
}

func (m *MockStorageProvider) GenerateStoragePath(
	fileID string,
	originalFilename string,
) string {
	args := m.Called(fileID, originalFilename)
	return args.String(0)
}

// Helper function to create a test logger
func createTestLogger() logger.Logger {
	return logger.Logger{} // You might want to use a mock or no-op logger in tests
}

func TestPrepareUpload(t *testing.T) {
	testCases := []struct {
		name           string
		request        *storagev1.PrepareUploadRequest
		mockBehavior   func(*MockFileMetadataRepository, *MockStorageProvider)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Successful Upload Preparation",
			request: &storagev1.PrepareUploadRequest{
				Filename:      "test.txt",
				FileSizeBytes: 1024,
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
				// Expect storage path generation
				msp.On("GenerateStoragePath", mock.Anything, "test.txt").
					Return("uploads/2025/02/06/test_file_id.txt")

				// Expect metadata creation
				mfmr.On("CreateFileMetadata", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:           "Nil Request",
			request:        nil,
			mockBehavior:   func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {},
			expectedError:  true,
			expectedErrMsg: "upload request cannot be nil",
		},
		{
			name: "Empty Filename",
			request: &storagev1.PrepareUploadRequest{
				Filename:      "",
				FileSizeBytes: 1024,
			},
			mockBehavior:   func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {},
			expectedError:  true,
			expectedErrMsg: "filename is required",
		},
		{
			name: "Invalid File Size",
			request: &storagev1.PrepareUploadRequest{
				Filename:      "test.txt",
				FileSizeBytes: 0,
			},
			mockBehavior:   func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {},
			expectedError:  true,
			expectedErrMsg: "invalid file size",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockRepo := new(MockFileMetadataRepository)
			mockStorageProvider := new(MockStorageProvider)
			mockLogger := createTestLogger()

			// Set up mock behaviors
			tc.mockBehavior(mockRepo, mockStorageProvider)

			// Create service with mocks
			service := NewFileStorageService(mockRepo, mockLogger, mockStorageProvider)

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
				assert.NotEmpty(t, resp.UploadToken)
				assert.NotEmpty(t, resp.StoragePath)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
			mockStorageProvider.AssertExpectations(t)
		})
	}
}

func TestCompleteUpload(t *testing.T) {
	testCases := []struct {
		name           string
		request        *storagev1.CompleteUploadRequest
		mockBehavior   func(*MockFileMetadataRepository, *MockStorageProvider)
		expectedError  bool
		expectedErrMsg string
		expectedFileID string
	}{
		{
			name: "Successful Upload Completion",
			request: &storagev1.CompleteUploadRequest{
				UploadId: "file_123",
				FileMetadata: &sharedv1.FileMetadata{
					OriginalFilename: "test_file.txt",
					FileSizeBytes:    1024,
					UploadTimestamp:  time.Now().Unix(),
				},
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
				// Mock creating metadata with specific file ID
				mfmr.On("CreateFileMetadata", mock.Anything, mock.MatchedBy(func(metadata *models.FileMetadataRecord) bool {
					return metadata.ID == "file_123"
				})).Return(nil)
			},
			expectedError:  false,
			expectedFileID: "file_123",
		},
		{
			name:           "Nil Request",
			request:        nil,
			mockBehavior:   func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {},
			expectedError:  true,
			expectedErrMsg: "complete upload request cannot be nil",
		},
		{
			name: "Empty Upload ID",
			request: &storagev1.CompleteUploadRequest{
				UploadId: "",
				FileMetadata: &sharedv1.FileMetadata{
					OriginalFilename: "test_file.txt",
				},
			},
			mockBehavior:   func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {},
			expectedError:  true,
			expectedErrMsg: "upload ID is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockRepo := new(MockFileMetadataRepository)
			mockStorageProvider := new(MockStorageProvider)
			mockLogger := createTestLogger()

			// Set up mock behaviors
			tc.mockBehavior(mockRepo, mockStorageProvider)

			// Create service with mocks
			service := NewFileStorageService(mockRepo, mockLogger, mockStorageProvider)

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
				assert.Equal(t, tc.expectedFileID, resp.ProcessedFileId)
				assert.True(t, resp.ProcessingStarted)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
			mockStorageProvider.AssertExpectations(t)
		})
	}
}

func TestListFiles(t *testing.T) {
	testCases := []struct {
		name           string
		request        *storagev1.ListFilesRequest
		mockBehavior   func(*MockFileMetadataRepository, *MockStorageProvider)
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
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
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
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
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
			mockBehavior:   func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {},
			expectedError:  true,
			expectedErrMsg: "list files request cannot be nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockRepo := new(MockFileMetadataRepository)
			mockStorageProvider := new(MockStorageProvider)
			mockLogger := createTestLogger()

			// Set up mock behaviors
			tc.mockBehavior(mockRepo, mockStorageProvider)

			// Create service with mocks
			service := NewFileStorageService(mockRepo, mockLogger, mockStorageProvider)

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
			mockStorageProvider.AssertExpectations(t)
		})
	}
}

func TestPrepareUploadBasic(t *testing.T) {
	mockRepo := new(MockFileMetadataRepository)
	mockStorageProvider := new(MockStorageProvider)
	mockLogger := createTestLogger()

	service := &fileStorageService{
		repo:            mockRepo,
		logger:          mockLogger,
		storageProvider: mockStorageProvider,
	}

	ctx := context.Background()
	req := &storagev1.PrepareUploadRequest{
		Filename:      "test.txt",
		FileSizeBytes: 1024,
	}

	mockStorageProvider.On("GenerateStoragePath", mock.Anything, "test.txt").
		Return("uploads/2025/02/06/test_file_id.txt")

	mockRepo.On("CreateFileMetadata", mock.Anything, mock.Anything).Return(nil)

	resp, err := service.PrepareUpload(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.UploadToken)
	assert.NotEmpty(t, resp.StoragePath)

	mockRepo.AssertExpectations(t)
	mockStorageProvider.AssertExpectations(t)
}

func TestCompleteUploadBasic(t *testing.T) {
	// Create mock repository
	mockRepo := new(MockFileMetadataRepository)
	mockRepo.On("CreateFileMetadata", mock.Anything, mock.Anything).Return(nil)

	// Create mock storage provider
	mockStorageProvider := new(MockStorageProvider)

	ctx := context.Background()
	req := &storagev1.CompleteUploadRequest{
		UploadId: "file_123",
		FileMetadata: &sharedv1.FileMetadata{
			OriginalFilename: "test_file.txt",
			FileSizeBytes:    1024,
			UploadTimestamp:  time.Now().Unix(),
		},
	}

	// Create service
	service := NewFileStorageService(mockRepo, createTestLogger(), mockStorageProvider)

	// Call method
	resp, err := service.CompleteUpload(ctx, req)

	// Validate results
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "file_123", resp.ProcessedFileId)
	assert.True(t, resp.ProcessingStarted)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
	mockStorageProvider.AssertExpectations(t)
}

func TestListFilesBasic(t *testing.T) {
	mockRepo := new(MockFileMetadataRepository)
	mockStorageProvider := new(MockStorageProvider)
	mockLogger := createTestLogger()

	service := &fileStorageService{
		repo:            mockRepo,
		logger:          mockLogger,
		storageProvider: mockStorageProvider,
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
	mockStorageProvider.AssertExpectations(t)
}
