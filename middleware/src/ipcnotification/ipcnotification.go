// Package ipcnotification reads notifications from a named pipe and passes them
// into a channel.
package ipcnotification

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"
)

// Notification represents an IPC notification that is passed via a named pipe.
type Notification struct {
	Version int         `json:"version"`
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload"`
}

func (notification *Notification) String() string {
	return fmt.Sprintf("[Version: %d, Topic: %s, Payload: %v]", notification.Version, notification.Topic, notification.Payload)
}

// Reader reads IPCNotifications from a named pipe into a channel.
type Reader struct {
	notifications chan Notification
	closing       bool
	filePath      string
	namedPipe     *os.File
}

// NewReader returns a new Reader that starts reading from the named pipe passed.
func NewReader(filepath string) (*Reader, error) {
	reader := &Reader{
		notifications: make(chan Notification),
		filePath:      filepath,
		namedPipe:     nil,
		closing:       false,
	}

	_, err := os.Stat(reader.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("Could not find named pipe '%s'. Creating new named pipe.", reader.filePath)
			// The file permission 0622 are set, so that everybody can write into the
			// pipe, but only the middleware can read.
			err := syscall.Mkfifo(reader.filePath, 0600)
			if err != nil {
				return nil, fmt.Errorf("could not create a new named pipe '%s': %w", reader.filePath, err)
			}
		} else {
			return nil, fmt.Errorf("could not stat the named pipe '%s': %w", reader.filePath, err)
		}
	} else {
		log.Printf("Using existing named pipe '%s' for IPC notifications.", reader.filePath)
	}

	// os.OpenFile opens (or creates and opens if not present) the named pipe.
	// Due to a quirk in the Unix named pipe implementation os.O_RDWR has to be
	// used instead of os.O_RDONLY. O_RDONLY blocks unit the pipe is written to.
	// https://stackoverflow.com/a/5782778/8896600
	namedPipe, err := os.OpenFile(reader.filePath, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, fmt.Errorf("could not open the named pipe in '%s': %w", reader.filePath, err)
	}
	reader.namedPipe = namedPipe

	go reader.read()
	return reader, nil
}

func (reader *Reader) read() {
	defer func() {
		err := reader.namedPipe.Close()
		if err != nil {
			log.Printf("Could not close named pipe: %s", err)
		}
	}()
	scanner := bufio.NewScanner(reader.namedPipe)

	for {
		if scanner.Scan() {
			notificationBytes := scanner.Bytes()
			notificationText := scanner.Text()

			// A write to a named pipe is only atomic if less than {PIPE_BUF} bytes
			// are written. {PIPE_BUF} for Linux is 4096. This is enforced by dropping
			// IPC notifications that are longer than 4096 byte.
			if len(notificationBytes) >= 4096 {
				log.Printf("IPC notification dropped: longer than 4095 byte (%d byte).\n", len(notificationBytes))
				continue
			}

			notification := Notification{}
			err := json.Unmarshal(notificationBytes, &notification)
			if err != nil {
				log.Printf("IPC notification dropped: could not unmarshal as JSON '%s': %s.\n", notificationText, err)
				continue
			}

			reader.notifications <- notification
		} else {
			err := scanner.Err()

			// handle EOF
			if err == nil {
				break
			}

			// handle file is closed after Stop() was called
			if errors.Is(err, os.ErrClosed) && reader.closing {
				break
			}

			log.Printf("Could not read from named pipe %s", err)
			break
		}
	}
}

// Notifications returns the notification channel for the Reader.
func (reader *Reader) Notifications() chan Notification {
	return reader.notifications
}

// Close closes the reader.
func (reader *Reader) Close() error {
	reader.closing = true
	err := reader.namedPipe.Close()
	if err != nil {
		return err
	}
	return nil
}
