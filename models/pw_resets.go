package models

import (
	"github.com/jinzhu/gorm"
	"github.com/ruckuus/dojo1/hash"
	"github.com/ruckuus/dojo1/rand"
)

type pwReset struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	Token     string `gorm:"-"`
	TokenHash string `gorm:"not null; unique_index"`
}

type pwResetDB interface {
	ByToken(token string) (*pwReset, error)
	Create(pwr *pwReset) error
	Delete(id uint) error
}

type pwResetGorm struct {
	db *gorm.DB
}

type pwResetValidator struct {
	pwResetDB
	hmac hash.HMAC
}

func newPwResetValidator(db pwResetDB, hmac hash.HMAC) *pwResetValidator {
	return &pwResetValidator{
		pwResetDB: db,
		hmac:      hmac,
	}
}

var _ pwResetDB = &pwResetGorm{}
var _ pwResetDB = &pwResetValidator{}

func (pwrg *pwResetGorm) ByToken(tokenHash string) (*pwReset, error) {
	var pwr pwReset
	err := first(pwrg.db.Where("token_hash = ?", tokenHash), &pwr)
	if err != nil {
		return nil, err
	}
	return &pwr, nil
}

func (pwrg *pwResetGorm) Create(pwr *pwReset) error {
	return pwrg.db.Create(pwr).Error
}

func (pwrg *pwResetGorm) Delete(id uint) error {
	pwr := pwReset{Model: gorm.Model{ID: id}}
	return pwrg.db.Delete(&pwr).Error
}

// Validator
func (pwrv *pwResetValidator) ByToken(token string) (*pwReset, error) {
	pwr := pwReset{
		Token: token,
	}

	err := runPwResetValidationFns(&pwr, pwrv.hmacToken)
	if err != nil {
		return nil, err
	}
	return pwrv.pwResetDB.ByToken(pwr.TokenHash)
}

func (pwrv *pwResetValidator) Create(pwr *pwReset) error {
	err := runPwResetValidationFns(pwr, pwrv.requireUserID, pwrv.setTokenIfUnset, pwrv.hmacToken)
	if err != nil {
		return err
	}
	return pwrv.pwResetDB.Create(pwr)
}

func (pwrv *pwResetValidator) Delete(id uint) error {
	if id <= 0 {
		return ErrIDInvalid
	}

	return pwrv.pwResetDB.Delete(id)
}

// validation funcs
type pwResetValFn func(*pwReset) error

// validation function implementation
func (pwvr *pwResetValidator) requireUserID(pwr *pwReset) error {
	if pwr.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (pwvr *pwResetValidator) setTokenIfUnset(pwr *pwReset) error {
	if pwr.Token != "" {
		return nil
	}

	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	pwr.Token = token
	return nil
}

func (pwvr *pwResetValidator) hmacToken(pwr *pwReset) error {
	if pwr.Token == "" {
		return ErrTokenInvalid
	}

	pwr.TokenHash = pwvr.hmac.Hash(pwr.Token)
	return nil
}

// validation runner
func runPwResetValidationFns(pwr *pwReset, fns ...pwResetValFn) error {
	for _, f := range fns {
		if err := f(pwr); err != nil {
			return err
		}
	}
	return nil
}
