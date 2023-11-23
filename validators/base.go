package validators

import (
	"github.com/Bendomey/nucleo-go"
	"github.com/Bendomey/nucleo-go/errors"
)

type Validator interface {
	Validate(params nucleo.Payload, schema interface{}) map[string]interface{}
	GetValidatorName() nucleo.ValidatorType
	Middlewares() nucleo.Middlewares
}

type ValidatorContext struct {
	Type         nucleo.ValidatorType
	GoValidator  *GoValidator
	JoiValidator *JoiValidator
}

type NewValidatorInput struct {
	Type nucleo.ValidatorType
}

func NewValidator(input NewValidatorInput) Validator {
	var goValidator *GoValidator
	if input.Type == nucleo.ValidatorGo {
		goValidator = NewGoValidator()
	}

	var joiValidator *JoiValidator
	if input.Type == nucleo.ValidatorGo {
		joiValidator = NewJoiValidator()
	}

	return &ValidatorContext{
		Type:         input.Type,
		GoValidator:  goValidator,
		JoiValidator: joiValidator,
	}
}

func (validator ValidatorContext) Validate(params nucleo.Payload, schema interface{}) map[string]interface{} {
	switch validator.Type {
	case nucleo.ValidatorGo:
		return validator.GoValidator.Validate(params, schema.(map[string]interface{}))
	case nucleo.ValidatorJoi:
		return validator.JoiValidator.Validate(params, schema)
	}

	// default to go validator
	return validator.GoValidator.Validate(params, schema.(map[string]interface{}))
}

func (validator ValidatorContext) GetValidatorName() nucleo.ValidatorType {
	return validator.Type
}

func (validator ValidatorContext) Middleware(rawCtx interface{}, next func(...interface{})) {
	context := rawCtx.(nucleo.BrokerContext)
	validationErrors := validator.Validate(context.Payload(), context.PayloadSchema())

	if len(validationErrors) > 0 {
		next(errors.NewNucleoValidationError(errors.NewNucleoValidationErrorInput{
			Message: "validation error", // @TODO: make this a constant and get better validation error message.
			Data:    validationErrors,
		}))
		return
	}

	next()
}

func (validator ValidatorContext) Middlewares() nucleo.Middlewares {
	return map[string]nucleo.MiddlewareHandler{
		"beforeLocalAction":  validator.Middleware,
		"beforeRemoteAction": validator.Middleware,
	}
}
