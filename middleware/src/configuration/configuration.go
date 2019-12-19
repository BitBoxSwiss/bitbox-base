// Package configuration provides the functionality to provide configuration
// values to the Middleware.
package configuration

// Args has the same fields as the `Configuration` struct, but the fields in
// `Args` are public. The struct is used as parameter to the `NewConfiguration()`
// factory function. The struct needs public fields to be settable the `main`
// package. However the `Configuration` can't have public fields, because the
// fields should be **READ ONLY** while the Middleware is running. Fields should
// only be accessible via defined Getter functions.
//
// Note: Go does not support named arguments for functions. While passing the
// arguments to a function would work, it would be rather easy to mistakenly
// switch e.g. two string parameters ending up with an invalid middleware
// configuration. Using the `Args` helps, because a Go struct can be initialized
// with named fields.
type Args struct {
	BBBCmdScript              string
	BBBConfigScript           string
	BBBSystemctlScript        string
	ElectrsRPCPort            string
	ImageUpdateInfoURL        string
	MiddlewarePort            string
	MiddlewareVersion         string
	Network                   string
	NotificationNamedPipePath string
	PrometheusURL             string
	RedisMock                 bool
	RedisPort                 string
	HsmFirmwareFile           string
}

// Configuration holds the configuration options for the Middleware.
//
// Note: adding / removing a field in this struct requires an update to the
// `Args` struct as well.
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
	hsmFirmwareFile           string
}

// NewConfiguration returns a new Configuration instance.
//
// Note: The `Args` struct supports named fields. Go functions don't support
// named parameters. The struct helps avoiding switched parameters.
func NewConfiguration(args Args) Configuration {
	config := Configuration{
		bbbCmdScript:              args.BBBCmdScript,
		bbbConfigScript:           args.BBBConfigScript,
		bbbSystemctlScript:        args.BBBSystemctlScript,
		electrsRPCPort:            args.ElectrsRPCPort,
		imageUpdateInfoURL:        args.ImageUpdateInfoURL,
		middlewarePort:            args.MiddlewarePort,
		middlewareVersion:         args.MiddlewareVersion,
		network:                   args.Network,
		notificationNamedPipePath: args.NotificationNamedPipePath,
		prometheusURL:             args.PrometheusURL,
		redisMock:                 args.RedisMock,
		redisPort:                 args.RedisPort,
		hsmFirmwareFile:           args.HsmFirmwareFile,
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

// GetHsmFirmwareFile is a getter for the location of the HSM firmware file.
func (config *Configuration) GetHsmFirmwareFile() string {
	return config.hsmFirmwareFile
}
