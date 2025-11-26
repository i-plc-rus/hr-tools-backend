package normalizevideo

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	s3client "hr-tools-backend/s3"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func Run(info minio.UploadInfo) (minio.UploadInfo, error) {
	ctx := context.Background()
	logger := log.WithFields(log.Fields{
		"bucket": info.Bucket,
		"key":    info.Key,
	})

	// Скачиваем исходный файл из S3
	s3file, err := s3client.Client.GetObject(ctx, info.Bucket, info.Key, minio.GetObjectOptions{})
	if err != nil {
		return info, errors.Wrap(err, "ошибка получения файла из S3")
	}
	defer s3file.Close()

	// Создаем временную директорию для файлов
	tempDir, err := os.MkdirTemp("", "video_normalize_*")
	if err != nil {
		return info, errors.Wrap(err, "ошибка создания временной директории")
	}
	defer os.RemoveAll(tempDir) // удаляем всю директорию со всеми файлами

	inputFile := filepath.Join(tempDir, "input_video")
	outputFile := filepath.Join(tempDir, "output_video.webm")

	// Сохраняем исходный файл во временный файл
	inputFileHandle, err := os.Create(inputFile)
	if err != nil {
		return info, errors.Wrap(err, "ошибка создания временного файла")
	}

	_, err = io.Copy(inputFileHandle, s3file)
	inputFileHandle.Close()
	if err != nil {
		return info, errors.Wrap(err, "ошибка сохранения исходного файла")
	}

	// Нормализуем видео через ffmpeg
	// webm, 1280x720, битрейт 2000k
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-vf", "scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2",
		"-c:v", "libvpx-vp9",
		"-b:v", "2000k",
		"-c:a", "libopus",
		"-y", // перезаписать выходной файл если существует
		outputFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.WithError(err).WithField("ffmpeg_output", string(output)).Error("ошибка выполнения ffmpeg")
		return info, errors.Wrapf(err, "ошибка нормализации видео: %s", string(output))
	}

	// Проверяем, что выходной файл создан
	outputStat, err := os.Stat(outputFile)
	if err != nil {
		return info, errors.Wrap(err, "выходной файл не создан")
	}

	// Формируем новое имя файла с меткой нормализации "_n"
	originalKey := info.Key
	ext := filepath.Ext(originalKey)
	baseName := strings.TrimSuffix(originalKey, ext)
	newKey := baseName + "_n.webm"

	// Открываем нормализованный файл для загрузки
	normalizedFile, err := os.Open(outputFile)
	if err != nil {
		return info, errors.Wrap(err, "ошибка открытия нормализованного файла")
	}
	defer normalizedFile.Close()

	// Загружаем нормализованный файл в S3 с новым именем
	uploadInfo, err := s3client.Client.PutObject(ctx, info.Bucket, newKey, normalizedFile, outputStat.Size(), minio.PutObjectOptions{
		ContentType: "video/webm",
	})
	if err != nil {
		return info, errors.Wrap(err, "ошибка загрузки нормализованного файла в S3")
	}

	// Удаляем исходный файл
	err = s3client.Client.RemoveObject(ctx, info.Bucket, originalKey, minio.RemoveObjectOptions{})
	if err != nil {
		logger.WithError(err).WithField("original_key", originalKey).Warn("ошибка удаления исходного файла")
		// Не возвращаем ошибку, так как нормализованный файл уже загружен
	}

	logger.WithFields(log.Fields{
		"original_key":    originalKey,
		"normalized_key":  newKey,
		"original_size":   info.Size,
		"normalized_size": uploadInfo.Size,
	}).Info("Видео успешно нормализовано")

	// Обновляем информацию о файле с новым ключом
	info.Key = newKey
	info.Size = uploadInfo.Size
	info.ETag = uploadInfo.ETag
	info.LastModified = uploadInfo.LastModified
	info.ChecksumCRC32 = uploadInfo.ChecksumCRC32
	info.ChecksumCRC32C = uploadInfo.ChecksumCRC32C
	info.ChecksumSHA1 = uploadInfo.ChecksumSHA1
	info.ChecksumSHA256 = uploadInfo.ChecksumSHA256

	return info, nil
}
