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
}

// NewEnvironment returns a new Environment instance.
func NewEnvironment(bitcoinRPCUser, bitcoinRPCPassword, bitcoinRPCPort, lightningRPCPath, electrsRPCPort, network string) Environment {
	// TODO(TheCharlatan) Instead of just accepting a long list of arguments, use a map here and check if the arguments can be read from a system config.
	environment := Environment{
		bitcoinRPCUser:     bitcoinRPCUser,
		bitcoinRPCPassword: bitcoinRPCPassword,
		bitcoinRPCPort:     bitcoinRPCPort,
		lightningRPCPath:   lightningRPCPath,
		ElectrsRPCPort:     electrsRPCPort,
		Network:            network,
	}
	return environment
}

// GetBitcoinRPCUser is a getter for bitcoinRPCUser
func (environment *Environment) GetBitcoinRPCUser() string {
	return environment.bitcoinRPCUser
}

// GetBitcoinRPCPassword is a getter for the bitcoinRPCPassword
func (environment *Environment) GetBitcoinRPCPassword() string {
	return environment.bitcoinRPCUser
}

// GetBitcoinRPCPort is a getter for the bitcoinRPCPort
func (environment *Environment) GetBitcoinRPCPort() string {
	return environment.bitcoinRPCPort
}

// GetLightningRPCPath is a getter for the lightningRPCPath
func (environment *Environment) GetLightningRPCPath() string {
	return environment.lightningRPCPath
}
