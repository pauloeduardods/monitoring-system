package validator

import (
	"github.com/go-playground/validator/v10"
)

type Validator interface {
	Validate(s interface{}) error
}

type ValidatorImpl struct {
	validate *validator.Validate
}

func NewValidatorImpl() *ValidatorImpl {
	return &ValidatorImpl{
		validate: validator.New(),
	}
}

func (v *ValidatorImpl) Validate(s interface{}) error {
	err := v.validate.Struct(s)
	if err != nil {
		return err
	}
	return nil
}
