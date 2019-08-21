//package system provides functionality to get data such as open ports and services running on the system the middleware is deployed on
package system

// Environment provides some information on the system we are running on.
type Environment struct {
	Network            string `json:"network"`
	ElectrsRPCPort     string `json:"electrsRPCPort"`
	bitcoinRPCUser     string
	bitcoinRPCPassword string
	bitcoinRPCPort     string
	lightningRPCPath   string
	bbbConfigScript    string
	bbbCmdScript       string
	prometheusURL      string
}

// NewEnvironment returns a new Environment instance.
//func NewEnvironment(bitcoinRPCUser, bitcoinRPCPassword, bitcoinRPCPort, lightningRPCPath, electrsRPCPort, network string) Environment {
func NewEnvironment(argumentMap map[string]string) Environment {
	// TODO(TheCharlatan) Instead of just accepting a long list of arguments, use a map here and check if the arguments can be read from a system config.
	environment := Environment{
		bitcoinRPCUser:     argumentMap["bitcoinRPCUser"],
		bitcoinRPCPassword: argumentMap["bitcoinRPCPassword"],
		bitcoinRPCPort:     argumentMap["bitcoinRPCPort"],
		lightningRPCPath:   argumentMap["lightningRPCPath"],
		ElectrsRPCPort:     argumentMap["electrsRPCPort"],
		Network:            argumentMap["network"],
		bbbConfigScript:    argumentMap["bbbConfigScript"],
		bbbCmdScript:       argumentMap["bbbCmdScript"],
		prometheusURL:      argumentMap["prometheusURL"],
	}
	return environment
}

// GetBitcoinRPCUser is a getter for bitcoinRPCUser
func (environment *Environment) GetBitcoinRPCUser() string {
	return environment.bitcoinRPCUser
}

// GetBitcoinRPCPassword is a getter for the bitcoinRPCPassword
func (environment *Environment) GetBitcoinRPCPassword() string {
	return environment.bitcoinRPCPassword
}

// GetBitcoinRPCPort is a getter for the bitcoinRPCPort
func (environment *Environment) GetBitcoinRPCPort() string {
	return environment.bitcoinRPCPort
}

// GetLightningRPCPath is a getter for the lightningRPCPath
func (environment *Environment) GetLightningRPCPath() string {
	return environment.lightningRPCPath
}

// GetBBBConfigScript is a getter for the location of the bbb config script
func (environment *Environment) GetBBBConfigScript() string {
	return environment.bbbConfigScript
}

// GetBBBCmdScript is a getter for the location of the bbb cmd script
func (environment *Environment) GetBBBCmdScript() string {
	return environment.bbbCmdScript
}

// GetPrometheusURL is a getter for the url the prometheus server is reachable on
func (environment *Environment) GetPrometheusURL() string {
	return environment.prometheusURL
}
