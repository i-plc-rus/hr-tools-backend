package filestorage

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"hr-tools-backend/config"
	"io"
)

type Provider interface {
	UploadFile(ctx context.Context, spaceID string, fileReader io.Reader, fileSize int64) error
	GetFile(ctx context.Context, spaceID string, fileID string) ([]byte, error)
	MakeSpaceBucket(ctx context.Context, spaceID string) error
}

var Instance Provider

type impl struct {
	s3client *minio.Client
}

func (i impl) GetFile(ctx context.Context, spaceID string, fileID string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (i impl) UploadFile(ctx context.Context, spaceID string, fileID string, fileReader io.Reader, fileSize int64) error {
	_, err := i.s3client.PutObject(ctx, config.Conf.S3.BucketName, "file-id", fileReader, fileSize, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return err
	}
	return nil
}

func (i impl) MakeSpaceBucket(ctx context.Context, spaceID string) error {
	bucketName := i.getSpaceBucketName(spaceID)
	location := "us-east-1"
	exists, err := i.s3client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = i.s3client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		return err
	}
	return nil
}

func (i impl) getSpaceBucketName(spaceID string) string {
	return fmt.Sprintf("%s-%s", config.Conf.S3.BucketName, spaceID)
}

func NewInstance(s3client *minio.Client) {
	Instance = &impl{
		s3client: s3client,
	}
}
