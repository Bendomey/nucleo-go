package validators

import "github.com/Bendomey/nucleo-go"

func Resolve(config nucleo.Config) Validator {
	if config.Validator == nucleo.ValidatorJoi {
		return NewValidator(NewValidatorInput{
			Type: nucleo.ValidatorGo,
		})
	}

	return NewValidator(NewValidatorInput{
		Type: nucleo.ValidatorGo,
	})
}
