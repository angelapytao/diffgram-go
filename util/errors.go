package util

import "fmt"

type AppError struct {
	Code       int
	Msg        string
	HTTPStatus int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

var (
	ErrNotFound      = &AppError{Code: 1001, Msg: "not found", HTTPStatus: 404}
	ErrUnauthorized  = &AppError{Code: 1002, Msg: "unauthorized", HTTPStatus: 401}
	ErrForbidden     = &AppError{Code: 1003, Msg: "forbidden", HTTPStatus: 403}
	ErrInvalidInput  = &AppError{Code: 1004, Msg: "invalid input", HTTPStatus: 400}
	ErrInternal      = &AppError{Code: 1005, Msg: "internal error", HTTPStatus: 500}
	ErrWrongPassword = &AppError{Code: 1006, Msg: "wrong password", HTTPStatus: 401}
)
