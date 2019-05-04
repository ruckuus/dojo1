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

// Image is used to represent images stored in a Gallery
// It is not stored in DB, the referenced data is stored
// on disk.
type Image struct {
	GalleryID uint
	Filename  string
}

func (i *Image) Path() string {
	return "/" + i.RelativePath()
}

func (i *Image) RelativePath() string {
	galleryID := fmt.Sprintf("%v", i.GalleryID)
	return filepath.ToSlash(filepath.Join("images", "galleries", galleryID, i.Filename))
}

// ImageService is the definition of image service operation
type ImageService interface {
	Create(galleryID uint, r io.Reader, filename string) error
	ByGalleryID(galleryID uint) ([]Image, error)
}

type imageService struct{}

func NewImageService() ImageService {
	return &imageService{}
}

func (im *imageService) imagePath(galleryID uint) string {
	return filepath.Join("images", "galleries", fmt.Sprintf("%v", galleryID))
}

func (im *imageService) mkImagePath(galleryID uint) (string, error) {
	galleryPath := im.imagePath(galleryID)
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

func (im *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	path := im.imagePath(galleryID)
	strings, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, err
	}

	// make slice of type Image
	ret := make([]Image, len(strings))
	for i, imgPath := range strings {
		ret[i] = Image{
			GalleryID: galleryID,
			Filename:  filepath.Base(imgPath),
		}
	}
	return ret, nil
}
