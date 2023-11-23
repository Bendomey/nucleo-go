package errors

type NucleoValidationError struct {
	NucleoClientError
}

type NewNucleoValidationErrorInput struct {
	Message string
	Data    interface{}
}

func NewNucleoValidationError(input NewNucleoValidationErrorInput) NucleoValidationError {

	validationCode := 422
	nucleoClientError := NewNucleoClientError(NewNucleoClientErrorInput{
		Message: &input.Message,
		Code:    &validationCode,
		Type:    "VALIDATION_ERROR",
		Data:    input.Data,
	})

	return NucleoValidationError{
		NucleoClientError: nucleoClientError,
	}
}

func (e *NucleoValidationError) Error() string {
	return e.Message
}
