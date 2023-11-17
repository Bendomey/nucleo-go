package errors

type NucleoError struct {
	Message   string
	Code      int
	Type      string
	Data      interface{}
	Retryable bool
}

type NewNucleoErrorInput struct {
	Message *string
	Code    *int
	Type    string
	Data    interface{}
}

func NewNucleoError(input NewNucleoErrorInput) NucleoError {
	code := 500
	if input.Code != nil {
		code = *input.Code
	}

	message := "Internal Server Error"
	if input.Message != nil {
		message = *input.Message
	}

	return NucleoError{
		Code:      code,
		Type:      input.Type,
		Data:      input.Data,
		Message:   message,
		Retryable: false,
	}
}

func (e *NucleoError) Error() string {
	return e.Message
}
