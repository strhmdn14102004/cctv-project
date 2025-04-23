package utils

import (
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()

	// You can add custom validations here if needed
	// Example:
	// _ = Validate.RegisterValidation("customValidation", func(fl validator.FieldLevel) bool {
	//     // validation logic
	// })
}
