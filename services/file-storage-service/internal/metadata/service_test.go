package metadata

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	domain "github.com/yaanno/upload-store-process/services/file-storage-service/internal/domain/metadata"
	"github.com/yaanno/upload-store-process/services/shared/pkg/logger"
	"go.uber.org/mock/gomock"
)

func TestMetadataServiceImpl_UpdateFileMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockFileMetadataRepository(ctrl)
	testLogger := logger.Logger{Logger: zerolog.New(nil)}

	service := &MetadataServiceImpl{
		metadataRepo: mockRepo,
		logger:       &testLogger,
	}

	ctx := context.Background()
	fileID := "test-file-id"
	record := &domain.FileMetadataRecord{
		ID: fileID,
	}

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "begin transaction fails",
			setup: func() {
				mockRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("begin tx failed"))
			},
			wantErr: true,
		},
		{
			name: "successful update",
			setup: func() {
				tx := "tx-1"
				mockRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
				mockRepo.EXPECT().UpdateFileMetadata(gomock.Any(), record).Return(nil)
				mockRepo.EXPECT().CommitTx(gomock.Any(), tx).Return(nil)
			},
			wantErr: false,
		},

		{
			name: "update fails and rollback succeeds",
			setup: func() {
				tx := "tx-2"
				mockRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
				mockRepo.EXPECT().UpdateFileMetadata(gomock.Any(), record).Return(errors.New("update failed"))
				mockRepo.EXPECT().RollbackTx(gomock.Any(), tx).Return(nil).Times(1)
			},
			wantErr: true,
		},
		{
			name: "update fails and rollback fails",
			setup: func() {
				tx := "tx-3"
				mockRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
				mockRepo.EXPECT().UpdateFileMetadata(gomock.Any(), record).Return(errors.New("update failed"))
				mockRepo.EXPECT().RollbackTx(gomock.Any(), tx).Return(errors.New("rollback failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "update succeeds but commit fails",
			setup: func() {
				tx := "tx-4"
				mockRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
				mockRepo.EXPECT().UpdateFileMetadata(gomock.Any(), record).Return(nil)
				mockRepo.EXPECT().CommitTx(gomock.Any(), tx).Return(errors.New("commit failed"))
				// Add expectation for RollbackTx since commit failure should trigger rollback
				mockRepo.EXPECT().RollbackTx(gomock.Any(), tx).Return(nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := service.UpdateFileMetadata(ctx, fileID, record)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateFileMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
