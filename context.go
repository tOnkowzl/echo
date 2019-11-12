package echo

import "github.com/labstack/echo/v4"

type Context interface {
	echo.Context

	SUCCESS(i interface{})
	ERROR(err error)
}
