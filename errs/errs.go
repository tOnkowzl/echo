package errs

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	// BadRequestCode is error code when request is invalid
	BadRequestCode = "20000"

	// InternalServerErrorCode is error code when error occur in app
	InternalServerErrorCode = "80000"

	// ExternalErrorCode is error code when other api return unexpected result
	ExternalErrorCode = "80001"

	// ExternalTimeoutCode is error code for other api timeout
	ExternalTimeoutCode = "80002"
)

func NewBadRequest(msg, description string) *Errs {
	return New(http.StatusBadRequest, BadRequestCode, msg, description)
}

func NewInternalServerError(msg, description string) *Errs {
	return New(http.StatusInternalServerError, InternalServerErrorCode, msg, description)
}

func NewExternalError(msg, description string) *Errs {
	return New(http.StatusInternalServerError, ExternalErrorCode, msg, description)
}

func NewExternalTimeout(msg, description string) *Errs {
	return New(http.StatusInternalServerError, ExternalTimeoutCode, msg, description)
}

func JSON(c echo.Context, err error) error {
	var errs *Errs
	if errors.As(err, &errs) {
		return c.JSON(err.(*Errs).HTTPStatusCode, err)
	}

	return c.JSON(http.StatusInternalServerError, NewInternalServerError("general", err.Error()))
}

func New(httpStatusCode int, code, msg, description string) *Errs {
	return &Errs{
		HTTPStatusCode: httpStatusCode,
		Code:           code,
		Msg:            msg,
		Description:    description,
	}
}

type Errs struct {
	HTTPStatusCode int    `json:"-"`
	Code           string `json:"code"`
	Msg            string `json:"message"`
	Description    string `json:"description"`
}

func (e *Errs) Error() string {
	return fmt.Sprintf("code:%s, msg:%s, description:%s", e.Code, e.Msg, e.Description)
}
