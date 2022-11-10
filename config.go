package telemetry

import (
	"errors"

	"github.com/luraproject/lura/v2/config"
)

var ErrNoConfig = errors.New("unable to load custom config")

func ConfigGetter(e config.ExtraConfig) (interface{}, error) {
	v, ok := e[Namespace]
	if !ok {
		return nil, ErrNoConfig
	}
	telemetryMap, ok := v.(map[string]interface{})
	if !ok {
		return nil, ErrNoConfig
	}
	telemetryCfg := TelemetryConfig{}

	if telemetryMap["logging"] != nil {

		loggingMap, ok := telemetryMap["logging"].(map[string]interface{})
		if !ok {
			return nil, ErrNoConfig
		}

		loggingCfg := defaultLoggingConfigGetter()
		if skipPaths, ok := loggingMap["skip_paths"].([]interface{}); ok {
			var paths []string
			for _, skipPath := range skipPaths {
				if path, ok := skipPath.(string); ok {
					paths = append(paths, path)
				}
			}
			loggingCfg.SkipPaths = paths
		}
		loggingCfg.Level = loggingMap["level"].(string)
		loggingCfg.Module = loggingMap["module"].(string)
		if loggingMap["json"] != nil {
			ecs := loggingMap["json"].(map[string]interface{})
			loggingCfg.ECSFormatter = &ElasticCommonSchemaFormatter{}
			if ecs["disable_html_escape"] != nil {
				loggingCfg.ECSFormatter.DisableHTMLEscape = ecs["disable_html_escape"].(bool)
			}
			if ecs["pretty_print"] != nil {
				loggingCfg.ECSFormatter.PrettyPrint = ecs["pretty_print"].(bool)
			}
			if ecs["data_key"] != nil {
				loggingCfg.ECSFormatter.DataKey = ecs["data_key"].(string)
			}
		}
		telemetryCfg.Logging = loggingCfg
	}

	if telemetryMap["tracing"] != nil {
		tracingMap, ok := telemetryMap["tracing"].(map[string]interface{})
		if !ok {
			return nil, ErrNoConfig
		}

		tracingCfg := TracingConfig{}
		tracingCfg.ExportUrl = tracingMap["exporter_url"].(string)
		tracingCfg.Attributes = defaultTraceResourceAttributes() // TODO : FINISH

		telemetryCfg.Tracing = tracingCfg
	}

	return telemetryCfg, nil
}

func defaultLoggingConfigGetter() LoggingConfig {
	return LoggingConfig{
		Level:  "INFO",
		Module: "DEFAULT",

		SkipPaths: []string{},
	}
}

type TelemetryConfig struct {
	Logging LoggingConfig `json:"logging`
	Tracing TracingConfig `json:"tracing`
}

type LoggingConfig struct {
	SkipPaths    []string                      `json:"skip_paths`
	Level        string                        `json:"level"`
	Module       string                        `json:"module"`
	ECSFormatter *ElasticCommonSchemaFormatter `json:"json"`
}

type ElasticCommonSchemaFormatter struct {
	DisableHTMLEscape bool   `json:"disable_html_escape"`
	DataKey           string `json:"data_key"`
	PrettyPrint       bool   `json:"pretty_print"`
}

type TracingConfig struct {
	ExportUrl  string                    `json:"exporter_url`
	Attributes TracingResourceAttributes `json:"attributes"`
}

func defaultTraceResourceAttributes() TracingResourceAttributes {
	return TracingResourceAttributes{
		service:     "krakend",
		environment: "development",
		id:          "1",
	}
}

type TracingResourceAttributes struct {
	service     string
	environment string
	id          string
}
