//package system provides functionality to get data such as open ports and services running on the system the middleware is deployed on
package system

// Environment provides some information on the system we are running on.
type Environment struct {
	network                                            string
	electrsRPCPort                                     string
	bitcoinRPCUser, bitcoinRPCPassword, bitcoinRPCPort string
	lightningRPCPath                                   string
}

// NewEnvironment returns a new Environment instance.
func NewEnvironment(bitcoinRPCUser, bitcoinRPCPassword, bitcoinRPCPort, lightningRPCPath, electrsRPCPort, network string) Environment {
	// TODO(TheCharlatan) Instead of just accepting a long list of arguments, use a map here and check if the arguments can be read from a system config.
	environment := Environment{
		bitcoinRPCUser:     bitcoinRPCUser,
		bitcoinRPCPassword: bitcoinRPCPassword,
		bitcoinRPCPort:     bitcoinRPCPort,
		lightningRPCPath:   lightningRPCPath,
		electrsRPCPort:     electrsRPCPort,
		network:            network,
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

// GetElectrsRPCPort is a getter for the electrsRPCPort
func (environment *Environment) GetElectrsRPCPort() string {
	return environment.electrsRPCPort
}

// GetNetwork is a getter for the network type (testnet/mainnet)
func (environment *Environment) GetNetwork() string {
	return environment.network
}
