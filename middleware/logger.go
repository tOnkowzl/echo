package middleware

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var (
	// DefaultSkipper default of skipper
	DefaultSkipper = func(c echo.Context) bool { return false }
)

const (
	logKeywordDontChange = "api_summary"
)

// Skipper skip middleware
type Skipper func(c echo.Context) bool

// Logger log request information
func Logger(appname string, logger logrus.FieldLogger, skipper Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			logger.WithFields(logrus.Fields{
				"id":        res.Header().Get(echo.HeaderXRequestID),
				"method":    req.Method,
				"path_uri":  req.RequestURI,
				"remote_ip": c.RealIP(),
				"status":    res.Status,
				"latency":   stop.Sub(start).String(),
				"service":   appname,
			}).Info(logKeywordDontChange)
			return
		}
	}
}

func loggerWithID(logger logrus.FieldLogger, c echo.Context) logrus.FieldLogger {
	return logger.WithFields(logrus.Fields{"id": c.Response().Header().Get(echo.HeaderXRequestID)})
}

func LogRequestBody(logger logrus.FieldLogger, skipper Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			reqBody := []byte{}
			if c.Request().Body != nil {
				reqBody, _ = ioutil.ReadAll(c.Request().Body)
			}
			c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

			loggerWithID(logger, c).WithFields(logrus.Fields{
				"header": c.Request().Header,
				"body":   string(reqBody),
			}).Info("request information")

			return next(c)
		}
	}
}

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func LogResponseBody(logger logrus.FieldLogger, skipper Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(c.Response().Writer, resBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
			c.Response().Writer = writer

			if err := next(c); err != nil {
				c.Error(err)
			}

			loggerWithID(logger, c).WithField("body", string(resBody.Bytes())).Info("response information")

			return nil
		}
	}
}
