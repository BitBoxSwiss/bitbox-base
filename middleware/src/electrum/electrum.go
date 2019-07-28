package electrum

import (
	"bufio"
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/log"
)

// Electrum makes a connection to an Electrum server and proxies messages.
type Electrum struct {
	connection        net.Conn
	onMessageReceived func([]byte)
}

// NewElectrum creates a new Electrum instance and tries to connect to the server.
func NewElectrum(address string, onMessageReceived func([]byte)) (*Electrum, error) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	electrum := &Electrum{
		connection:        connection,
		onMessageReceived: onMessageReceived,
	}
	go electrum.read()
	return electrum, nil
}

func (electrum *Electrum) read() {
	reader := bufio.NewReader(electrum.connection)
	for {
		line, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			log.Error(fmt.Sprintf("electrum read error: %v", err))
			break
		}
		electrum.onMessageReceived(line)
	}
}

// Send sends a raw message to the Electrum server.
func (electrum *Electrum) Send(msg []byte) error {
	_, err := electrum.connection.Write(msg)
	return err
}
