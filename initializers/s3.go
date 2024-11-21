package initializers

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
	"hr-tools-backend/config"
	s3client "hr-tools-backend/s3"
)

func InitS3() {
	minioClient, err := minio.New(config.Conf.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Conf.S3.AccessKeyID, config.Conf.S3.SecretAccessKey, ""),
		Secure: *config.Conf.S3.UseSSL,
	})
	if err != nil {
		log.WithError(err).Error("Ошибка инициализации клиента S3")
		return
	}
	s3client.Client = minioClient
}
