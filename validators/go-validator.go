package validators

import (
	"github.com/Bendomey/nucleo-go"
	goValidator "github.com/go-playground/validator/v10"
)

type GoValidator struct {
	validator *goValidator.Validate
}

func NewGoValidator() *GoValidator {
	validator := goValidator.New()
	return &GoValidator{
		validator: validator,
	}
}

func (g *GoValidator) Validate(params nucleo.Payload, schema map[string]interface{}) map[string]interface{} {
	errs := g.validator.ValidateMap(params.RawMap(), schema)

	if len(errs) > 0 {
		return errs
	}

	return nil
}
