// Package system provides functionality to get data such as open ports and services running on the system the middleware is deployed on
package system

// Environment provides some information on the system we are running on.
type Environment struct {
	middlewarePort            string
	Network                   string `json:"network"`
	ElectrsRPCPort            string `json:"electrsRPCPort"`
	bbbConfigScript           string
	bbbCmdScript              string
	bbbSystemctlScript        string
	prometheusURL             string
	redisPort                 string
	middlewareVersion         string
	imageUpdateInfoURL        string
	notificationNamedPipePath string
}

// NewEnvironment returns a new Environment instance.
func NewEnvironment(argumentMap map[string]string) Environment {
	// TODO(TheCharlatan) Instead of just accepting a long list of arguments, use a map here and check if the arguments can be read from a system config.
	environment := Environment{
		middlewarePort:            argumentMap["middlewarePort"],
		ElectrsRPCPort:            argumentMap["electrsRPCPort"],
		Network:                   argumentMap["network"],
		bbbConfigScript:           argumentMap["bbbConfigScript"],
		bbbCmdScript:              argumentMap["bbbCmdScript"],
		bbbSystemctlScript:        argumentMap["bbbSystemctlScript"],
		prometheusURL:             argumentMap["prometheusURL"],
		redisPort:                 argumentMap["redisPort"],
		middlewareVersion:         argumentMap["middlewareVersion"],
		imageUpdateInfoURL:        argumentMap["imageUpdateInfoURL"],
		notificationNamedPipePath: argumentMap["notificationNamedPipePath"],
	}
	return environment
}

// GetBBBConfigScript is a getter for the location of the bbb config script
func (environment *Environment) GetBBBConfigScript() string {
	return environment.bbbConfigScript
}

// GetBBBCmdScript is a getter for the location of the bbb cmd script
func (environment *Environment) GetBBBCmdScript() string {
	return environment.bbbCmdScript
}

// GetBBBSystemctlScript is a getter for the location of the bbb-systemctl.sh script
func (environment *Environment) GetBBBSystemctlScript() string {
	return environment.bbbSystemctlScript
}

// GetPrometheusURL is a getter for the url the prometheus server is reachable on
func (environment *Environment) GetPrometheusURL() string {
	return environment.prometheusURL
}

// GetRedisPort is a getter for the port the redis server is listening on
func (environment *Environment) GetRedisPort() string {
	return environment.redisPort
}

// GetMiddlewareVersion is a getter for the middleware version
func (environment *Environment) GetMiddlewareVersion() string {
	return environment.middlewareVersion
}

// GetMiddlewarePort is a getter for the port the middleware is listening on
func (environment *Environment) GetMiddlewarePort() string {
	return environment.middlewarePort
}

// GetImageUpdateInfoURL is a getter for the URL that specifies where the middleware queries the update info.
func (environment *Environment) GetImageUpdateInfoURL() string {
	return environment.imageUpdateInfoURL
}

// GetNotificationNamedPipePath is a getter for the path where the middleware creates and looks for the
func (environment *Environment) GetNotificationNamedPipePath() string {
	return environment.notificationNamedPipePath
}
