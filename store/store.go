package store

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
	"path/filepath"
)

type StoreProvider interface {
	Store(path, filename string, body io.Reader) (string, error)
	Delete(fullPath string) error
}

type fsStore struct{}
type s3Store struct {
	AWSSession *session.Session
	S3Bucket   string
}

func NewFSStore() StoreProvider {
	return &fsStore{}
}

func NewS3Store(sess *session.Session, bucketName string) StoreProvider {
	return &s3Store{
		AWSSession: sess,
		S3Bucket:   bucketName,
	}
}

func (fss *fsStore) Store(path, filename string, body io.Reader) (string, error) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(path, filename)
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, body)
	if err != nil {
		return "", err
	}

	return fullPath, nil
}

func (fss *fsStore) Delete(fullPath string) error {
	return os.Remove(fullPath)
}

func (s3s *s3Store) Store(path, filename string, body io.Reader) (string, error) {
	var b bytes.Buffer
	_, err := b.ReadFrom(body)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(path, filename)

	uploader := s3manager.NewUploader(s3s.AWSSession)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3s.S3Bucket),
		Key:    aws.String(fullPath),
		Body:   bytes.NewReader(b.Bytes()),
	})

	if err != nil {
		return "", err
	}

	return fullPath, nil
}

func (s3s *s3Store) Delete(fullPath string) error {
	batcher := s3manager.NewBatchDelete(s3s.AWSSession)
	objects := []s3manager.BatchDeleteObject{
		{
			Object: &s3.DeleteObjectInput{
				Key:    aws.String(fullPath),
				Bucket: aws.String(s3s.S3Bucket),
			},
		},
	}

	if err := batcher.Delete(aws.BackgroundContext(), &s3manager.DeleteObjectsIterator{
		Objects: objects,
	}); err != nil {
		return err
	}
	return nil
}
