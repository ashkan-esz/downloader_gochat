package cloudStorage

import (
	"context"
	"downloader_gochat/configs"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type IS3Storage interface {
	UploadFile(bucketName string, fileName string, file multipart.File) (*s3.PutObjectOutput, error)
	UploadLargeFile(bucketName string, fileName string, file multipart.File) (*manager.UploadOutput, error)
	RemoveFile(bucketName string, fileName string) error
}

type S3Storage struct {
	client  *s3.Client
	Configs *configs.ConfigStruct
}

func StartS3StorageService() *S3Storage {
	config := configs.GetConfigs()
	options := aws.Config{
		Region:       "default",
		Credentials:  credentials.NewStaticCredentialsProvider(config.CloudStorageAccessKey, config.CloudStorageSecretAccessKey, ""),
		BaseEndpoint: aws.String(config.CloudStorageEndpoint),
	}

	return &S3Storage{
		client:  s3.NewFromConfig(options),
		Configs: &config,
	}
}

const (
	MediaFileBucketName               = "media-file"
	DownloadAppBucketName             = "download-app"
	ProfileImageBucketName            = "profile-image"
	DownloadTrailerBucketName         = "download-trailer"
	PosterBucketName                  = "poster"
	DownloadSubtitleBucketName        = "download-subtitle"
	CastBucketName                    = "cast"
	ServerStaticFilesBucketName       = "serverstatic"
	partMiBs                    int64 = 5
	publicReadACL                     = "public-read"
)

//------------------------------------------
//------------------------------------------

func (s *S3Storage) UploadFile(bucketName string, fileName string, file multipart.File) (*s3.PutObjectOutput, error) {
	if s.Configs.CloudStorageBucketNamePrefix != "" {
		bucketName = s.Configs.CloudStorageBucketNamePrefix + bucketName
	}

	result, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   file,
		ACL:    publicReadACL,
	})

	return result, err
}

func (s *S3Storage) UploadLargeFile(bucketName string, fileName string, file multipart.File) (*manager.UploadOutput, error) {
	if s.Configs.CloudStorageBucketNamePrefix != "" {
		bucketName = s.Configs.CloudStorageBucketNamePrefix + bucketName
	}

	uploader := manager.NewUploader(s.client, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})
	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   file,
		ACL:    publicReadACL,
	})

	return result, err
}

func (s *S3Storage) RemoveFile(bucketName string, fileName string) error {
	if s.Configs.CloudStorageBucketNamePrefix != "" {
		bucketName = s.Configs.CloudStorageBucketNamePrefix + bucketName
	}

	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})

	return err
}
