package filestorage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"hr-tools-backend/config"
	"hr-tools-backend/db"
	filesdbstorage "hr-tools-backend/lib/file-storage/storage"
	filesapimodels "hr-tools-backend/models/api/files"
	dbmodels "hr-tools-backend/models/db"

	"io"
)

type Provider interface {
	UploadResume(ctx context.Context, spaceID, applicantID string, file []byte, fileName string) error
	UploadDoc(ctx context.Context, spaceID, applicantID string, file []byte, fileName string) error
	GetFile(ctx context.Context, spaceID, fileID string) ([]byte, error)
	GetResume(ctx context.Context, spaceID, applicantID string) ([]byte, error)
	GetDocList(ctx context.Context, applicantID string) ([]filesapimodels.FileView, error)
	MakeSpaceBucket(ctx context.Context, spaceID string) error
}

var Instance Provider

type impl struct {
	s3client       *minio.Client
	filesDBStorage filesdbstorage.Provider
}

func (i impl) GetDocList(ctx context.Context, applicantID string) (listView []filesapimodels.FileView, err error) {
	docList, err := i.filesDBStorage.GetFileListByType(applicantID, dbmodels.ApplicantDoc)
	if err != nil {
		log.
			WithField("applicant_id", applicantID).
			WithError(err).
			Error("ошибка получения списка документов кандидата")
		return nil, err
	}
	for _, doc := range docList {
		listView = append(listView, doc.ToModel())
	}
	return listView, nil
}

func (i impl) GetResume(ctx context.Context, spaceID, applicantID string) ([]byte, error) {
	logger := log.WithFields(log.Fields{
		"space_id":     spaceID,
		"applicant_id": applicantID,
	})
	fileID, err := i.filesDBStorage.GetFileIDByType(applicantID, dbmodels.ApplicantResume)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка поиска ID файла резюме")
		return nil, err
	}
	return i.GetFile(ctx, spaceID, fileID)
}

func (i impl) GetFile(ctx context.Context, spaceID string, fileID string) ([]byte, error) {
	logger := log.WithFields(log.Fields{
		"space_id": spaceID,
		"file_id":  fileID,
	})
	s3file, err := i.s3client.GetObject(ctx, i.getSpaceBucketName(spaceID), fileID, minio.GetObjectOptions{})
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка получения файла из S3")
		return nil, err
	}
	defer s3file.Close()
	body, err := io.ReadAll(s3file)
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка чтения файла из S3")
		return nil, err
	}
	return body, nil
}

func (i impl) UploadResume(ctx context.Context, spaceID, applicantID string, file []byte, fileName string) error {
	logger := log.WithFields(log.Fields{
		"space_id":     spaceID,
		"applicant_id": applicantID,
		"file_name":    fileName,
	})
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		fileDB := filesdbstorage.NewInstance(tx)
		rec := dbmodels.FileStorage{
			BaseSpaceModel: dbmodels.BaseSpaceModel{
				SpaceID: spaceID,
			},
			Name:        fileName,
			ApplicantID: applicantID,
			Type:        dbmodels.ApplicantResume,
		}
		fileID, err := fileDB.SaveFile(rec)
		if err != nil {
			return errors.Wrap(err, "ошибка сохранения информации о файле резюме")
		}
		err = i.uploadFile(ctx, i.getSpaceBucketName(spaceID), fileID, bytes.NewReader(file), len(file))
		if err != nil {
			return errors.Wrap(err, "ошибка загрузки резюме в S3")
		}
		return nil
	})
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка загрузки резюме кандидата")
		return err
	}
	return nil
}

func (i impl) UploadDoc(ctx context.Context, spaceID, applicantID string, file []byte, fileName string) error {
	logger := log.WithFields(log.Fields{
		"space_id":     spaceID,
		"applicant_id": applicantID,
		"file_name":    fileName,
	})
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		fileDB := filesdbstorage.NewInstance(tx)
		rec := dbmodels.FileStorage{
			BaseSpaceModel: dbmodels.BaseSpaceModel{
				SpaceID: spaceID,
			},
			Name:        fileName,
			ApplicantID: applicantID,
			Type:        dbmodels.ApplicantDoc,
		}
		fileID, err := fileDB.SaveFile(rec)
		if err != nil {
			return errors.Wrap(err, "ошибка сохранения информации о файле документа")
		}
		err = i.uploadFile(ctx, i.getSpaceBucketName(spaceID), fileID, bytes.NewReader(file), len(file))
		if err != nil {
			return errors.Wrap(err, "ошибка загрузки документа в S3")
		}
		return nil
	})
	if err != nil {
		logger.
			WithError(err).
			Error("ошибка загрузки документа кандидата")
		return err
	}
	return nil
}

func (i impl) uploadFile(ctx context.Context, bucketName, fileID string, fileReader io.Reader, fileSize int) error {
	_, err := i.s3client.PutObject(ctx, bucketName, fileID, fileReader, int64(fileSize), minio.PutObjectOptions{ContentType: "application/octet-stream"})
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
	return fmt.Sprintf("%s-%s", config.Conf.S3.BucketNamePrefix, spaceID)
}

func NewInstance(s3client *minio.Client) {
	Instance = &impl{
		s3client:       s3client,
		filesDBStorage: filesdbstorage.NewInstance(db.DB),
	}
}
