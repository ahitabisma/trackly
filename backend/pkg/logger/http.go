package logger

import (
	"trackly-backend/pkg/httpx"

	"github.com/sirupsen/logrus"
)

// LogHTTPResponse logs HTTP responses dengan format: level, msg(json stringify), time
func LogHTTPResponse(log *logrus.Logger, resp httpx.Response, level logrus.Level, fields map[string]interface{}) {
	log.WithFields(logrus.Fields{
		"error": resp.Errors,
		"meta":  resp.Meta,
	}).Log(level, resp.Message)
}

// LogHTTPError is a convenience function to log HTTP errors
func LogHTTPError(log *logrus.Logger, resp httpx.Response, fields map[string]interface{}) {
	LogHTTPResponse(log, resp, logrus.WarnLevel, fields)
}

// LogHTTPSuccess is a convenience function to log HTTP success
func LogHTTPSuccess(log *logrus.Logger, resp httpx.Response, fields map[string]interface{}) {
	LogHTTPResponse(log, resp, logrus.InfoLevel, fields)
}

// LogHTTPInternalError is a convenience function to log HTTP internal server errors
func LogHTTPInternalError(log *logrus.Logger, resp httpx.Response, fields map[string]interface{}) {
	LogHTTPResponse(log, resp, logrus.ErrorLevel, fields)
}
