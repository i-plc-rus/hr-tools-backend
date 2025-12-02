package videonormalize

import (
	"context"
	"hr-tools-backend/config"
	filestorage "hr-tools-backend/lib/file-storage"
	dbmodels "hr-tools-backend/models/db"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func Run(ctx context.Context, fileInfo dbmodels.UploadFileInfo, fileID string) (string, error) {
	logger := log.WithFields(log.Fields{
		"space_id":     fileInfo.SpaceID,
		"applicant_id": fileInfo.ApplicantID,
		"file_id":      fileID,
		"file_name":    fileInfo.FileName,
	})

	// Получаем файл из S3
	fileReader, err := filestorage.Instance.GetFileObject(ctx, fileInfo.SpaceID, fileID)
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения файла из S3")
	}
	defer fileReader.Close()

	// Создаем временные файлы
	tmpDir, err := os.MkdirTemp("", "video-normalize-*")
	if err != nil {
		return "", errors.Wrap(err, "ошибка создания временной директории")
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "input.webm")
	outputFile := filepath.Join(tmpDir, "output.webm")

	// Сохраняем исходный файл во временный файл
	inputFileHandle, err := os.Create(inputFile)
	if err != nil {
		return "", errors.Wrap(err, "ошибка создания временного файла для входного видео")
	}
	defer inputFileHandle.Close()

	_, err = io.Copy(inputFileHandle, fileReader)
	if err != nil {
		return "", errors.Wrap(err, "ошибка копирования файла во временный файл")
	}
	inputFileHandle.Close()

	// Нормализуем видео с помощью ffmpeg (1280x720, формат webm)
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inputFile,
		"-c:v", "libvpx-vp9", // кодек для webm
		"-b:v", config.Conf.Survey.VideoNormalizeBitrate, // битрейт видео - kbps
		"-vf", "scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2", // с сохранением пропорций
		"-c:a", "libopus", // аудио кодек для webm
		"-b:a", "128k", // битрейт аудио
		"-y", // перезаписать выходной файл если существует
		outputFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.WithError(err).
			WithField("ffmpeg_output", string(output)).
			Error("ошибка выполнения ffmpeg")
		return "", errors.Wrapf(err, "ошибка нормализации видео: %s", string(output))
	}

	// Открываем нормализованный файл для загрузки
	normalizedFileHandle, err := os.Open(outputFile)
	if err != nil {
		return "", errors.Wrap(err, "ошибка открытия нормализованного файла")
	}
	defer normalizedFileHandle.Close()

	// Получаем размер файла
	fileStat, err := normalizedFileHandle.Stat()
	if err != nil {
		return "", errors.Wrap(err, "ошибка получения информации о нормализованном файле")
	}
	fileSize := int(fileStat.Size())

	// Загружаем нормализованный файл в S3
	// Используем IsUniqueByName=true, чтобы заменить старый файл с тем же именем
	fileInfoUpload := fileInfo
	fileInfoUpload.ContentType = "video/webm" // webm формат
	fileInfoUpload.IsUniqueByName = true

	newFileID, err := filestorage.Instance.UploadObject(ctx, fileInfoUpload, normalizedFileHandle, fileSize)
	if err != nil {
		return "", errors.Wrap(err, "ошибка загрузки нормализованного файла в S3")
	}

	// Закрываем файл перед удалением
	normalizedFileHandle.Close()

	// Удаляем временные файлы сразу после загрузки в S3
	os.RemoveAll(tmpDir)

	logger.WithField("new_file_id", newFileID).Info("видео успешно нормализовано")
	return newFileID, nil
}
