package broker

import "github.com/Bendomey/nucleo-go/nucleo"

func mergeMaps(base, new map[string]interface{}) map[string]interface{} {
	if base == nil {
		base = map[string]interface{}{}
	}
	for key, value := range new {
		base[key] = value
	}
	return base
}

func mergeConfigs(baseConfig nucleo.Config, userConfig []*nucleo.Config) nucleo.Config {
	if len(userConfig) > 0 {
		for _, config := range userConfig {
			if config.Services != nil {
				baseConfig.Services = mergeMaps(baseConfig.Services, config.Services)
			}

			if config.LogLevel != "" {
				baseConfig.LogLevel = config.LogLevel
			}
			if config.LogFormat != "" {
				baseConfig.LogFormat = config.LogFormat
			}
			if config.DiscoverNodeID != nil {
				baseConfig.DiscoverNodeID = config.DiscoverNodeID
			}
			if config.Transporter != "" {
				baseConfig.Transporter = config.Transporter
			}
			if config.TransporterFactory != nil {
				baseConfig.TransporterFactory = config.TransporterFactory
			}
			if config.StrategyFactory != nil {
				baseConfig.StrategyFactory = config.StrategyFactory
			}
			if config.DisableInternalMiddlewares {
				baseConfig.DisableInternalMiddlewares = config.DisableInternalMiddlewares
			}
			if config.DisableInternalServices {
				baseConfig.DisableInternalServices = config.DisableInternalServices
			}

			if config.DontWaitForNeighbours {
				baseConfig.DontWaitForNeighbours = config.DontWaitForNeighbours
			}

			if config.Middlewares != nil {
				baseConfig.Middlewares = config.Middlewares
			}
			if config.RequestTimeout != 0 {
				baseConfig.RequestTimeout = config.RequestTimeout
			}

			if config.Namespace != "" {
				baseConfig.Namespace = config.Namespace
			}
		}
	}
	return baseConfig
}
