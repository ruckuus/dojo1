package models

import (
	"fmt"
	"github.com/ruckuus/dojo1/store"
	"io"
	"net/url"
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
	temp := url.URL{
		Path: "/" + i.RelativePath(),
	}
	return temp.String()
}

func (i *Image) RelativePath() string {
	galleryID := fmt.Sprintf("%v", i.GalleryID)
	return filepath.ToSlash(filepath.Join("images", "galleries", galleryID, i.Filename))
}

// ImageService is the definition of image service operation
type ImageService interface {
	Create(galleryID uint, r io.Reader, filename string) error
	ByGalleryID(galleryID uint) ([]Image, error)
	Delete(i *Image) error
}

type imageService struct {
	Storage store.Storage
}

func NewImageService(storage store.Storage) ImageService {
	return &imageService{
		storage,
	}
}

func (im *imageService) imagePath(galleryID uint) string {
	return filepath.Join("images", "galleries", fmt.Sprintf("%v", galleryID))
}

func (im *imageService) Create(galleryID uint, r io.Reader, filename string) error {
	galleryPath := im.imagePath(galleryID)

	err := im.Storage.FileSystemStore(galleryPath, filename, r)

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

func (im *imageService) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}
