package initializers

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/config"
	s3client "hr-tools-backend/s3"
)

func InitS3(ctx context.Context) {
	minioClient, err := minio.New(config.Conf.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Conf.S3.AccessKeyID, config.Conf.S3.SecretAccessKey, ""),
		Secure: *config.Conf.S3.UseSSL,
	})
	if err != nil {
		log.WithError(err).Error("Ошибка инициализации клиента S3")
		return
	}
	s3client.Client = minioClient
	err = MakeBucket(ctx, minioClient)
	if err != nil {
		log.WithError(err).Error("Ошибка создания бакета в S3")
	}
}

func MakeBucket(ctx context.Context, minioClient *minio.Client) error {

	bucketName := config.Conf.S3.BucketName
	location := "us-east-1"
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		return err
	}
	return nil
}
