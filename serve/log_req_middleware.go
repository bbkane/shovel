package main

import (
	"fmt"
	"net/http/httputil"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// LogReqMiddleware logs the HTTP request to the server. Maybe at some point I'll log the resp too, but that looks harder
func LogReqMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			dumpedReq, err := httputil.DumpRequest(req, true)
			if err != nil {
				return fmt.Errorf("error dumping req: %w", err)
			}
			// resp := c.Response().
			// dumpedResp, err := httputil.DumpResponse(resp, true)
			c.Logger().Debugj(log.JSON{
				"message": "req dump",
				"req":     string(dumpedReq),
			})
			return next(c)
		}
	}
}
