package validators

import "github.com/Bendomey/nucleo-go"

type JoiValidator struct {
}

func NewJoiValidator() *JoiValidator {
	return &JoiValidator{}
}

func (g *JoiValidator) Validate(params nucleo.Payload, schema interface{}) map[string]interface{} {

	// @TODO: implement joi validator
	return map[string]interface{}{}

}
