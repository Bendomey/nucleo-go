package validators

import "github.com/Bendomey/nucleo-go"

func Resolve(config nucleo.Config) Validator {
	return NewValidator(NewValidatorInput{
		Type: config.Validator,
	})
}
