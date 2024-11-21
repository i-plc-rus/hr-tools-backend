package filestorage

import (
	"context"
)

type Provider interface {
	UploadResume(ctx context.Context, spaceID, applicantID string, file []byte, fileName string) error
	UploadDoc(ctx context.Context, spaceID, applicantID string, file []byte, fileName string) error
	GetFile(ctx context.Context, spaceID, fileID string) ([]byte, error)
	GetResume(ctx context.Context, spaceID, applicantID string) ([]byte, error)
	GetDocList(ctx context.Context, applicantID string) ([]interface{}, error)
	MakeSpaceBucket(ctx context.Context, spaceID string) error
}

var Instance Provider
