package echo

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/tOnkowzl/echo/middleware"
)

func New(logger logrus.FieldLogger, appName string, skipPath []string) *Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	skipper := makeSkipper(skipPath)

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover(logger))
	e.Use(middleware.LogRequestBody(logger, skipper))
	e.Use(middleware.Logger(appName, logger, skipper))
	e.Use(middleware.LogResponseBody(logger, skipper))

	return &Echo{
		Echo:   e,
		Logger: logger,
	}
}

type Echo struct {
	*echo.Echo

	Logger logrus.FieldLogger
}

func (e *Echo) Start(port string) {
	go func() {
		e.Logger.Info("Listening on port: ", port)
		e.Logger.Info(e.Echo.Start(":" + port))
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal("shutdown server:", err)
	}
}

func makeSkipper(skipPath []string) middleware.Skipper {
	var skipper map[string]bool
	for _, v := range skipPath {
		skipper[v] = true
	}

	return func(c echo.Context) bool {
		return skipper[c.Path()]
	}
}
