package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"github.com/ruckuus/dojo1/hash"
	"github.com/ruckuus/dojo1/rand"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null; unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null, unique_index"`
}

type UserService interface {
	Authenticate(email, password string) (*User, error)
	UserDB
}

// UserDB is used to interact with the users database.
//
// For pretty much all single user queries:
// If the user is found, we will return a nil error
// If the user is not found, we will return ErrNotFound
// If there is another error, we will return an error with
// more information about what went wrong. This may not be
// an error generated by the models package.
//
// For single user queries, any error but ErrNotFound should
// probably result in a 500 error until we make "public"
// facing errors.
type UserDB interface {
	// Find user by parameter
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	// Method for altering user
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	// Close DB connection
	Close() error

	// Migration
	AutoMigrate() error
	DestructiveReset() error
}

// userService
type userService struct {
	UserDB
}

// userGorm represents the database interaction layer
// and implements UserDB inteface fully
type userGorm struct {
	db *gorm.DB
}

// userValidator is our validation layer that validates
// and normalizes data before passing it on to the next
// UserDB in our interface chain.
type userValidator struct {
	UserDB
	hmac hash.HMAC
}

var userPasswordPepper = "HALUSINOGEN2019$$"
var userHMACSecretKey = "SuperSecret2019!$"

// Check, it must error during compilation
var _ UserDB = &userGorm{}
var _ UserService = &userService{}

var (
	//ErrNotFound is returned when a resource cannot be found
	ErrNotFound = errors.New("models: resource not found")

	//ErrInvalidID is returned when provided ID is invalid
	ErrInvalidID = errors.New("models: ID provided is invalid")

	//ErrInvalidPassword is returned when provided password is invalid
	ErrInvalidPassword = errors.New("models: incorrect password provided")
)

// NewUserService Create new UserService instance
func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}

	hmac := hash.NewHMAC(userHMACSecretKey)

	return &userService{
		UserDB: &userValidator{
			UserDB: ug,
			hmac:   hmac,
		},
	}, nil
}

// newUserGorm Create new userGorm instance
func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	// set Log
	db.LogMode(true)

	return &userGorm{
		db: db,
	}, nil
}

// Close the UserService DB Connection
func (ug *userGorm) Close() error {
	return ug.db.Close()
}

func preparePassword(password string) string {
	return password + userPasswordPepper
}

// Create new record
func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
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
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ByEmail find a record by email
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ByRemember find users record from database by remember token
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User

	db := ug.db.Where("remember_hash = ?", rememberHash)

	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update will update the provided user with all the data in the provided user object
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(&user).Error
}

// Delete will delete the record for the user
func (ug *userGorm) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}

	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

// AutoMigrate will run migration
func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}

// DestructiveReset only used for development
func (ug *userGorm) DestructiveReset() error {
	err := ug.db.DropTableIfExists(&User{}).Error
	if err != nil {
		return err
	}

	return ug.AutoMigrate()
}

// Authenticate will return error when password provided mismatch
func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(preparePassword(password)))
	switch err {
	case nil:
		return foundUser, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrInvalidPassword
	default:
		return nil, err
	}
}

func (uv *userValidator) Create(user *User) error {
	passwordRaw := []byte(preparePassword(user.Password))
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordRaw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = ""
	user.PasswordHash = string(hashedBytes)

	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
	}

	user.RememberHash = uv.hmac.Hash(user.Remember)
	return uv.UserDB.Create(user)
}

// Update will update the provided user with all the data in the provided user object
func (uv *userValidator) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = uv.hmac.Hash(user.Remember)
	}
	return uv.UserDB.Update(user)
}

// User Validation + Normalization
// ByRemember find users record from database by remember token
func (uv *userValidator) ByRemember(token string) (*User, error) {
	hashedToken := uv.hmac.Hash(token)
	return uv.UserDB.ByRemember(hashedToken)
}
