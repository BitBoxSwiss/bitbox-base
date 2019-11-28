// Package hsm contains the API to talk to the BitBoxBase HSM. The HSM is a platform+edition flavor
// of the BitBox02 firmware. There is an HSM bootloader and firmware; communication happens with
// usart framing over a serial port.
package hsm

import (
	"io"
	"time"

	bb02bootloader "github.com/digitalbitbox/bitbox02-api-go/api/bootloader"
	"github.com/digitalbitbox/bitbox02-api-go/api/common"
	bb02firmware "github.com/digitalbitbox/bitbox02-api-go/api/firmware"
	"github.com/digitalbitbox/bitbox02-api-go/communication/usart"
	"github.com/digitalbitbox/bitbox02-api-go/util/errp"
	"github.com/digitalbitbox/bitbox02-api-go/util/semver"
	"github.com/tarm/serial"
)

const (
	firmwareCMD   = 0x80 + 0x40 + 0x01
	bootloaderCMD = 0x80 + 0x40 + 0x03

	// waiting this long for the firmware or bootloader to boot, after UART communication is
	// established.
	bootTimeoutSeconds = 30
)

// HSM lets you interact with the BitBox02-in-the-BitBoxBase (bootloader and firmware).
type HSM struct {
	serialPort string
}

// NewHSM creates a new instance of HSM, which allows you to talk to the HSM bootloader and
// firmware.
func NewHSM(serialPort string) *HSM {
	return &HSM{
		serialPort: serialPort,
	}
}

// openSerial opens the serial port. This operation is cheap (file descriptor open/close), so it can
// be done before every use. readTimeout is the timeout when waiting for a response of the HSM (this
// function does not block on this!). If nil, defaults to blocking (no timeout).
func (hsm *HSM) openSerial(readTimeout *time.Duration) (*serial.Port, error) {
	readTimeoutOrDefault := 0 * time.Second // blocking
	if readTimeout != nil {
		readTimeoutOrDefault = *readTimeout
	}
	conn, err := serial.OpenPort(&serial.Config{
		Name:        hsm.serialPort,
		Baud:        115200,
		ReadTimeout: readTimeoutOrDefault,
	})
	if err != nil {
		return nil, errp.WithStack(err)
	}
	return conn, nil
}

// getFirmware returns a firmware API instance, with the pairing/handshake already
// processed. Returns an error if the serial port could not be opened or io.EOF if the
// firmware/bootloader behind it is not responding.
func (hsm *HSM) getFirmware() (*bb02firmware.Device, error) {
	conn, err := hsm.openSerial(nil)
	if err != nil {
		return nil, err
	}

	device := bb02firmware.NewDevice(
		// version and product inferred via OP_INFO
		nil, nil,
		&bitbox02Config{},
		&usartCommunication{usart.NewCommunication(conn, firmwareCMD)},
		&bitbox02Logger{},
	)
	if err := device.Init(); err != nil {
		return nil, err
	}
	status := device.Status()
	switch status {
	case bb02firmware.StatusUnpaired:
		// expected, proceed below.
	case bb02firmware.StatusRequireAppUpgrade:
		return nil, errp.New("firmware unsupported, update of the BitBoxBase (middleware) is required")
	case bb02firmware.StatusPairingFailed:
		return nil, errp.New("device was expected to autoconfirm the pairing")
	default:
		return nil, errp.Newf("unexpected status: %v ", status)
	}
	// autoconfirm pairing on the host
	device.ChannelHashVerify(true)
	return device, nil
}

func (hsm *HSM) getBootloader(readTimeout *time.Duration) (*bb02bootloader.Device, error) {
	conn, err := hsm.openSerial(readTimeout)
	if err != nil {
		return nil, err
	}

	return bb02bootloader.NewDevice(
		// hardcoded version for now, in the future could be `nil` with autodetection using
		// OP_INFO
		semver.NewSemVer(1, 0, 1),
		common.ProductBitBoxBaseStandard,
		&usartCommunication{usart.NewCommunication(conn, bootloaderCMD)},
		func(status *bb02bootloader.Status) {
			// TODO
		},
	), nil
}

func (hsm *HSM) isFirmware() (bool, error) {
	quickTimeout := 200 * time.Millisecond
	bootloader, err := hsm.getBootloader(&quickTimeout)
	if err != nil {
		return false, err
	}
	_, _, err = bootloader.Versions()
	if errp.Cause(err) == usart.ErrEndpointUnavailable {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

// waitForBootloader returns a bootloader to use. If the HSM is booted into the firmware, we try to
// reboot into the bootloader first. After a certain timeout, an error is returned.
func (hsm *HSM) waitForBootloader() (*bb02bootloader.Device, error) {
	for second := 0; second < bootTimeoutSeconds; second++ {
		isFirmware, err := hsm.isFirmware()
		if errp.Cause(err) == io.EOF {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !isFirmware {
			return hsm.getBootloader(nil)
		}
		firmware, err := hsm.getFirmware()
		if err != nil {
			return nil, err
		}
		// This is simply the reboot command.
		if err := firmware.UpgradeFirmware(); err != nil {
			firmware.Close()
			return nil, err
		}
		firmware.Close()

		if second > 0 {
			time.Sleep(1 * time.Second)
		}
	}
	return nil, errp.New("waiting for bootloader timed out")
}

// WaitForFirmware returns a firmware to use. If the HSM is booted into the bootloader, we try to
// rebot into the firmware fist. After a certain timeout, an error is returned. In case of error,
// the returned device instance is `nil`.
func (hsm *HSM) WaitForFirmware() (*bb02firmware.Device, error) {
	for second := 0; second < bootTimeoutSeconds; second++ {
		isFirmware, err := hsm.isFirmware()
		if errp.Cause(err) == io.EOF {
			continue
		}
		if err != nil {
			return nil, err
		}
		if isFirmware {
			return hsm.getFirmware()
		}

		bootloader, err := hsm.getBootloader(nil)
		if err != nil {
			return nil, err
		}
		// This is simply the reboot command.
		if err := bootloader.Reboot(); err != nil {
			bootloader.Close()
			return nil, err
		}
		bootloader.Close()
		if second > 0 {
			time.Sleep(time.Second)
		}
	}
	return nil, errp.New("waiting for firmware timed out")
}

// InteractWithBootloader lets you talk to the bootloader, rebooting into it from the firmware first
// if necessary. Returns an error if we fail to connect to it.
func (hsm *HSM) InteractWithBootloader(f func(*bb02bootloader.Device)) error {
	device, err := hsm.waitForBootloader()
	if err != nil {
		return err
	}
	defer device.Close()

	f(device)
	return nil
}
