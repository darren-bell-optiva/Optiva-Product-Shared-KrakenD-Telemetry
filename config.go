package telemetry

import (
	"github.com/luraproject/lura/v2/config"
)

func ConfigGetter(e config.ExtraConfig) interface{} {
	v, ok := e[Namespace]
	if !ok {
		return nil
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}

	cfg := defaultConfigGetter()
	if skipPaths, ok := tmp["skip_paths"].([]interface{}); ok {
		var paths []string
		for _, skipPath := range skipPaths {
			if path, ok := skipPath.(string); ok {
				paths = append(paths, path)
			}
		}
		cfg.SkipPaths = paths
	}
	cfg.Level = tmp["level"].(string)
	cfg.Module = tmp["module"].(string)
	if tmp["json"] != nil {
		ecs := tmp["json"].(map[string]interface{})
		cfg.ECSFormatter = &ElasticCommonSchemaFormatter{}
		if ecs["disable_html_escape"] != nil {
			cfg.ECSFormatter.DisableHTMLEscape = ecs["disable_html_escape"].(bool)
		}
		if ecs["pretty_print"] != nil {
			cfg.ECSFormatter.PrettyPrint = ecs["pretty_print"].(bool)
		}
		if ecs["data_key"] != nil {
			cfg.ECSFormatter.DataKey = ecs["data_key"].(string)
		}
	}

	return cfg
}

func defaultConfigGetter() Config {
	return Config{
		Level:  "INFO",
		Module: "DEFAULT",

		SkipPaths: []string{},
	}
}

type Config struct {
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
