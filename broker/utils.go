package broker

import (
	"github.com/Bendomey/nucleo-go/nucleo"
)

func mergeMaps(base, new map[string]interface{}) map[string]interface{} {
	if base == nil {
		base = map[string]interface{}{}
	}
	for key, value := range new {
		base[key] = value
	}
	return base
}

func mergeConfigs(baseConfig nucleo.Config, userConfig nucleo.Config) nucleo.Config {
	if userConfig.Services != nil {
		baseConfig.Services = mergeMaps(baseConfig.Services, userConfig.Services)
	}

	if userConfig.LogLevel != "" {
		baseConfig.LogLevel = userConfig.LogLevel
	}
	if userConfig.LogFormat != "" {
		baseConfig.LogFormat = userConfig.LogFormat
	}
	if userConfig.DiscoverNodeID != nil {
		baseConfig.DiscoverNodeID = userConfig.DiscoverNodeID
	}

	if userConfig.RequestTimeout != 0 {
		baseConfig.RequestTimeout = userConfig.RequestTimeout
	}

	if userConfig.Namespace != "" {
		baseConfig.Namespace = userConfig.Namespace
	}
	return baseConfig
}
