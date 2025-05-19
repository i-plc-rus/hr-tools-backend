package initializers

import (
	"context"
	"hr-tools-backend/config"
	s3client "hr-tools-backend/s3"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

func InitS3() {
	minioClient, err := minio.New(config.Conf.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV2(config.Conf.S3.AccessKeyID, config.Conf.S3.SecretAccessKey, ""),
		Secure: *config.Conf.S3.UseSSL,
	})
	if err != nil {
		log.WithError(err).Error("Ошибка инициализации клиента S3")
		return
	}

	// Проверка соединения
	_, err = minioClient.ListBuckets(context.Background())
	if err != nil {
		log.WithError(err).Error("S3 соединение не удалось — ListBuckets вернул ошибку")
	}

	s3client.Client = minioClient
	log.Info("S3 клиент успешно инициализирован")
}
