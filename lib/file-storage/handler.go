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
	s3client "hr-tools-backend/s3"
	"io"
)

type Provider interface {
	GetDocList(ctx context.Context, applicantID string) ([]filesapimodels.FileView, error)
	GetFileByType(ctx context.Context, spaceID, applicantID string, fileType dbmodels.FileType) ([]byte, *dbmodels.FileStorage, error)
	GetFile(ctx context.Context, spaceID, fileID string) (body []byte, contentType, name string, err error)
	GetFileObject(ctx context.Context, spaceID string, fileID string) (*minio.Object, error)
	Upload(ctx context.Context, spaceID, applicantID string, file []byte, fileName string, fileType dbmodels.FileType, contentType string) error
	UploadObject(ctx context.Context, fileInfo dbmodels.UploadFileInfo, reader io.Reader, fileSize int) (fileID string, err error)
	DeleteFile(ctx context.Context, spaceID, fileID string) error
	DeleteFileByType(ctx context.Context, spaceID, applicantID string, fileType dbmodels.FileType) error
	MakeSpaceBucket(ctx context.Context, spaceID string) error
}

var Instance Provider

func NewHandler() {
	Instance = &impl{
		s3client:       s3client.Client,
		filesDBStorage: filesdbstorage.NewInstance(db.DB),
	}
}

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

func (i impl) GetFileByType(ctx context.Context, spaceID, applicantID string, fileType dbmodels.FileType) ([]byte, *dbmodels.FileStorage, error) {
	file, err := i.filesDBStorage.GetFileIDByType(applicantID, fileType)
	if err != nil {
		return nil, nil, err
	}
	if file == nil {
		return nil, nil, nil
	}
	fileBody, _, _, err := i.GetFile(ctx, spaceID, file.ID)
	return fileBody, file, err
}

func (i impl) GetFile(ctx context.Context, spaceID string, fileID string) (body []byte, contentType, name string, err error) {
	if fileID == "" {
		return nil, "", "", nil
	}
	rec, err := i.filesDBStorage.GetByID(fileID)
	if err != nil {
		return nil, "", "", err
	}

	s3file, err := i.s3client.GetObject(ctx, i.getSpaceBucketName(spaceID), fileID, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", "", errors.Wrap(err, "ошибка получения файла из S3")
	}
	defer s3file.Close()
	body, err = io.ReadAll(s3file)
	if err != nil {
		return nil, "", "", errors.Wrap(err, "ошибка чтения файла из S3")
	}
	return body, rec.ContentType, rec.Name, nil
}

func (i impl) GetFileObject(ctx context.Context, spaceID string, fileID string) (*minio.Object, error) {
	s3file, err := i.s3client.GetObject(ctx, i.getSpaceBucketName(spaceID), fileID, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "ошибка получения файла из S3")
	}
	return s3file, nil
}

func (i impl) Upload(ctx context.Context, spaceID, applicantID string, file []byte, fileName string, fileType dbmodels.FileType, contentType string) error {
	fileInfo := dbmodels.UploadFileInfo{
		SpaceID:        spaceID,
		ApplicantID:    applicantID,
		FileName:       fileName,
		FileType:       fileType,
		ContentType:    contentType,
		IsUniqueByName: false,
	}
	_, err := i.UploadObject(ctx, fileInfo, bytes.NewReader(file), len(file))
	return err
}

func (i impl) UploadObject(ctx context.Context, fileInfo dbmodels.UploadFileInfo, reader io.Reader, fileSize int) (fileID string, err error) {
	logger := log.WithFields(log.Fields{
		"space_id":     fileInfo.SpaceID,
		"applicant_id": fileInfo.ApplicantID,
		"file_name":    fileInfo.FileName,
		"file_type":    fileInfo.FileType,
	})
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		fileDB := filesdbstorage.NewInstance(tx)
		removedID := ""
		if fileInfo.IsUniqueByName {
			removedID, err = i.deletingIrrelevantFileByName(fileDB, fileInfo.SpaceID, fileInfo.ApplicantID, fileInfo.FileType, fileInfo.FileName)
			if err != nil {
				logger.
					WithError(err).
					Error("ошибка удаления информации о существующем файле")
				removedID = ""
			}
		} else {
			removedID, err = i.deletingIrrelevantFile(fileDB, fileInfo.SpaceID, fileInfo.ApplicantID, fileInfo.FileType)
			if err != nil {
				logger.
					WithError(err).
					Error("ошибка удаления информации о существующем файле")
				removedID = ""
			}
		}
		rec := dbmodels.FileStorage{
			BaseSpaceModel: dbmodels.BaseSpaceModel{
				SpaceID: fileInfo.SpaceID,
			},
			Name:        fileInfo.FileName,
			ApplicantID: fileInfo.ApplicantID,
			Type:        fileInfo.FileType,
			ContentType: fileInfo.ContentType,
		}
		fileID, err = fileDB.SaveFile(rec)
		if err != nil {
			return errors.Wrap(err, "ошибка сохранения информации о файле")
		}
		bucketName := i.getSpaceBucketName(fileInfo.SpaceID)
		err = i.uploadFile(ctx, bucketName, fileID, reader, fileSize)
		if err != nil {
			return errors.Wrap(err, "ошибка сохранения файла в S3")
		}
		if removedID != "" {
			err = i.s3client.RemoveObject(ctx, bucketName, removedID, minio.RemoveObjectOptions{})
			if err != nil {
				logger.
					WithError(err).
					WithField("file_id", removedID).
					Error("ошибка удаления существующего файла")
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return fileID, nil
}

func (i impl) DeleteFile(ctx context.Context, spaceID, fileID string) error {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		fileDB := filesdbstorage.NewInstance(tx)
		ok, err := fileDB.DeleteFile(fileID, spaceID)
		if err != nil {
			return errors.Wrap(err, "ошибка удаления информации о файле")
		}
		if !ok {
			return nil
		}
		bucketName := i.getSpaceBucketName(spaceID)
		err = i.s3client.RemoveObject(ctx, bucketName, fileID, minio.RemoveObjectOptions{})
		if err != nil {
			return errors.Wrap(err, "ошибка удаления файла")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (i impl) DeleteFileByType(ctx context.Context, spaceID, applicantID string, fileType dbmodels.FileType) error {
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		fileDB := filesdbstorage.NewInstance(tx)
		file, err := fileDB.GetFileIDByType(applicantID, fileType)
		if err != nil {
			return errors.Wrap(err, "ошибка получения информации о файле")
		}

		if file == nil {
			return nil
		}
		ok, err := fileDB.DeleteFile(file.ID, spaceID)
		if err != nil {
			return errors.Wrap(err, "ошибка удаления информации о файле")
		}
		if !ok {
			return nil
		}
		bucketName := i.getSpaceBucketName(spaceID)
		err = i.s3client.RemoveObject(ctx, bucketName, file.ID, minio.RemoveObjectOptions{})
		if err != nil {
			return errors.Wrap(err, "ошибка удаления файла из S3")
		}
		return nil
	})
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

func (i impl) uploadFile(ctx context.Context, bucketName, fileID string, fileReader io.Reader, fileSize int) error {
	_, err := i.s3client.PutObject(ctx, bucketName, fileID, fileReader, int64(fileSize), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return err
	}
	return nil
}

func (i impl) getSpaceBucketName(spaceID string) string {
	return fmt.Sprintf("%s-%s", config.Conf.S3.BucketNamePrefix, spaceID)
}

func (i impl) deletingIrrelevantFile(fileDB filesdbstorage.Provider, spaceID, applicantID string, fileType dbmodels.FileType) (removedID string, err error) {
	if fileType == dbmodels.ApplicantDoc {
		return "", nil
	}
	file, err := fileDB.GetFileIDByType(applicantID, fileType)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения информации о существующем файле")
	}
	if file == nil {
		return "", nil
	}
	_, err = fileDB.DeleteFile(file.ID, spaceID)
	if err != nil {
		return "", errors.Wrap(err, "ошибка удаления информации о существующем файле")
	}

	return file.ID, nil
}

func (i impl) deletingIrrelevantFileByName(fileDB filesdbstorage.Provider, spaceID, applicantID string, fileType dbmodels.FileType, fileName string) (removedID string, err error) {
	if fileType == dbmodels.ApplicantDoc {
		return "", nil
	}
	list, err := fileDB.GetFileListByType(applicantID, fileType)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения информации о существующем файле")
	}
	if len(list) == 0 {
		return "", nil
	}
	for _, file := range list {
		if file.Name == fileName {
			_, err = fileDB.DeleteFile(file.ID, spaceID)
			if err != nil {
				return "", errors.Wrap(err, "ошибка удаления информации о существующем файле")
			}

			return file.ID, nil
		}
	}

	return "", nil
}
