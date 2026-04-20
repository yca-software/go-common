package validator

import (
	"errors"

	validationLib "github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Tag   string `json:"tag"`
	Param string `json:"param"`
	Value any    `json:"value"`
	Error string `json:"error"`
}

type Validator interface {
	ValidateStruct(s any) *map[string]ValidationError
}

type validatorImpl struct {
	validate *validationLib.Validate
}

func New() Validator {
	return &validatorImpl{
		validate: validationLib.New(),
	}
}

func (v *validatorImpl) ValidateStruct(s any) *map[string]ValidationError {
	if s == nil {
		return nil
	}

	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var verrs validationLib.ValidationErrors
	if !errors.As(err, &verrs) {
		// Non-validation error (e.g. invalid struct type); return single generic entry
		errMap := map[string]ValidationError{"": {Error: err.Error()}}
		return &errMap
	}

	result := make(map[string]ValidationError, len(verrs))
	for _, e := range verrs {
		result[e.Field()] = ValidationError{
			Tag:   e.Tag(),
			Param: e.Param(),
			Value: e.Value(),
			Error: e.Error(),
		}
	}
	return &result
}
