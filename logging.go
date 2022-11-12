// Copyright 2021 Faisal Alam. All rights reserved.
// Use of this source code is governed by a Apache style
// license that can be found in the LICENSE file.

package telemetry

import (
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
	"go.opentelemetry.io/otel/trace"
)

const (
	Namespace  = "github_com/darren-bell-optiva/optiva-product-shared-krakend-telemetry"
	moduleName = "telemetry"
)

func NewGinLogger(cfg config.ExtraConfig, loggerConfig gin.LoggerConfig) gin.HandlerFunc {
	telemetryConfig, err := ConfigGetter(cfg)
	if err != nil {
		panic(err)
	}
	logrusGinConfiguration := telemetryConfig.(TelemetryConfig).Logging
	var logger = logrus.StandardLogger()
	logger.SetFormatter(&ecslogrus.Formatter{
		DisableHTMLEscape: logrusGinConfiguration.ECSFormatter.DisableHTMLEscape,
		// DataKey:           logrusGinConfiguration.ECSFormatter.DataKey,
		PrettyPrint: logrusGinConfiguration.ECSFormatter.PrettyPrint,
	})

	// if !ok {
	// 	return gin.LoggerWithConfig(loggerConfig)
	// }

	loggerConfig.SkipPaths = logrusGinConfiguration.SkipPaths
	logger.Info(fmt.Sprintf("%s: total skip paths set: %d", moduleName, len(logrusGinConfiguration.SkipPaths)))

	loggerConfig.Output = io.Discard
	logrusFormatter := LogrusFormatter{logger, logrusGinConfiguration}
	loggerConfig.Formatter = logrusFormatter.AccessLogFormatter

	return gin.LoggerWithConfig(loggerConfig)
}

type LogrusFormatter struct {
	logger *logrus.Logger
	config LoggingConfig
}

func (lf LogrusFormatter) AccessLogFormatter(params gin.LogFormatterParams) string {

	span := trace.SpanFromContext(params.Request.Context())
	fields := logrus.Fields{
		"http.request.method":       params.Method,
		"http.hostname":             params.Request.Host,
		"url.original":              params.Path,
		"http.response.status_code": params.StatusCode,
		"user_agent.original":       params.Request.UserAgent(),
		"source.ip":                 params.ClientIP,
		"event.kind":                "event",
		"event.category":            "web",
		"event.type":                "access",
		"event.module":              "krakend",
		"event.duration":            params.Latency,
		"event.start":               params.TimeStamp.Add(-params.Latency),
		"event.end":                 params.TimeStamp,
		"span.id":                   span.SpanContext().SpanID().String(),
		"trace.id":                  span.SpanContext().TraceID().String(),
	}

	lf.logger.WithFields(fields).Info(params.Method + " " + params.Path)

	return ""
}

/////
////
////

// ErrWrongConfig is the error returned when there is no config under the namespace
var ErrWrongConfig = errors.New("getting the extra config for the krakend-logrus module")

// NewLogger returns a krakend logger wrapping a logrus logger
func NewApplicationLogger(cfg config.ExtraConfig) (*Logger, error) {
	telemetryConfig, err := ConfigGetter(cfg)
	if err != nil {
		panic(err)
	}
	logConfig := telemetryConfig.(TelemetryConfig).Logging

	level, ok := logLevels[logConfig.Level]
	if !ok {
		return nil, fmt.Errorf("unknown log level: %s", logConfig.Level)
	}

	l := logrus.New()
	l.Formatter = &ecslogrus.Formatter{
		DisableHTMLEscape: logConfig.ECSFormatter.DisableHTMLEscape,
		// DataKey:           logConfig.ECSFormatter.DataKey,
		PrettyPrint: logConfig.ECSFormatter.PrettyPrint,
	}
	l.Level = level

	return &Logger{
		Logger: l,
		level:  level,
		module: logConfig.Module,
	}, nil
}

// Logger is a wrapper over a github.com/sirupsen/logrus logger
type Logger struct {
	Logger *logrus.Logger
	level  logrus.Level
	module string
}

// Debug implements the logger interface
func (l *Logger) Debug(v ...interface{}) {
	if l.level < logrus.DebugLevel {
		return
	}

	l.Logger.WithField("module", l.module).Debug(v...)
}

// Info implements the logger interface
func (l *Logger) Info(v ...interface{}) {
	if l.level < logrus.InfoLevel {
		return
	}
	l.Logger.WithField("module", l.module).Info(v...)
}

// Warning implements the logger interface
func (l *Logger) Warning(v ...interface{}) {
	if l.level < logrus.WarnLevel {
		return
	}
	l.Logger.WithField("module", l.module).Warning(v...)
}

// Error implements the logger interface
func (l *Logger) Error(v ...interface{}) {
	if l.level < logrus.ErrorLevel {
		return
	}
	l.Logger.WithField("module", l.module).Error(v...)
}

// Critical implements the logger interface but demotes to the error level
func (l *Logger) Critical(v ...interface{}) {
	l.Logger.WithField("module", l.module).Error(v...)
}

// Fatal implements the logger interface
func (l *Logger) Fatal(v ...interface{}) {
	l.Logger.WithField("module", l.module).Fatal(v...)
}

var logLevels = map[string]logrus.Level{
	"DEBUG":    logrus.DebugLevel,
	"INFO":     logrus.InfoLevel,
	"WARNING":  logrus.WarnLevel,
	"ERROR":    logrus.ErrorLevel,
	"CRITICAL": logrus.FatalLevel,
}
