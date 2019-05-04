package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	ErrImageInvalidPath modelError = "Error"
)

type ImageService interface {
	Create(galleryID uint, r io.Reader, filename string) error
}

type imageService struct{}

func NewImageService() ImageService {
	return &imageService{}
}

func (im *imageService) mkImagePath(galleryID uint) (string, error) {
	galleryPath := filepath.Join("images", "galleries", fmt.Sprintf("%v", galleryID))
	err := os.MkdirAll(galleryPath, 0755)
	if err != nil {
		return "", err
	}

	return galleryPath, nil
}

func (im *imageService) Create(galleryID uint, r io.Reader, filename string) error {
	path, err := im.mkImagePath(galleryID)
	if err != nil {
		return err
	}

	dst, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}
