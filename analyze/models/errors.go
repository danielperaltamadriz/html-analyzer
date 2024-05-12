package models

type ErrorType int

const (
	ErrTypeUnknown         ErrorType = iota
	ErrTypeInvalidURL                // 1
	ErrInvalidRequest                // 2
	ErrTypeInvalidResponse           // 3
)

type Error struct {
	Type               ErrorType
	ResponseStatusCode int
	Message            string
}

func NewError(t ErrorType, msg string) *Error {
	return &Error{
		Type:    t,
		Message: msg,
	}
}

func NewErrorWithStatusCode(t ErrorType, msg string, statusCode int) *Error {
	return &Error{
		Type:               t,
		ResponseStatusCode: statusCode,
		Message:            msg,
	}
}

func (e *Error) Error() string {
	return e.Message
}
