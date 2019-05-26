package store

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
	FileSystemStore(path, filename string, body io.Reader) error
	S3Store(path, filename string, body io.Reader) error
}

type store struct {
	FSRootPath string
	AWSSession *session.Session
	S3Bucket   string
}

func NewStore(path string, s *session.Session, bucket string) Storage {
	return &store{
		FSRootPath: path,
		AWSSession: s,
		S3Bucket:   bucket,
	}
}

func (s *store) mkImagePath(inputPath string) (string, error) {
	finalPath := filepath.Join(s.FSRootPath, inputPath)
	err := os.MkdirAll(finalPath, 0755)
	if err != nil {
		return "", err
	}

	return finalPath, nil
}

func (s *store) FileSystemStore(dir, filename string, body io.Reader) error {
	fmt.Println("Received filePath: ", dir)
	imagePath, err := s.mkImagePath(dir)
	if err != nil {
		return err
	}

	dst, err := os.Create(filepath.Join(imagePath, filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, body)
	if err != nil {
		return err
	}

	return nil
}

func (s *store) S3Store(path, filename string, body io.Reader) error {
	var b bytes.Buffer
	_, err := b.ReadFrom(body)
	if err != nil {
		return err
	}

	fullPath := filepath.Join(path, filename)

	uploader := s3manager.NewUploader(s.AWSSession)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.S3Bucket),
		Key:    aws.String(fullPath),
		Body:   bytes.NewReader(b.Bytes()),
	})

	if err != nil {
		return err
	}

	return err
}
