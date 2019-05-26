package models

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jinzhu/gorm"
	"github.com/ruckuus/dojo1/store"
)

type Services struct {
	User        UserService
	Gallery     GalleryService
	Image       ImageService
	Property    PropertyService
	db          *gorm.DB
	AWSSession  *session.Session
	S3Bucket    string
	Store       store.Storage
	StorageType string
}

type ServicesConfig func(*Services) error

func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, cfg := range cfgs {
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func WithGorm(dialect, connectionInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(dialect, connectionInfo)
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}

func WithLogMode(mode bool) ServicesConfig {
	return func(s *Services) error {
		s.db.LogMode(mode)
		return nil
	}
}

func WithUser(pepper, hmacKey string) ServicesConfig {
	return func(s *Services) error {
		s.User = NewUserService(s.db, pepper, hmacKey)
		return nil
	}
}

func WithGallery() ServicesConfig {
	return func(s *Services) error {
		s.Gallery = NewGalleryService(s.db)
		return nil
	}
}

func WithAWSSession(sess *session.Session) ServicesConfig {
	return func(s *Services) error {
		s.AWSSession = sess
		return nil
	}
}

func WithS3Bucket(bucketName string) ServicesConfig {
	return func(s *Services) error {
		s.S3Bucket = bucketName
		return nil
	}
}

func WithStore(rootPath string) ServicesConfig {
	return func(s *Services) error {
		storage := store.NewStore(rootPath, s.AWSSession, s.S3Bucket)
		s.Store = storage
		return nil
	}
}

func WithStorageType(storeType string) ServicesConfig {
	return func(s *Services) error {
		s.StorageType = storeType
		return nil
	}
}

func WithImage() ServicesConfig {
	return func(s *Services) error {
		s.Image = NewImageService(s.Store, s.db, s.StorageType)
		return nil
	}
}

func WithProperty() ServicesConfig {
	return func(s *Services) error {
		s.Property = NewPropertyService(s.db)
		return nil
	}
}

// I will keep this commented, for future reference
//func NewServices(dialect, connectionInfo string) (*Services, error) {
//
//	db, err := gorm.Open(dialect, connectionInfo)
//	if err != nil {
//		return nil, err
//	}
//
//	db.LogMode(true)
//
//	return &Services{
//		User:    NewUserService(db),
//		Gallery: NewGalleryService(db),
//		Image:   NewImageService(),
//		db:      db,
//	}, nil
//}

func (s *Services) Close() error {
	return s.db.Close()
}

func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&User{}, &Gallery{}, &pwReset{}, &Property{}).Error
}

func (s *Services) DestructiveReset() error {
	err := s.db.DropTableIfExists(&User{}, &Gallery{}, &pwReset{}, &Property{}).Error
	if err != nil {
		return err
	}
	return s.AutoMigrate()
}
