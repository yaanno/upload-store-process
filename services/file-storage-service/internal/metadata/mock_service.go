// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/A200246910/workspace/upload-store-process/services/file-storage-service/internal/metadata/service.go
//
// Generated by this command:
//
//	mockgen -source=/Users/A200246910/workspace/upload-store-process/services/file-storage-service/internal/metadata/service.go -destination=/Users/A200246910/workspace/upload-store-process/services/file-storage-service/internal/metadata/mock_service.go -package=metadata MetadataService
//

// Package metadata is a generated GoMock package.
package metadata

import (
	context "context"
	reflect "reflect"

	metadata "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	gomock "go.uber.org/mock/gomock"
)

// MockMetadataService is a mock of MetadataService interface.
type MockMetadataService struct {
	ctrl     *gomock.Controller
	recorder *MockMetadataServiceMockRecorder
	isgomock struct{}
}

// MockMetadataServiceMockRecorder is the mock recorder for MockMetadataService.
type MockMetadataServiceMockRecorder struct {
	mock *MockMetadataService
}

// NewMockMetadataService creates a new mock instance.
func NewMockMetadataService(ctrl *gomock.Controller) *MockMetadataService {
	mock := &MockMetadataService{ctrl: ctrl}
	mock.recorder = &MockMetadataServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetadataService) EXPECT() *MockMetadataServiceMockRecorder {
	return m.recorder
}

// BeginTx mocks base method.
func (m *MockMetadataService) BeginTx(ctx context.Context) (context.Context, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginTx", ctx)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginTx indicates an expected call of BeginTx.
func (mr *MockMetadataServiceMockRecorder) BeginTx(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTx", reflect.TypeOf((*MockMetadataService)(nil).BeginTx), ctx)
}

// CleanupExpiredMetadata mocks base method.
func (m *MockMetadataService) CleanupExpiredMetadata(ctx context.Context) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CleanupExpiredMetadata", ctx)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CleanupExpiredMetadata indicates an expected call of CleanupExpiredMetadata.
func (mr *MockMetadataServiceMockRecorder) CleanupExpiredMetadata(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CleanupExpiredMetadata", reflect.TypeOf((*MockMetadataService)(nil).CleanupExpiredMetadata), ctx)
}

// CommitTx mocks base method.
func (m *MockMetadataService) CommitTx(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommitTx", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// CommitTx indicates an expected call of CommitTx.
func (mr *MockMetadataServiceMockRecorder) CommitTx(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommitTx", reflect.TypeOf((*MockMetadataService)(nil).CommitTx), ctx)
}

// CreateFileMetadata mocks base method.
func (m *MockMetadataService) CreateFileMetadata(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateFileMetadata", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateFileMetadata indicates an expected call of CreateFileMetadata.
func (mr *MockMetadataServiceMockRecorder) CreateFileMetadata(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFileMetadata", reflect.TypeOf((*MockMetadataService)(nil).CreateFileMetadata), ctx)
}

// DeleteFileMetadata mocks base method.
func (m *MockMetadataService) DeleteFileMetadata(ctx context.Context, userID, fileID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFileMetadata", ctx, userID, fileID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFileMetadata indicates an expected call of DeleteFileMetadata.
func (mr *MockMetadataServiceMockRecorder) DeleteFileMetadata(ctx, userID, fileID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFileMetadata", reflect.TypeOf((*MockMetadataService)(nil).DeleteFileMetadata), ctx, userID, fileID)
}

// GetFileMetadata mocks base method.
func (m *MockMetadataService) GetFileMetadata(ctx context.Context, userID, fileID string) (*metadata.FileMetadataRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFileMetadata", ctx, userID, fileID)
	ret0, _ := ret[0].(*metadata.FileMetadataRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFileMetadata indicates an expected call of GetFileMetadata.
func (mr *MockMetadataServiceMockRecorder) GetFileMetadata(ctx, userID, fileID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFileMetadata", reflect.TypeOf((*MockMetadataService)(nil).GetFileMetadata), ctx, userID, fileID)
}

// ListFileMetadata mocks base method.
func (m *MockMetadataService) ListFileMetadata(ctx context.Context, opts *metadata.FileMetadataListOptions) ([]*metadata.FileMetadataRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListFileMetadata", ctx, opts)
	ret0, _ := ret[0].([]*metadata.FileMetadataRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListFileMetadata indicates an expected call of ListFileMetadata.
func (mr *MockMetadataServiceMockRecorder) ListFileMetadata(ctx, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFileMetadata", reflect.TypeOf((*MockMetadataService)(nil).ListFileMetadata), ctx, opts)
}

// PrepareUpload mocks base method.
func (m *MockMetadataService) PrepareUpload(ctx context.Context, params *PrepareUploadParams) (*PrepareUploadResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrepareUpload", ctx, params)
	ret0, _ := ret[0].(*PrepareUploadResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PrepareUpload indicates an expected call of PrepareUpload.
func (mr *MockMetadataServiceMockRecorder) PrepareUpload(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrepareUpload", reflect.TypeOf((*MockMetadataService)(nil).PrepareUpload), ctx, params)
}

// RetrieveFileMetadataByID mocks base method.
func (m *MockMetadataService) RetrieveFileMetadataByID(ctx context.Context, fileID string) (*metadata.FileMetadataRecord, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RetrieveFileMetadataByID", ctx, fileID)
	ret0, _ := ret[0].(*metadata.FileMetadataRecord)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RetrieveFileMetadataByID indicates an expected call of RetrieveFileMetadataByID.
func (mr *MockMetadataServiceMockRecorder) RetrieveFileMetadataByID(ctx, fileID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RetrieveFileMetadataByID", reflect.TypeOf((*MockMetadataService)(nil).RetrieveFileMetadataByID), ctx, fileID)
}

// RollbackTx mocks base method.
func (m *MockMetadataService) RollbackTx(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RollbackTx", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// RollbackTx indicates an expected call of RollbackTx.
func (mr *MockMetadataServiceMockRecorder) RollbackTx(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RollbackTx", reflect.TypeOf((*MockMetadataService)(nil).RollbackTx), ctx)
}

// UpdateFileMetadata mocks base method.
func (m *MockMetadataService) UpdateFileMetadata(ctx context.Context, fileID string, record *metadata.FileMetadataRecord) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFileMetadata", ctx, fileID, record)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateFileMetadata indicates an expected call of UpdateFileMetadata.
func (mr *MockMetadataServiceMockRecorder) UpdateFileMetadata(ctx, fileID, record any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFileMetadata", reflect.TypeOf((*MockMetadataService)(nil).UpdateFileMetadata), ctx, fileID, record)
}
