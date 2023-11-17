package errors

type NucleoRetryableError struct {
	NucleoError
}

type NewNucleoRetryableErrorInput struct {
	Code *int
	Type string
	Data interface{}
}

func NewNucleoRetryableError(input NewNucleoRetryableErrorInput) NucleoRetryableError {

	return NucleoRetryableError{
		NucleoError: NucleoError{
			Type:      input.Type,
			Data:      input.Data,
			Retryable: true,
		},
	}
}

func (e *NucleoRetryableError) Error() string {
	return e.Message
}
