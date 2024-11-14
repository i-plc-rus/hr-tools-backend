package s3client

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"hr-tools-backend/config"
)

var Client *minio.Client

type s3client struct {
	minioClient *minio.Client
}

func (s s3client) MakeBucket(ctx context.Context) error {

	// Make a new bucket called testbucket.
	bucketName := config.Conf.S3.BucketName
	location := "us-east-1"
	exists, err := s.minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = s.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		return err
	}
	return nil
}

func NewClient() (Provider, error) {
	minioClient, err := minio.New(config.Conf.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Conf.S3.AccessKeyID, config.Conf.S3.SecretAccessKey, ""),
		Secure: *config.Conf.S3.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &s3client{minioClient: minioClient}, nil
}
