package errors

type NucleoClientError struct {
	NucleoError
}

type NewNucleoClientErrorInput struct {
	Message *string
	Code    *int
	Type    string
	Data    interface{}
}

func NewNucleoClientError(input NewNucleoClientErrorInput) NucleoClientError {
	code := 400
	if input.Code != nil {
		code = *input.Code
	}

	message := "A client error occurred"
	if input.Message != nil {
		message = *input.Message
	}

	nucleoError := NewNucleoError(NewNucleoErrorInput{
		Message: &message,
		Code:    &code,
		Type:    input.Type,
		Data:    input.Data,
	})

	return NucleoClientError{
		NucleoError: nucleoError,
	}
}

func (e *NucleoClientError) Error() string {
	return e.Message
}
