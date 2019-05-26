package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
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
	gorm.Model
	ExternalType string `gorm:"external_type,not_null"`
	ExternalID   uint   `gorm:"external_id, not_null"`
	Filename     string `gorm:"filename, not_null"`
	Location     string `gorm:"location, not_null"`
}

func (i *Image) Path() string {
	temp := url.URL{
		Path: "/" + i.RelativePath(),
	}
	return temp.String()
}

func (i *Image) RelativePath() string {
	externalID := fmt.Sprintf("%v", i.ExternalID)
	return filepath.ToSlash(filepath.Join("images", i.ExternalType, externalID, i.Filename))
}

// ImageService is the definition of image service operation
type ImageService interface {
	ImageDB
}

type ImageDB interface {
	Create(externalType string, externalID uint, r io.Reader, filename string) error
	ByExternalTypeAndID(ExternalType string, ExternalID uint) ([]Image, error)
	Delete(i *Image) error
}

type imageService struct {
	StorageType string
	Storage     store.Storage
	ImageDB
}

type imageValidator struct {
	ImageDB
}

type imageGorm struct {
	db *gorm.DB
}

func NewImageService(storage store.Storage, db *gorm.DB, storageType string) ImageService {
	return &imageService{
		ImageDB: &imageValidator{
			ImageDB: &imageGorm{
				db: db,
			},
		},
		Storage:     storage,
		StorageType: storageType,
	}
}

func (im *imageService) imagePath(externalType string, externalID uint) string {
	return filepath.Join("images", externalType, fmt.Sprintf("%v", externalID))
}

func (im *imageService) Create(externalType string, externalID uint, r io.Reader, filename string) error {
	imagePath := im.imagePath(externalType, externalID)

	resultPath, err := im.Storage.Store(im.StorageType, imagePath, filename, r)

	if err != nil {
		return err
	}
	// Store in the DB
	_ = resultPath
	return nil
}

func (im *imageService) ByExternalTypeAndID(ExternalType string, ExternalID uint) ([]Image, error) {
	path := im.imagePath(ExternalType, ExternalID)
	strings, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, err
	}

	// make slice of type Image
	ret := make([]Image, len(strings))
	for i, imgPath := range strings {
		ret[i] = Image{
			ExternalID:   ExternalID,
			ExternalType: ExternalType,
			Filename:     filepath.Base(imgPath),
		}
	}
	return ret, nil
}

func (im *imageService) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}

func (ig *imageGorm) Create(externalType string, externalID uint, r io.Reader, filename string) error {
	return errors.New("Not implemented")
}

func (ig *imageGorm) ByExternalTypeAndID(ExternalType string, ExternalID uint) ([]Image, error) {
	return nil, nil
}

func (ig *imageGorm) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}
