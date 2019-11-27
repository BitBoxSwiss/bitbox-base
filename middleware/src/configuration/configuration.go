// Package configuration provides the functionality to provide configuration
// values to the Middleware.
package configuration

// Configuration holds the configuration options for the Middleware.
type Configuration struct {
	bbbCmdScript              string
	bbbConfigScript           string
	bbbSystemctlScript        string
	electrsRPCPort            string
	imageUpdateInfoURL        string
	middlewarePort            string
	middlewareVersion         string
	network                   string
	notificationNamedPipePath string
	prometheusURL             string
	redisMock                 bool
	redisPort                 string
}

// NewConfiguration returns a new Configuration instance.
func NewConfiguration(
	bbbCmdScript string, bbbConfigScript string, bbbSystemctlScript string,
	electrsRPCPort string, imageUpdateInfoURL string, middlewarePort string,
	middlewareVersion string, network string, notificationNamedPipePath string,
	prometheusURL string, redisMock bool, redisPort string,
) Configuration {
	config := Configuration{
		bbbCmdScript:              bbbCmdScript,
		bbbConfigScript:           bbbConfigScript,
		bbbSystemctlScript:        bbbSystemctlScript,
		electrsRPCPort:            electrsRPCPort,
		imageUpdateInfoURL:        imageUpdateInfoURL,
		middlewarePort:            middlewarePort,
		middlewareVersion:         middlewareVersion,
		network:                   network,
		notificationNamedPipePath: notificationNamedPipePath,
		prometheusURL:             prometheusURL,
		redisMock:                 redisMock,
		redisPort:                 redisPort,
	}
	return config
}

// GetBBBConfigScript is a getter for the location of the bbb config script
func (config *Configuration) GetBBBConfigScript() string {
	return config.bbbConfigScript
}

// GetBBBCmdScript is a getter for the location of the bbb cmd script
func (config *Configuration) GetBBBCmdScript() string {
	return config.bbbCmdScript
}

// GetBBBSystemctlScript is a getter for the location of the bbb-systemctl.sh script
func (config *Configuration) GetBBBSystemctlScript() string {
	return config.bbbSystemctlScript
}

// GetPrometheusURL is a getter for the url the prometheus server is reachable on
func (config *Configuration) GetPrometheusURL() string {
	return config.prometheusURL
}

// GetRedisPort is a getter for the port the redis server is listening on
func (config *Configuration) GetRedisPort() string {
	return config.redisPort
}

// IsRedisMock is a getter for the value of the mock parameter for redis
func (config *Configuration) IsRedisMock() bool {
	return config.redisMock
}

// GetMiddlewareVersion is a getter for the middleware version
func (config *Configuration) GetMiddlewareVersion() string {
	return config.middlewareVersion
}

// GetMiddlewarePort is a getter for the port the middleware is listening on
func (config *Configuration) GetMiddlewarePort() string {
	return config.middlewarePort
}

// GetImageUpdateInfoURL is a getter for the URL that specifies where the middleware queries the update info.
func (config *Configuration) GetImageUpdateInfoURL() string {
	return config.imageUpdateInfoURL
}

// GetNotificationNamedPipePath is a getter for the path where the middleware creates and looks for the
func (config *Configuration) GetNotificationNamedPipePath() string {
	return config.notificationNamedPipePath
}

// GetNetwork is a getter for the Bitcoin network (mainnet, testnet, regtest,
//...) the base is configured to use.
func (config *Configuration) GetNetwork() string {
	return config.network
}

// GetElectrsRPCPort is a getter for the electrs RPC port.
func (config *Configuration) GetElectrsRPCPort() string {
	return config.electrsRPCPort
}
