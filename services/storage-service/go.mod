module github.com/yaanno/upload-store-process/services/storage-service

go 1.22

toolchain go1.22.4

require (
	google.golang.org/grpc v1.70.0
	google.golang.org/protobuf v1.36.4
)

replace (
	github.com/yaanno/upload-store-process/proto => ../../proto/gen
)
