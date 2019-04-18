package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null; unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
}

type UserService struct {
	db *gorm.DB
}

var UserPasswordPepper = "HALUSINOGEN2019$$"

var (
	//ErrNotFound is returned when a resource cannot be found
	ErrNotFound  = errors.New("models: resource not found")
	ErrInvalidID = errors.New("models: ID provided is invalid")
)

// NewUserService Create new UserService instance
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	// set Log
	db.LogMode(true)

	return &UserService{
		db: db,
	}, nil
}

// Close the UserService DB Connection
func (us *UserService) Close() error {
	return us.db.Close()
}

// Create new record
func (us *UserService) Create(user *User) error {
	passwordRaw := []byte(user.Password + UserPasswordPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordRaw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = ""
	user.PasswordHash = string(hashedBytes)
	return us.db.Create(user).Error
}

// first will query using the provided gorm.DB and it will
// get the first item returned and place it in dst. If nothing
// is found in the query, it will return ErrNotFound
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

// ByID find a record by ID
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ByEmail find a record by email
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update will update the provided user with all the data in the provided user object
func (us *UserService) Update(user *User) error {
	return us.db.Save(&user).Error
}

// Delete will delete the record for the user
func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}

	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(&user).Error
}

// AutoMigrate will run migration
func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}

// DestructiveReset only used for development
func (us *UserService) DestructiveReset() error {
	err := us.db.DropTableIfExists(&User{}).Error
	if err != nil {
		return err
	}

	return us.AutoMigrate()
}
