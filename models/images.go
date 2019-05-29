package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
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
	Create(image *Image, r io.Reader) error
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

func (im *imageService) Create(image *Image, r io.Reader) error {
	imagePath := im.imagePath(image.ExternalType, image.ExternalID)

	resultPath, err := im.Storage.Store(im.StorageType, imagePath, image.Filename, r)

	if err != nil {
		return err
	}

	image.Location = resultPath

	err = im.ImageDB.Create(image, r)

	return nil
}

func (im *imageService) Delete(i *Image) error {
	err := os.Remove(i.RelativePath())
	if err != nil {
		return err
	}

	fmt.Println("Relative Path: ", i.RelativePath())
	i.Location = i.RelativePath()

	fmt.Println("Object: ", i)

	return im.ImageDB.Delete(i)
}

func (ig *imageGorm) Create(image *Image, r io.Reader) error {
	return ig.db.Create(image).Error
}

func (ig *imageGorm) ByExternalTypeAndID(externalType string, externalID uint) ([]Image, error) {
	var images []Image
	db := ig.db.Where("external_type = ? AND external_id = ?", externalType, externalID)
	err := db.Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (ig *imageGorm) Delete(i *Image) error {
	var image Image
	db := ig.db.Where(i)
	err := db.Find(&image).Error
	if err != nil {
		return err
	}

	return db.Delete(image).Error
}
