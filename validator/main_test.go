package validator_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	validator "github.com/yca-software/go-common/validator"
)

type ValidStruct struct {
	Email    string `validate:"required,email"`
	Age      int    `validate:"min=18,max=100"`
	Username string `validate:"required,min=3,max=20"`
}

type InvalidStruct struct {
	Email    string `validate:"required,email"`
	Age      int    `validate:"min=18,max=100"`
	Username string `validate:"required,min=3,max=20"`
}

type ValidationTestSuite struct {
	suite.Suite
	validator validator.Validator
}

func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}

func (s *ValidationTestSuite) SetupTest() {
	s.validator = validator.New()
}

func (s *ValidationTestSuite) TestNew() {
	validator := validator.New()
	s.NotNil(validator)
}

func (s *ValidationTestSuite) TestValidateStruct_Valid() {
	valid := ValidStruct{
		Email:    "user@example.com",
		Age:      25,
		Username: "johndoe",
	}

	errors := s.validator.ValidateStruct(valid)
	s.Nil(errors)
}

func (s *ValidationTestSuite) TestValidateStruct_Invalid() {
	invalid := InvalidStruct{
		Email:    "invalid-email", // Invalid email format
		Age:      15,              // Below minimum
		Username: "ab",            // Too short
	}

	errors := s.validator.ValidateStruct(invalid)
	s.NotNil(errors)
	s.Greater(len(*errors), 0)
}

func (s *ValidationTestSuite) TestValidateStruct_MultipleErrors() {
	invalid := InvalidStruct{
		Email:    "",  // Missing required
		Age:      200, // Above maximum
		Username: "",  // Missing required
	}

	errors := s.validator.ValidateStruct(invalid)
	s.NotNil(errors)

	errorMap := *errors
	s.Contains(errorMap, "Email")
	s.Contains(errorMap, "Age")
	s.Contains(errorMap, "Username")
}

func (s *ValidationTestSuite) TestValidateStruct_ErrorStructure() {
	invalid := InvalidStruct{
		Email:    "invalid-email",
		Age:      25,
		Username: "johndoe",
	}

	errors := s.validator.ValidateStruct(invalid)
	s.NotNil(errors)

	errorMap := *errors
	emailError, exists := errorMap["Email"]
	s.True(exists)
	s.NotEmpty(emailError.Tag)
	s.NotEmpty(emailError.Error)
}

func (s *ValidationTestSuite) TestValidateStruct_EmptyStruct() {
	type EmptyStruct struct{}

	empty := EmptyStruct{}
	errors := s.validator.ValidateStruct(empty)
	s.Nil(errors)
}

func (s *ValidationTestSuite) TestValidateStruct_NilInput() {
	// Nil input is guarded; returns nil (no validation errors)
	errors := s.validator.ValidateStruct(nil)
	s.Nil(errors)
}

func (s *ValidationTestSuite) TestValidateStruct_NonStructReturnsGenericError() {
	// Passing non-struct (e.g. string) yields non-ValidationErrors error; should not panic
	errs := s.validator.ValidateStruct("not a struct")
	s.NotNil(errs)
	s.Contains(*errs, "")
	s.NotEmpty((*errs)[""].Error)
}

func (s *ValidationTestSuite) TestValidationError_Fields() {
	invalid := InvalidStruct{
		Email:    "not-an-email",
		Age:      10,
		Username: "x",
	}

	errors := s.validator.ValidateStruct(invalid)
	s.NotNil(errors)

	errorMap := *errors
	for field, validationError := range errorMap {
		s.NotEmpty(field)
		s.NotEmpty(validationError.Tag)
		s.NotEmpty(validationError.Error)
		// Param and Value may be empty depending on the validation rule
	}
}
