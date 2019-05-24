package models

import "github.com/jinzhu/gorm"

const (
	ErrPropertyNameRequired    modelError = "models: property name is required"
	ErrPropertyAddressRequired modelError = "models: property address is required"
	ErrPostalCodeRequired      modelError = "models: postal code is required"
)

type Property struct {
	gorm.Model
	UserID     uint   `gorm:"not_null;index"`
	Name       string `gorm:"not_null"`
	Address    string `gorm:"not_null"`
	PostalCode string `gorm:"not_null"`
}

// PropertyDB is the main interface,
// accessible from outside model package
type PropertyDB interface {
	ByID(id uint) (*Property, error)
	//ByUserID(id uint) (*Property, error)
	Create(property *Property) error
	//Update(property *Property) error
	//Delete(property *Property) error
}

// PropertyService has the same method as
// PropertyDB interface
type PropertyService interface {
	PropertyDB
}

// concrete type that is returned by New* method
type propertyService struct {
	PropertyDB
}

// propertyValidator is the concrete type that implements
// validation
type propertyValidator struct {
	PropertyDB
}

// propertyGorm implements DB operation, it also implements
// PropertyDB method
type propertyGorm struct {
	db *gorm.DB
}

// NewPropertyService return a service object to be used by
// external code
func NewPropertyService(db *gorm.DB) PropertyService {
	return &propertyService{
		PropertyDB: &propertyValidator{
			PropertyDB: &propertyGorm{
				db: db,
			},
		},
	}
}

// DB Implementation
func (pg *propertyGorm) ByID(id uint) (*Property, error) {
	var property Property
	db := pg.db.Where("id = ?", id)
	err := first(db, &property)
	if err != nil {
		return nil, err
	}
	return &property, nil
}

func (pg *propertyGorm) Create(p *Property) error {
	return pg.db.Create(p).Error
}

// Validator implementation
func (pv *propertyValidator) Create(p *Property) error {
	if err := runPropertyValFns(p,
		pv.userIDRequired,
		pv.propertyNameRequired,
		pv.propertyAddressRequired,
		pv.postalCodeRequired); err != nil {
		return err
	}
	return pv.PropertyDB.Create(p)
}

// Validation functions
func (pv *propertyValidator) userIDRequired(p *Property) error {
	if p.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (pv *propertyValidator) propertyNameRequired(p *Property) error {
	if p.Name == "" {
		return ErrPropertyNameRequired
	}
	return nil
}

func (pv *propertyValidator) propertyAddressRequired(p *Property) error {
	if p.Address == "" {
		return ErrPropertyAddressRequired
	}
	return nil
}

func (pv *propertyValidator) postalCodeRequired(p *Property) error {
	if p.PostalCode == "" {
		return ErrPostalCodeRequired
	}
	return nil
}

// Validator functions
type propertyValidationFn func(p *Property) error

func runPropertyValFns(p *Property, fns ...propertyValidationFn) error {
	for _, fn := range fns {
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}
