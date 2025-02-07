package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"io"

	storagev1 "github.com/yaanno/upload-store-process/gen/go/filestorage/v1"
	sharedv1 "github.com/yaanno/upload-store-process/gen/go/shared/v1"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/models"
	"github.com/yaanno/upload-store-process/services/file-storage-service/internal/repository"
	"github.com/yaanno/upload-store-process/services/shared/pkg/auth"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
)

// MockFileMetadataRepository is a mock implementation of FileMetadataRepository
type MockFileMetadataRepository struct {
	mock.Mock
}

// UpdateFileMetadata implements repository.FileMetadataRepository.
func (m *MockFileMetadataRepository) UpdateFileMetadata(ctx context.Context, metadata *models.FileMetadataRecord) error {
	args := m.Called(ctx, metadata)
	return args.Error(0)
}

// CreateFileMetadata implements repository.FileMetadataRepository.
func (m *MockFileMetadataRepository) CreateFileMetadata(ctx context.Context, metadata *models.FileMetadataRecord) error {
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

func (m *MockFileMetadataRepository) ListFiles(ctx context.Context, opts *repository.FileMetadataListOptions) ([]*models.FileMetadataRecord, int, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, 0, args.Error(1)
	}
	return args.Get(0).([]*models.FileMetadataRecord), args.Int(1), args.Error(2)
}

func (m *MockFileMetadataRepository) RemoveFileMetadata(ctx context.Context, fileID string) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *MockFileMetadataRepository) GetFileMetadata(ctx context.Context, fileID string) (*models.FileMetadataRecord, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileMetadataRecord), args.Error(1)
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

func TestPrepareUpload(t *testing.T) {
	testCases := []struct {
		name           string
		request        *storagev1.PrepareUploadRequest
		mockBehavior   func(*MockFileMetadataRepository, *MockStorageProvider, *MockTokenValidator)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Successful Upload Preparation",
			request: &storagev1.PrepareUploadRequest{
				GlobalUploadId: "upload_123",
				Filename:       "test.txt",
				FileSizeBytes:  1024,
				JwtToken:       "valid_token",
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider, mtv *MockTokenValidator) {
				// Mock token validation
				mtv.On("ValidateToken", "valid_token").Return(&auth.Claims{
					UserID: "test-user",
				}, nil)

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
			mockBehavior:   func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider, mtv *MockTokenValidator) {},
			expectedError:  true,
			expectedErrMsg: "upload request cannot be nil",
		},
		{
			name: "Empty Filename",
			request: &storagev1.PrepareUploadRequest{
				GlobalUploadId: "upload_123",
				Filename:       "",
				FileSizeBytes:  1024,
				JwtToken:       "invalid_filename_token",
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider, mtv *MockTokenValidator) {
				// Mock token validation
				mtv.On("ValidateToken", "invalid_filename_token").Return(&auth.Claims{
					UserID: "test-user",
				}, nil)
			},
			expectedError:  true,
			expectedErrMsg: "filename is required",
		},
		{
			name: "Invalid File Size",
			request: &storagev1.PrepareUploadRequest{
				GlobalUploadId: "upload_123",
				Filename:       "test.txt",
				FileSizeBytes:  0,
				JwtToken:       "invalid_size_token",
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider, mtv *MockTokenValidator) {
				// Mock token validation
				mtv.On("ValidateToken", "invalid_size_token").Return(&auth.Claims{
					UserID: "test-user",
				}, nil)
			},
			expectedError:  true,
			expectedErrMsg: "invalid file size",
		},
		{
			name: "Invalid Token",
			request: &storagev1.PrepareUploadRequest{
				GlobalUploadId: "upload_123",
				Filename:       "test.txt",
				FileSizeBytes:  1024,
				JwtToken:       "invalid_token",
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider, mtv *MockTokenValidator) {
				// Mock token validation with error
				mtv.On("ValidateToken", "invalid_token").Return(nil, errors.New("invalid token"))
			},
			expectedError:  true,
			expectedErrMsg: "invalid JWT token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockFileMetadataRepository)
			mockStorageProvider := new(MockStorageProvider)
			mockLogger := createTestLogger()
			mockTokenValidator := new(MockTokenValidator)

			// Set up mock behaviors
			tc.mockBehavior(mockRepo, mockStorageProvider, mockTokenValidator)

			// Create service with mocks
			service := NewFileStorageService(
				mockRepo,
				mockLogger,
				mockStorageProvider,
				mockTokenValidator,
			)

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
				assert.NotEmpty(t, resp.StorageUploadToken)
				assert.NotEmpty(t, resp.StoragePath)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
			mockStorageProvider.AssertExpectations(t)
			mockTokenValidator.AssertExpectations(t)
		})
	}
}

func TestListFiles(t *testing.T) {
	testCases := []struct {
		name           string
		request        *storagev1.ListFilesRequest
		mockBehavior   func(*MockFileMetadataRepository, *MockStorageProvider, *MockTokenValidator)
		expectedError  bool
		expectedErrMsg string
		expectedCount  int
	}{
		{
			name: "Successful File Listing",
			request: &storagev1.ListFilesRequest{
				JwtToken: "valid_token",
				PageSize: 10,
				Page:     1,
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider, mtv *MockTokenValidator) {
				// Mock token validation
				mtv.On("ValidateToken", "valid_token").Return(&auth.Claims{
					UserID: "test-user",
				}, nil)

				mockFileMetadata := []*models.FileMetadataRecord{
					{
						ID: "file1",
						Metadata: &sharedv1.FileMetadata{
							OriginalFilename: "test1.txt",
						},
					},
					{
						ID: "file2",
						Metadata: &sharedv1.FileMetadata{
							OriginalFilename: "test2.txt",
						},
					},
				}

				// Expect repository call
				mfmr.On("ListFiles", mock.Anything, mock.Anything).Return(mockFileMetadata, 2, nil)
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name: "Page Size Zero",
			request: &storagev1.ListFilesRequest{
				JwtToken: "valid_token_zero_page",
				PageSize: 0,
				Page:     1,
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider, mtv *MockTokenValidator) {
				// Mock token validation
				mtv.On("ValidateToken", "valid_token_zero_page").Return(&auth.Claims{
					UserID: "test-user",
				}, nil)

				mockFileMetadata := []*models.FileMetadataRecord{
					{
						ID: "file1",
						Metadata: &sharedv1.FileMetadata{
							OriginalFilename: "test1.txt",
						},
					},
				}

				// Expect repository call with default page size
				mfmr.On("ListFiles", mock.Anything, mock.Anything).Return(mockFileMetadata, 1, nil)
			},
			expectedError: false,
			expectedCount: 1,
		},
		{
			name:           "Nil Request",
			request:        nil,
			mockBehavior:   func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider, mtv *MockTokenValidator) {},
			expectedError:  true,
			expectedErrMsg: "list files request cannot be nil",
		},
		{
			name: "Invalid Token",
			request: &storagev1.ListFilesRequest{
				JwtToken: "invalid_token",
				PageSize: 10,
				Page:     1,
			},
			mockBehavior: func(mfmr *MockFileMetadataRepository, msp *MockStorageProvider, mtv *MockTokenValidator) {
				// Mock token validation with error
				mtv.On("ValidateToken", "invalid_token").Return(nil, errors.New("invalid token"))
			},
			expectedError:  true,
			expectedErrMsg: "invalid JWT token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockFileMetadataRepository)
			mockStorageProvider := new(MockStorageProvider)
			mockLogger := createTestLogger()
			mockTokenValidator := new(MockTokenValidator)

			// Set up mock behaviors
			tc.mockBehavior(mockRepo, mockStorageProvider, mockTokenValidator)

			// Create service with mocks
			service := NewFileStorageService(
				mockRepo,
				mockLogger,
				mockStorageProvider,
				mockTokenValidator,
			)

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
				assert.Equal(t, tc.expectedCount, len(resp.Files))
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
			mockStorageProvider.AssertExpectations(t)
			mockTokenValidator.AssertExpectations(t)
		})
	}
}

func TestPrepareUploadBasic(t *testing.T) {
	mockRepo := new(MockFileMetadataRepository)
	mockStorageProvider := new(MockStorageProvider)
	mockLogger := createTestLogger()
	mockTokenValidator := new(MockTokenValidator)

	service := &FileStorageServiceImpl{
		repo:            mockRepo,
		logger:          mockLogger,
		storageProvider: mockStorageProvider,
		tokenValidator:  mockTokenValidator,
	}

	ctx := context.Background()
	req := &storagev1.PrepareUploadRequest{
		GlobalUploadId: "upload_123",
		Filename:       "test.txt",
		FileSizeBytes:  1024,
		JwtToken:       "valid_token",
	}

	mockTokenValidator.On("ValidateToken", "valid_token").Return(&auth.Claims{
		UserID: "test-user",
	}, nil)

	mockStorageProvider.On("GenerateStoragePath", mock.Anything, "test.txt").
		Return("uploads/2025/02/06/test_file_id.txt")

	mockRepo.On("CreateFileMetadata", mock.Anything, mock.Anything).Return(nil)

	resp, err := service.PrepareUpload(ctx, req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.StorageUploadToken)
	assert.NotEmpty(t, resp.StoragePath)

	mockRepo.AssertExpectations(t)
	mockStorageProvider.AssertExpectations(t)
	mockTokenValidator.AssertExpectations(t)
}

func TestListFilesBasic(t *testing.T) {
	mockRepo := new(MockFileMetadataRepository)
	mockStorageProvider := new(MockStorageProvider)
	mockLogger := createTestLogger()
	mockTokenValidator := new(MockTokenValidator)

	service := &FileStorageServiceImpl{
		repo:            mockRepo,
		logger:          mockLogger,
		storageProvider: mockStorageProvider,
		tokenValidator:  mockTokenValidator,
	}

	ctx := context.Background()
	req := &storagev1.ListFilesRequest{
		JwtToken: "valid_token",
		PageSize: 10,
		Page:     1,
	}

	// Mock token validation
	mockTokenValidator.On("ValidateToken", "valid_token").Return(&auth.Claims{
		UserID: "test-user",
	}, nil)

	mockFileMetadata := []*models.FileMetadataRecord{
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
	mockTokenValidator.AssertExpectations(t)
}
