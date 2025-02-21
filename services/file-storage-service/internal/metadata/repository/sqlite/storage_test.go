package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"io"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/storage"

	// "github.com/yaanno/upload-store-process/services/file-storage-service/internal/service"
	"github.com/yaanno/upload-store-process/services/shared/pkg/auth"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

// MockFileMetadataRepository is a mock implementation of FileMetadataRepository
type MockFileMetadataRepository struct {
	mock.Mock
}

// UpdateFileMetadata implements repository.FileMetadataRepository.
func (m *MockFileMetadataRepository) UpdateFileMetadata(ctx context.Context, metadata *domain.FileMetadataRecord) error {
	args := m.Called(ctx, metadata)
	return args.Error(0)
}

// RemoveFileMetadata implements repository.FileMetadataRepository.
func (m *MockFileMetadataRepository) IsFileOwnedByUser(ctx context.Context, options *domain.FileMetadataListOptions) (bool, error) {
	args := m.Called(ctx, options)
	return args.Bool(0), args.Error(1)
}

func (m *MockFileMetadataRepository) SoftDeleteFile(ctx context.Context, fileID, userID string) error {
	args := m.Called(ctx, fileID, userID)
	return args.Error(0)
}

// CreateFileMetadata implements repository.FileMetadataRepository.
func (m *MockFileMetadataRepository) CreateFileMetadata(ctx context.Context, metadata *domain.FileMetadataRecord) error {
	args := m.Called(ctx, metadata)
	return args.Error(0)
}

func (m *MockFileMetadataRepository) RetrieveFileMetadataByID(ctx context.Context, fileID string) (*domain.FileMetadataRecord, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FileMetadataRecord), args.Error(1)
}

func (m *MockFileMetadataRepository) ListFileMetadata(ctx context.Context, opts *domain.FileMetadataListOptions) ([]*domain.FileMetadataRecord, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FileMetadataRecord), args.Error(1)
}

func (m *MockFileMetadataRepository) ListFiles(ctx context.Context, opts *domain.FileMetadataListOptions) ([]*domain.FileMetadataRecord, int, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, 0, args.Error(1)
	}
	return args.Get(0).([]*domain.FileMetadataRecord), args.Int(1), args.Error(2)
}

func (m *MockFileMetadataRepository) RemoveFileMetadata(ctx context.Context, fileID string) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *MockFileMetadataRepository) GetFileMetadata(ctx context.Context, fileID string) (*domain.FileMetadataRecord, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FileMetadataRecord), args.Error(1)
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

// Mock token validator for testing
type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateToken(tokenString string) (*auth.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

// Helper function to create a test logger
func createTestLogger() logger.Logger {
	return logger.Logger{} // You might want to use a mock or no-op logger in tests
}

// func TestPrepareUpload(t *testing.T) {
// 	testCases := []struct {
// 		name           string
// 		request        *storagev1.PrepareUploadRequest
// 		mockBehavior   func(*MockFileMetadataRepository, *MockStorageProvider)
// 		expectedError  bool
// 		expectedErrMsg string
// 	}{
// 		{
// 			name: "Successful Upload Preparation",
// 			request: &storagev1.PrepareUploadRequest{
// 				GlobalUploadId: "upload_123",
// 				Filename:       "test.txt",
// 				FileSizeBytes:  1024,
// 				UserId:         "user_123",
// 			},
// 			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
// 				// Mock token validation
// 				// mtv.On("ValidateToken", "valid_token").Return(&auth.Claims{
// 				// 	UserID: "test-user",
// 				// }, nil)

// 				// Expect storage path generation
// 				msp.On("GenerateStoragePath", mock.Anything, "test.txt").
// 					Return("uploads/2025/02/06/test_file_id.txt")

// 				// Expect metadata creation
// 				mfmr.On("CreateFileMetadata", mock.Anything, mock.Anything).Return(nil)
// 			},
// 			expectedError: false,
// 		},
// 		{
// 			name:           "Nil Request",
// 			request:        nil,
// 			mockBehavior:   func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {},
// 			expectedError:  true,
// 			expectedErrMsg: "upload request cannot be nil",
// 		},
// 		{
// 			name: "Empty Filename",
// 			request: &storagev1.PrepareUploadRequest{
// 				GlobalUploadId: "upload_123",
// 				Filename:       "",
// 				FileSizeBytes:  1024,
// 				UserId:         "user_123",
// 			},
// 			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
// 			},
// 			expectedError:  true,
// 			expectedErrMsg: "filename is required",
// 		},
// 		{
// 			name: "Invalid File Size",
// 			request: &storagev1.PrepareUploadRequest{
// 				GlobalUploadId: "upload_123",
// 				Filename:       "test.txt",
// 				FileSizeBytes:  0,
// 				UserId:         "user_123",
// 			},
// 			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
// 				// Mock token validation with error
// 				// mtv.On("ValidateToken", "invalid_token").Return(nil, errors.New("invalid token"))
// 			},
// 			expectedError:  true,
// 			expectedErrMsg: "invalid file size",
// 		},
// 		{
// 			name: "Invalid Token",
// 			request: &storagev1.PrepareUploadRequest{
// 				GlobalUploadId: "upload_123",
// 				Filename:       "test.txt",
// 				FileSizeBytes:  1024,
// 				UserId:         "",
// 			},
// 			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
// 				// Mock token validation with error
// 				// mtv.On("ValidateToken", "invalid_token").Return(nil, errors.New("invalid token"))
// 			},
// 			expectedError:  true,
// 			expectedErrMsg: " user ID is empty",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			mockRepo := new(MockFileMetadataRepository)
// 			mockStorageProvider := new(MockStorageProvider)
// 			mockLogger := createTestLogger()

// 			// Set up mock behaviors
// 			tc.mockBehavior(mockRepo, mockStorageProvider)

// 			// Create service with mocks
// 			service := service.NewFileStorageService(
// 				mockRepo,
// 				mockLogger,
// 				mockStorageProvider,
// 			)

// 			// Call method
// 			resp, err := service.PrepareUpload(context.Background(), tc.request)

// 			// Validate results
// 			if tc.expectedError {
// 				assert.Error(t, err)
// 				if tc.expectedErrMsg != "" {
// 					// assert.Contains(t, err.Error(), tc.expectedErrMsg)
// 				}
// 				assert.Nil(t, resp)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, resp)
// 				assert.NotEmpty(t, resp.StorageUploadToken)
// 				assert.NotEmpty(t, resp.StoragePath)
// 			}

// 			// Verify mock expectations
// 			mockRepo.AssertExpectations(t)
// 			mockStorageProvider.AssertExpectations(t)
// 		})
// 	}
// }

// func TestListFiles(t *testing.T) {
// 	testCases := []struct {
// 		name           string
// 		request        *storagev1.ListFilesRequest
// 		mockBehavior   func(*MockFileMetadataRepository, *MockStorageProvider)
// 		expectedError  bool
// 		expectedErrMsg string
// 		expectedCount  int
// 	}{
// 		{
// 			name: "Successful File Listing",
// 			request: &storagev1.ListFilesRequest{
// 				PageSize: 10,
// 				Page:     1,
// 				UserId:   "user_123",
// 			},
// 			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
// 				mockFileMetadata := []*domain.FileMetadataRecord{
// 					{
// 						ID: "file1",
// 						Metadata: &sharedv1.FileMetadata{
// 							OriginalFilename: "test1.txt",
// 						},
// 					},
// 					{
// 						ID: "file2",
// 						Metadata: &sharedv1.FileMetadata{
// 							OriginalFilename: "test2.txt",
// 						},
// 					},
// 				}

// 				// Expect repository call
// 				mfmr.On("ListFiles", mock.Anything, mock.Anything).Return(mockFileMetadata, 2, nil)
// 			},
// 			expectedError: false,
// 			expectedCount: 2,
// 		},
// 		{
// 			name: "Page Size Zero",
// 			request: &storagev1.ListFilesRequest{
// 				PageSize: 0,
// 				Page:     1,
// 				UserId:   "user_123",
// 			},
// 			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
// 				mockFileMetadata := []*domain.FileMetadataRecord{
// 					{
// 						ID: "file1",
// 						Metadata: &sharedv1.FileMetadata{
// 							OriginalFilename: "test1.txt",
// 						},
// 					},
// 				}

// 				// Expect repository call with default page size
// 				mfmr.On("ListFiles", mock.Anything, mock.Anything).Return(mockFileMetadata, 1, nil)
// 			},
// 			expectedError: true,
// 		},
// 		// {
// 		// 	name:           "Nil Request",
// 		// 	request:        nil,
// 		// 	mockBehavior:   func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {},
// 		// 	expectedError:  true,
// 		// 	expectedErrMsg: "list files request cannot be nil",
// 		// },
// 		{
// 			name: "Invalid User ID",
// 			request: &storagev1.ListFilesRequest{
// 				PageSize: 10,
// 				Page:     1,
// 				UserId:   "",
// 			},
// 			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider) {
// 				// Mock token validation with error
// 				// mtv.On("ValidateToken", "invalid_token").Return(nil, errors.New("invalid token"))
// 			},
// 			expectedError:  true,
// 			expectedErrMsg: "user ID is required",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			mockRepo := new(MockFileMetadataRepository)
// 			mockStorageProvider := new(MockStorageProvider)
// 			mockLogger := createTestLogger()

// 			// Set up mock behaviors
// 			tc.mockBehavior(mockRepo, mockStorageProvider)

// 			// Create service with mocks
// 			service := service.NewFileStorageService(
// 				mockRepo,
// 				mockLogger,
// 				mockStorageProvider,
// 			)

// 			// Call method
// 			resp, err := service.ListFiles(context.Background(), tc.request)

// 			// Validate results
// 			if tc.expectedError {
// 				assert.Error(t, err)
// 				if tc.expectedErrMsg != "" {
// 					assert.Contains(t, err.Error(), tc.expectedErrMsg)
// 				}
// 				assert.Nil(t, resp)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, resp)
// 				assert.Equal(t, tc.expectedCount, len(resp.Files))
// 			}

// 			mockStorageProvider.AssertExpectations(t)
// 		})
// 	}
// }

func TestPrepareUploadBasic(t *testing.T) {
	mockRepo := new(MockFileMetadataRepository)
	mockStorageProvider := new(MockStorageProvider)
	mockLogger := createTestLogger()

	service := storage.NewFileStorageService(
		mockRepo,
		mockLogger,
		mockStorageProvider,
	)

	ctx := context.Background()
	req := &storagev1.PrepareUploadRequest{
		GlobalUploadId: "upload_123",
		Filename:       "test.txt",
		FileSizeBytes:  1024,
		UserId:         "user_123",
	}

	mockStorageProvider.On("GenerateStoragePath", mock.Anything, "test.txt").
		Return("uploads/2025/02/06/test_file_id.txt")

	mockRepo.On("CreateFileMetadata", mock.Anything, mock.Anything).Return(nil)

	resp, err := service.PrepareUpload(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.StorageUploadToken)
	assert.NotEmpty(t, resp.StoragePath)

	mockRepo.AssertExpectations(t)
	mockStorageProvider.AssertExpectations(t)
}

func TestListFilesBasic(t *testing.T) {
	mockRepo := new(MockFileMetadataRepository)
	mockStorageProvider := new(MockStorageProvider)
	mockLogger := createTestLogger()

	service := storage.NewFileStorageService(
		mockRepo,
		mockLogger,
		mockStorageProvider,
	)

	ctx := context.Background()
	req := &storagev1.ListFilesRequest{
		PageSize: 10,
		Page:     1,
		UserId:   "user_123",
	}

	mockFileMetadata := []*domain.FileMetadataRecord{
		{
			ID: "file1",
			Metadata: &sharedv1.FileMetadata{
				OriginalFilename: "test1.txt",
			},
		},
	}

	mockRepo.On("ListFiles", mock.Anything, mock.Anything).Return(mockFileMetadata, 1, nil)

	resp, err := service.ListFiles(ctx, req)

	assert.NoError(t, err)
	assert.Len(t, resp.Files, 1)
	assert.Equal(t, int32(len(mockFileMetadata)), resp.TotalFiles)
	assert.Equal(t, int32(1), resp.TotalPages)

	mockRepo.AssertExpectations(t)
	mockStorageProvider.AssertExpectations(t)
}
