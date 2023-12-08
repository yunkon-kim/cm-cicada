package echo

import (
	"fmt"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"strconv"
)

var e *echo.Echo

func Init() {
	e = echo.New()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:    true,
		LogURI:       true,
		LogHost:      true,
		LogRemoteIP:  true,
		LogUserAgent: true,
		LogStatus:    true,
		LogError:     true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.Println(logger.DEBUG, false, "ECHO: Request received. ("+
					"Method: "+v.Method+", "+
					"URI: "+v.URI+", "+
					"RemoteIP: "+v.RemoteIP+", "+
					"UserAgent: "+v.UserAgent+", "+
					"Status: "+strconv.Itoa(v.Status)+", "+
					"Parameters: "+fmt.Sprintf("%v", c.QueryParams())+")")
			} else {
				logger.Println(logger.ERROR, false, "ECHO: Error occurred while processing the request. ("+
					"Method: "+v.Method+", "+
					"URI: "+v.URI+", "+
					"RemoteIP: "+v.RemoteIP+", "+
					"UserAgent: "+v.UserAgent+", "+
					"Status: "+strconv.Itoa(v.Status)+", "+
					"Error: "+v.Error.Error()+", "+
					"Parameters: "+fmt.Sprintf("%v", c.QueryParams())+")")
			}

			return nil
		},
	}))

	DAG()

	err := e.Start(":" + config.CMCicadaConfig.CMCicada.Listen.Port)
	logger.Panicln(logger.ERROR, true, err)
}
