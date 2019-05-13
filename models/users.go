package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/ruckuus/dojo1/hash"
	"github.com/ruckuus/dojo1/rand"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
	"time"
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
	InitiateReset(email string) (string, error)
	CompleteReset(token, newPassword string) (*User, error)
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
}

// userService
type userService struct {
	UserDB
	pepper    string
	pwResetDB pwResetDB
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
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
	pepper     string
}

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])

	return strings.Join(split, " ")
}

// Check, it must error during compilation
var _ UserDB = &userGorm{}
var _ UserService = &userService{}

var (
	//ErrNotFound is returned when a resource cannot be found
	ErrNotFound modelError = "models: resource not found"

	//ErrIDInvalid is returned when provided ID is invalid
	ErrIDInvalid modelError = "models: ID provided is invalid"

	//ErrInvalidPassword is returned when provided password is invalid
	ErrPasswordInvalid modelError = "models: incorrect password provided"

	//ErrEmailRequired is returned with email field is not present
	ErrEmailRequired modelError = "models: email is required"

	//ErrEmailInvalid
	ErrEmailInvalid modelError = "models: email address is not valid"

	//ErrEmailTaken
	ErrEmailTaken modelError = "models: email is already taken"

	// ErrPasswordTooShort
	ErrPasswordTooShort modelError = "models: password must be at least 8 characters"

	// ErrPasswordRequired
	ErrPasswordRequired modelError = "models: password is required"

	//ErrRememberRequired
	ErrRememberRequired modelError = "models: remember token is required"

	// ErrRememberTooShort
	ErrRememberTooShort modelError = "models: remember token must be at least 32 bytes"

	// ErrTokenInvalid
	ErrTokenInvalid modelError = "models: token provided is not valid"
)

// NewUserService Create new UserService instance
func NewUserService(db *gorm.DB, pepper, hmacKey string) UserService {
	ug := &userGorm{db}

	hmac := hash.NewHMAC(hmacKey)
	uv := newUserValidator(ug, hmac, pepper)

	return &userService{
		UserDB:    uv,
		pepper:    pepper,
		pwResetDB: newPwResetValidator(&pwResetGorm{db}, hmac),
	}
}

func newUserValidator(udb UserDB, hmac hash.HMAC, pepper string) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hmac,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
		pepper:     pepper,
	}
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
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

// Authenticate will return error when password provided mismatch
func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash),
		[]byte(password+us.pepper))
	switch err {
	case nil:
		return foundUser, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrPasswordInvalid
	default:
		return nil, err
	}
}

func (us *userService) InitiateReset(email string) (string, error) {
	user, err := us.ByEmail(email)
	if err != nil {
		return "", nil
	}

	pwr := pwReset{
		UserID: user.ID,
	}

	if err := us.pwResetDB.Create(&pwr); err != nil {
		return "", nil
	}

	return pwr.Token, nil
}

func (us *userService) CompleteReset(token, newPassword string) (*User, error) {
	pwr, err := us.pwResetDB.ByToken(token)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrTokenInvalid
		}
		return nil, err
	}
	// check token validity

	if time.Now().Sub(pwr.CreatedAt) < (12 * time.Hour) {
		return nil, ErrTokenInvalid
	}

	user, err := us.ByID(pwr.UserID)
	if err != nil {
		return nil, err
	}
	user.Password = newPassword

	err = us.Update(user)
	return user, nil
}

// Create will perform user validation before calling user creation function
func (uv *userValidator) Create(user *User) error {

	err := runUserValidationFunctions(user,
		uv.passwordMinLength,
		uv.passwordIsRequired,
		uv.requireEmail,
		uv.emailFormat,
		uv.normalizeEmail,
		uv.emailIsAvail,
		uv.bcryptPassword,
		uv.setRememberIfUnset,
		uv.rememberMinBytes,
		uv.hmacRemember)
	if err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

// Update will update the provided user with all the data in the provided user object
func (uv *userValidator) Update(user *User) error {
	err := runUserValidationFunctions(user,
		uv.passwordMinLength,
		uv.passwordHashIsRequired,
		uv.requireEmail,
		uv.emailFormat,
		uv.normalizeEmail,
		uv.bcryptPassword,
		uv.hmacRemember,
		uv.rememberHashIsRequired)

	if err != nil {
		return err
	}

	return uv.UserDB.Update(user)
}

// ByEmail validate and normalize email address before passing it to UserDB
func (uv *userValidator) ByEmail(email string) (*User, error) {
	var user User
	user.Email = email

	err := runUserValidationFunctions(&user,
		uv.requireEmail,
		uv.normalizeEmail)

	if err != nil {
		return nil, err
	}

	return uv.UserDB.ByEmail(user.Email)
}

// ByRemember find users record from database by remember token
func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}

	if err := runUserValidationFunctions(&user, uv.hmacRemember); err != nil {
		return nil, err
	}
	return uv.UserDB.ByRemember(user.RememberHash)
}

func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id

	err := runUserValidationFunctions(&user, uv.idGreaterThan(0))

	if err != nil {
		return err
	}

	return uv.UserDB.Delete(id)
}

// bcryptPassword normalize password input before used by DB layer
func (uv *userValidator) bcryptPassword(user *User) error {
	// If password does not change, we do not need to hash it
	if user.Password == "" {
		return nil
	}

	passwordBytes := user.Password + uv.pepper
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(passwordBytes), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

// hmacRemember hmac remember token
func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}

	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}

	token, err := rand.RememberToken()
	if err != nil {
		return err
	}

	user.Remember = token
	return nil
}

func (uv *userValidator) idGreaterThan(n uint) userValidationFn {
	return userValidationFn(func(user *User) error {
		if user.ID <= n {
			return ErrIDInvalid
		}
		return nil
	})
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.TrimSpace(user.Email)
	user.Email = strings.ToLower(user.Email)
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}

	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}

	return nil
}

func (uv *userValidator) emailIsAvail(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		return nil
	}

	if err != nil {
		return err
	}

	if user.ID != existing.ID {
		return ErrEmailTaken
	}

	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}

	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}

	return nil
}

func (uv *userValidator) passwordIsRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) passwordHashIsRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}

	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}

	if n < 32 {
		return ErrRememberTooShort
	}
	return nil
}

func (uv *userValidator) rememberHashIsRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberRequired
	}
	return nil
}

// userValidationFn accepts pointer to user, it returns error
type userValidationFn func(user *User) error

// runUserValidationFunctions takes user pointer and a list of
// userValidationFn to be executed, error will be thrown or nil
// for successful validation.
func runUserValidationFunctions(user *User, fns ...userValidationFn) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}
